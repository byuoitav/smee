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
func (c *Cache) activeRoomIssue(room string) (smee.Issue, bool) {
	for _, issue := range c.issues {
		if issue.Active() && issue.Room == room {
			return issue, true
		}
	}

	return smee.Issue{}, false
}

func (c *Cache) ActiveAlertExists(ctx context.Context, room, device, typ string) (bool, error) {
	c.issuesMu.RLock()
	defer c.issuesMu.RUnlock()

	issue, ok := c.activeRoomIssue(room)
	if !ok {
		return false, nil
	}

	for _, alert := range issue.Alerts {
		if alert.Active() && alert.Device == device && alert.Type == typ {
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
