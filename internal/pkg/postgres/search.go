package postgres

import (
	"context"
	"fmt"
	"strconv"

	"github.com/byuoitav/smee/internal/smee"
)

func (c *Client) ActiveIssue(ctx context.Context, roomID string) (smee.Issue, error) {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to start tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	issID, err := c.activeIssueID(ctx, tx, roomID)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to get active issue ID: %w", err)
	}

	smeeIss, err := c.smeeIssue(ctx, tx, issID)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to get smeeIssue: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return smee.Issue{}, fmt.Errorf("unable to commit tx: %w", err)
	}

	return smeeIss, nil
}

func (c *Client) ActiveIssues(ctx context.Context) ([]smee.Issue, error) {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to start tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	issIds, err := c.activeIssueIDs(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("unable to get active issue ids: %w", err)
	}

	var smeeIssues []smee.Issue
	for _, issID := range issIds {
		smeeIss, err := c.smeeIssue(ctx, tx, issID)
		if err != nil {
			return nil, fmt.Errorf("unable to get smeeIssue (%v): %w", issID, err)
		}

		smeeIssues = append(smeeIssues, smeeIss)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("unable to commit tx: %w", err)
	}

	return smeeIssues, nil
}

func (c *Client) ActiveAlertExists(ctx context.Context, roomID, deviceID, typ string) (bool, error) {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("unable to start tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	exists, err := c.activeAlertExists(ctx, tx, roomID, deviceID, typ)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("unable to commit tx: %w", err)
	}

	return exists, nil
}

func (c *Client) ActiveAlerts(ctx context.Context) ([]smee.Alert, error) {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to start tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	alerts, err := c.activeAlerts(ctx, tx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("unable to commit tx: %w", err)
	}

	var smeeAlerts []smee.Alert
	for _, alert := range alerts {
		smeeAlerts = append(smeeAlerts, convertAlert(alert))
	}

	return smeeAlerts, nil
}

func (c *Client) ActiveAlertsByType(ctx context.Context, typ string) ([]smee.Alert, error) {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to start tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	alerts, err := c.activeAlertsByType(ctx, tx, typ)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("unable to commit tx: %w", err)
	}

	var smeeAlerts []smee.Alert
	for _, alert := range alerts {
		smeeAlerts = append(smeeAlerts, convertAlert(alert))
	}

	return smeeAlerts, nil
}

func convertAlert(a alert) smee.Alert {
	smeeAlert := smee.Alert{
		ID:      strconv.Itoa(a.ID),
		IssueID: strconv.Itoa(a.IssueID),
		Device: smee.Device{
			ID:   a.CouchDeviceID,
			Name: a.CouchDeviceID,
			Room: smee.Room{
				ID:   a.CouchRoomID,
				Name: a.CouchRoomID,
			},
		},
		Type:  a.AlertType,
		Start: a.StartTime,
	}

	if a.EndTime != nil {
		smeeAlert.End = *a.EndTime
	}

	return smeeAlert
}
