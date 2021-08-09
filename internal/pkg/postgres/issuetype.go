package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/byuoitav/smee/internal/smee"
	"github.com/jackc/pgx/v4"
)

type issuetype struct {
	AlertTypeId sql.NullString
	KbArticle   sql.NullString
}

func (c *Client) IssueType(ctx context.Context) (map[string]smee.IssueType, error) {
	service := make(map[string]smee.IssueType)
	var issT issuetype

	_, err := c.pool.QueryFunc(ctx,
		"SELECT * FROM alert_types",
		[]interface{}{},
		[]interface{}{&issT.AlertTypeId, &issT.KbArticle},
		func(pgx.QueryFuncRow) error {
			smeeSN := convertIssueType(issT)
			service[smeeSN.IdAlertType] = smeeSN
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to queryFunc: %w", err)
	}
	return service, nil
}

func convertIssueType(sn issuetype) smee.IssueType {

	smeeIssueType := smee.IssueType{}

	if sn.AlertTypeId.Valid && sn.KbArticle.Valid {
		smeeIssueType.IdAlertType = sn.AlertTypeId.String
		smeeIssueType.KbArticle = sn.KbArticle.String
	} else {
		smeeIssueType.IdAlertType = ""
		smeeIssueType.KbArticle = ""
	}
	return smeeIssueType

}
