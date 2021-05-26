package issuecache

import (
	"context"

	"github.com/byuoitav/smee/internal/smee"
)

func hasActiveAlerts(issue smee.Issue) bool {
	for _, alert := range issue.Alerts {
		if alert.Active() {
			return true
		}
	}

	return false
}

// activeRoomIssue assumes issuesMu is already read locked
// change to return error/smee.ErrRoomIssueNotFound
func (c *Cache) activeRoomIssue(roomID string) (smee.Issue, bool) {
	for _, issue := range c.issues {
		if issue.Active() && issue.Room.ID == roomID {
			return issue, true
		}
	}

	return smee.Issue{}, false
}

func (c *Cache) ActiveAlertExists(ctx context.Context, roomID, deviceID, typ string) (bool, error) {
	c.issuesMu.RLock()
	defer c.issuesMu.RUnlock()

	issue, ok := c.activeRoomIssue(roomID)
	if !ok {
		return false, nil
	}

	for _, alert := range issue.Alerts {
		if alert.Active() && alert.Device.ID == deviceID && alert.Type == typ {
			return true, nil
		}
	}

	return false, nil
}

func (c *Cache) ActiveAlerts(ctx context.Context) ([]smee.Alert, error) {
	c.issuesMu.RLock()
	defer c.issuesMu.RUnlock()

	var res []smee.Alert
	for _, issue := range c.issues {
		if !issue.Active() {
			continue
		}

		for _, alert := range issue.Alerts {
			if alert.Active() {
				res = append(res, alert)
			}
		}
	}

	return res, nil
}

func (c *Cache) ActiveAlertsByType(ctx context.Context, typ string) ([]smee.Alert, error) {
	c.issuesMu.RLock()
	defer c.issuesMu.RUnlock()

	var res []smee.Alert
	for _, issue := range c.issues {
		if !issue.Active() {
			continue
		}

		for _, alert := range issue.Alerts {
			if alert.Active() && alert.Type == typ {
				res = append(res, alert)
			}
		}
	}

	return res, nil
}

func (c *Cache) ActiveIssue(ctx context.Context, roomID string) (smee.Issue, error) {
	c.issuesMu.RLock()
	defer c.issuesMu.RUnlock()

	issue, ok := c.activeRoomIssue(roomID)
	if !ok {
		return smee.Issue{}, smee.ErrRoomIssueNotFound
	}

	return issue, nil
}

func (c *Cache) ActiveIssues(ctx context.Context) ([]smee.Issue, error) {
	c.issuesMu.RLock()
	defer c.issuesMu.RUnlock()

	var res []smee.Issue
	for _, issue := range c.issues {
		if issue.Active() {
			res = append(res, issue)
		}
	}

	return res, nil
}
