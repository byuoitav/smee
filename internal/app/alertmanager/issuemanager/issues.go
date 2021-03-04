package issues

import (
	"context"

	"github.com/byuoitav/smee/internal/smee"
)

type Manager struct {
	IssueStore smee.IssueStore
}

func (m *Manager) Run(ctx context.Context, alert smee.Alert) error {
}
