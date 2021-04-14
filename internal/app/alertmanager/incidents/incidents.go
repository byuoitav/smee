package incidents

import (
	"context"
	"fmt"

	"github.com/byuoitav/smee/internal/pkg/servicenow"
	"github.com/byuoitav/smee/internal/smee"
)

// Store is a wrapper around package servicenow that
// implements smee.IncidentStore
type Store struct {
	Client *servicenow.Client
}

func convert(from servicenow.Incident) smee.Incident {
	return smee.Incident{
		ID:   from.ID,
		Name: from.Number,
	}
}

func (s *Store) Incident(ctx context.Context, id string) (smee.Incident, error) {
	inc, err := s.Client.Incident(ctx, id)
	if err != nil {
		return smee.Incident{}, err
	}

	return convert(inc), nil
}

func (s *Store) IncidentByName(ctx context.Context, name string) (smee.Incident, error) {
	inc, err := s.Client.IncidentByNumber(ctx, name)
	if err != nil {
		return smee.Incident{}, err
	}

	return convert(inc), nil
}

func (s *Store) AddIssueEvents(ctx context.Context, id string, events ...smee.IssueEvent) error {
	for i, event := range events {
		data, err := event.ParseData()
		if err != nil {
			continue
		}

		switch v := data.(type) {
		case smee.SystemMessage:
			if err := s.Client.AddInternalNote(ctx, id, v.Message); err != nil {
				return fmt.Errorf("unable to add event %d/%d: %w", i+1, len(events), err)
			}
		default:
			// skip it
		}
	}

	return nil
}
