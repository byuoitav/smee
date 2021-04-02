package issuecache

import (
	"context"
	"fmt"
	"sync"

	"github.com/byuoitav/smee/internal/smee"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

type cache struct {
	persistent smee.IssueStore
	log        *zap.Logger

	cache map[string]smee.Issue
	sync.RWMutex
}

func New(ctx context.Context, persistent smee.IssueStore, log *zap.Logger) (*cache, error) {
	c := &cache{
		persistent: persistent,
		cache:      make(map[string]smee.Issue),
		log:        log,
	}

	if persistent != nil {
		issues, err := persistent.ActiveIssues(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to get persistent active issues: %w", err)
		}

		for i := range issues {
			c.cache[issues[i].ID] = issues[i]
		}
	}

	return c, nil
}

func (c *cache) CreateIssue(ctx context.Context, issue smee.Issue) (smee.Issue, error) {
	c.log.Info("Creating issue", zap.String("room", issue.Room))
	c.Lock()
	defer c.Unlock()

	switch {
	case issue.ID != "":
	case c.persistent != nil:
		var err error
		issue, err = c.persistent.CreateIssue(ctx, issue)
		if err != nil {
			return issue, fmt.Errorf("unable to create persistent issue: %w", err)
		}
	default:
		issue.ID = ksuid.New().String()
	}

	c.cache[issue.ID] = issue
	return issue, nil
}

func (c *cache) CloseIssue(ctx context.Context, id string) error {
	c.Lock()
	defer c.Unlock()

	c.log.Info("Closing issue", zap.String("room", c.cache[id].Room))

	if c.persistent != nil {
		if err := c.persistent.CloseIssue(ctx, id); err != nil {
			return fmt.Errorf("unable to close persistent issue: %w", err)
		}
	}

	delete(c.cache, id)
	return nil
}

func (c *cache) ActiveIssues(ctx context.Context) ([]smee.Issue, error) {
	var res []smee.Issue

	c.RLock()
	defer c.RUnlock()

	for _, issue := range c.cache {
		res = append(res, issue)
	}

	return res, nil
}

func (c *cache) ActiveIssueForRoom(ctx context.Context, room string) (smee.Issue, bool, error) {
	c.RLock()
	defer c.RUnlock()

	for _, issue := range c.cache {
		if issue.Room == room {
			return issue, true, nil
		}
	}

	return smee.Issue{}, false, nil
}

func (c *cache) AddIssueEvent(ctx context.Context, id string, event smee.IssueEvent) error {
	c.Lock()
	defer c.Unlock()

	issue, ok := c.cache[id]
	if !ok {
		return fmt.Errorf("invalid issue id")
	}

	if c.persistent != nil {
		if err := c.persistent.AddIssueEvent(ctx, id, event); err != nil {
			return fmt.Errorf("unable to add persistent issue event: %w", err)
		}
	}

	issue.Events = append(issue.Events, event)
	c.cache[id] = issue
	return nil
}
