package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

type incidentMapping struct {
	IssueID        int
	SNSysID        string
	SNTicketNumber string
}

func (c *Client) incidentMappings(ctx context.Context, tx pgx.Tx, issueID int) ([]incidentMapping, error) {
	var incs []incidentMapping
	var inc incidentMapping

	_, err := tx.QueryFunc(ctx,
		"SELECT * FROM sn_incident_mappings WHERE issue_id = $1",
		[]interface{}{issueID},
		[]interface{}{&inc.IssueID, &inc.SNSysID, &inc.SNTicketNumber},
		func(pgx.QueryFuncRow) error {
			incs = append(incs, incidentMapping{
				IssueID:        inc.IssueID,
				SNSysID:        inc.SNSysID,
				SNTicketNumber: inc.SNTicketNumber,
			})

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to queryFunc: %w", err)
	}

	return incs, nil
}

func (c *Client) createIncidentMapping(ctx context.Context, tx pgx.Tx, mapping incidentMapping) error {
	_, err := tx.Exec(ctx,
		"INSERT INTO sn_incident_mappings (issue_id, sn_sys_id, sn_ticket_number) VALUES ($1, $2, $3)",
		mapping.IssueID, mapping.SNSysID, mapping.SNTicketNumber)
	if err != nil {
		return fmt.Errorf("unable to exec: %w", err)
	}

	return nil
}
