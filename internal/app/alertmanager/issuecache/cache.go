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

// TODO some timer to sync the cache with the issue store, or a way to force it (handler?)
// probably for all of the caches, not just this one (issuecache, maintenancecache, ...maybe that's all of them)

type Cache struct {
	IssueStore    smee.IssueStore
	IncidentStore smee.IncidentStore
	Log           *zap.Logger

	// issues is a map of issueID to the currently active issue for that room
	issues map[string]smee.Issue
	// issuesMu protects issues
	issuesMu sync.RWMutex
}

func (c *Cache) Sync(ctx context.Context) error {
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

	c.Log.Info("Synced cache", zap.Int("issueCount", len(c.issues)))
	return nil
}

func (c *Cache) CreateAlert(ctx context.Context, alert smee.Alert) (smee.Issue, error) {
	c.issuesMu.Lock()
	defer c.issuesMu.Unlock()

	if c.IssueStore != nil {
		iss, err := c.IssueStore.CreateAlert(ctx, alert)
		if err != nil {
			return smee.Issue{}, fmt.Errorf("unable to create alert on substore: %w", err)
		}

		// update the cache
		c.issues[iss.ID] = iss
		return iss, nil
	}

	alert.ID = ksuid.New().String()

	issue, ok := c.activeRoomIssue(alert.Device.Room.ID)
	if !ok {
		// create an issue if needed
		issue = smee.Issue{
			ID:        ksuid.New().String(),
			Room:      alert.Device.Room,
			Start:     alert.Start,
			Alerts:    make(map[string]smee.Alert),
			Incidents: make(map[string]smee.Incident),
		}

		c.Log.Info("Creating issue", zap.String("roomID", issue.Room.ID), zap.String("issueID", issue.ID))
	}

	c.Log.Info("Creating alert", zap.String("roomID", alert.Device.Room.ID), zap.String("issueID", issue.ID), zap.String("alertID", alert.ID), zap.String("deviceID", alert.Device.ID), zap.String("type", alert.Type))

	alert.IssueID = issue.ID
	issue.Alerts[alert.ID] = alert
	c.issues[issue.ID] = issue
	return issue, nil
}

func (c *Cache) CloseAlert(ctx context.Context, issueID, alertID string) (smee.Issue, error) {
	c.issuesMu.Lock()
	defer c.issuesMu.Unlock()

	if c.IssueStore != nil {
		iss, err := c.IssueStore.CloseAlert(ctx, issueID, alertID)
		if err != nil {
			return smee.Issue{}, fmt.Errorf("unable to close alert on substore: %w", err)
		}

		// update the cache
		if iss.Active() {
			c.issues[iss.ID] = iss
		} else {
			delete(c.issues, iss.ID)
		}

		return iss, nil
	}

	issue, ok := c.issues[issueID]
	if !ok {
		return smee.Issue{}, errors.New("issue does not exist")
	}

	alert, ok := issue.Alerts[alertID]
	if !ok {
		return smee.Issue{}, errors.New("alert does not exist on issue")
	}

	c.Log.Info("Closing alert", zap.String("roomID", alert.Device.Room.ID), zap.String("issueID", issue.ID), zap.String("alertID", alert.ID), zap.String("deviceID", alert.Device.ID), zap.String("type", alert.Type))

	alert.End = time.Now()
	issue.Alerts[alert.ID] = alert

	// close the issue if there are no more active alerts
	if hasActiveAlerts(issue) {
		c.issues[issue.ID] = issue
	} else {
		c.Log.Info("Closing issue", zap.String("roomID", issue.Room.ID), zap.String("issueID", issueID))

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
			return fmt.Errorf("unable to add issue event on substore: %w", err)
		}
	}

	issue, ok := c.issues[issueID]
	if !ok {
		// TODO
		// something is weird if the above didn't return an error (and c.IssueStore != nil)..
		// like the cache is wrong?
		return nil
	}

	issue.Events = append(issue.Events, events...)
	c.issues[issue.ID] = issue

	if c.IncidentStore != nil {
		for incID := range issue.Incidents {
			if err := c.IncidentStore.AddIssueEvents(ctx, incID, events...); err != nil {
				return fmt.Errorf("unable to add issue events to incident %q", incID)
			}
		}
	}

	return nil
}

func (c *Cache) LinkIncident(ctx context.Context, issueID string, inc smee.Incident) (smee.Issue, error) {
	c.issuesMu.Lock()
	defer c.issuesMu.Unlock()

	if c.IssueStore != nil {
		iss, err := c.IssueStore.LinkIncident(ctx, issueID, inc)
		if err != nil {
			return smee.Issue{}, fmt.Errorf("unable to link incident on substore: %w", err)
		}

		// update the cache
		c.issues[iss.ID] = iss
		return iss, nil
	}

	issue, ok := c.issues[issueID]
	if !ok {
		// for the cache, we're just going to assume this issue has been closed
		// and that linking at this point doesn't do anything anyways
		return smee.Issue{}, nil
	}

	issue.Incidents[inc.ID] = inc
	c.issues[issue.ID] = issue
	return issue, nil
}
