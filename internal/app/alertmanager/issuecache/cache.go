package issuecache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/byuoitav/smee/internal/smee"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

type Cache struct {
	IssueStore smee.IssueStore
	Log        *zap.Logger

	// issues is a map of issueID to the currently active issue for that room
	issues map[string]smee.Issue
	// issuesMu protects issues
	issuesMu sync.RWMutex
}

func (c *Cache) Populate(ctx context.Context) error {
	c.issuesMu.Lock()
	defer c.issuesMu.Unlock()

	c.issues = make(map[string]smee.Issue)

	if c.IssueStore != nil {
		issues, err := c.IssueStore.ActiveIssues(ctx)
		if err != nil {
			return fmt.Errorf("unable to get active issues: %w", err)
		}

		for i := range issues {
			c.issues[issues[i].ID] = issues[i]
		}
	}

	c.Log.Info("Populated cache", zap.Int("issueCount", len(c.issues)))
	return nil
}

func (c *Cache) CreateAlert(ctx context.Context, alert smee.Alert) (smee.Issue, error) {
	c.issuesMu.Lock()
	defer c.issuesMu.Unlock()

	if c.IssueStore != nil {
		issue, err := c.IssueStore.CreateAlert(ctx, alert)
		if err != nil {
			return smee.Issue{}, fmt.Errorf("unable to create alert on substore: %w", err)
		}

		c.issues[issue.Room] = issue
		return issue, nil
	}

	alert.ID = ksuid.New().String()
	c.Log.Info("Creating alert", zap.String("room", alert.Room), zap.String("device", alert.Device), zap.String("type", alert.Type))

	issue, ok := c.activeRoomIssue(alert.Room)
	if !ok {
		// create an issue if needed
		issue = smee.Issue{
			ID:     ksuid.New().String(),
			Room:   alert.Room,
			Start:  alert.Start,
			Alerts: make(map[string]smee.Alert),
		}
	}

	alert.IssueID = issue.ID
	issue.Alerts[alert.ID] = alert
	c.issues[issue.ID] = issue
	return issue, nil
}

func (c *Cache) CloseAlert(ctx context.Context, issueID, alertID string) (smee.Issue, error) {
	c.issuesMu.Lock()
	defer c.issuesMu.Unlock()

	if c.IssueStore != nil {
		issue, err := c.IssueStore.CloseAlert(ctx, issueID, alertID)
		if err != nil {
			return smee.Issue{}, fmt.Errorf("unable to close alert on substore: %w", err)
		}

		if issue.Active() {
			c.issues[issue.ID] = issue
		} else {
			delete(c.issues, issue.ID)
		}

		return issue, nil
	}

	c.Log.Info("Closing alert", zap.String("alertID", alertID))

	issue, ok := c.issues[issueID]
	if !ok {
		return smee.Issue{}, errors.New("issue does not exist")
	}

	alert, ok := issue.Alerts[alertID]
	if !ok {
		return smee.Issue{}, errors.New("alert does not exist on issue")
	}

	alert.End = time.Now()
	issue.Alerts[alert.ID] = alert

	// close the issue if there are no more active alerts
	if hasActiveAlerts(issue) {
		c.issues[issue.ID] = issue
	} else {
		issue.End = time.Now()
		delete(c.issues, issue.ID)
	}

	return issue, nil
}

func (c *Cache) AddIssueEvents(ctx context.Context, issueID string, events ...smee.IssueEvent) error {
	c.issuesMu.Lock()
	defer c.issuesMu.Unlock()

	if c.IssueStore != nil {
		if err := c.IssueStore.AddIssueEvents(ctx, issueID, events...); err != nil {
			return fmt.Errorf("unable to add issue event  on substore: %w", err)
		}

		return nil
	}

	issue, ok := c.issues[issueID]
	if !ok {
		return errors.New("issue does not exist")
	}

	issue.Events = append(issue.Events, events...)
	c.issues[issue.ID] = issue
	return nil
}
