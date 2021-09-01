package smee

import (
	"context"
)

type IssueTypeStore interface {
	IssueType(ctx context.Context) (map[string]IssueType, error)
}

type IssueType struct {
	IdAlertType string `json:"idAlertType"`
	KbArticle   string `json:"kbArticle"`
}
