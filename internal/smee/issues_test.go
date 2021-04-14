package smee

import (
	"testing"

	"github.com/matryer/is"
)

func TestIssueEventTypeSystemMessage(t *testing.T) {
	is := is.New(t)

	msg := "this is my message"
	event := IssueEvent{
		Type: TypeSystemMessage,
		Data: NewSystemMessage(msg),
	}

	data, err := event.ParseData()
	is.NoErr(err)

	v, ok := data.(SystemMessage)
	is.True(ok)
	is.Equal(v.Message, msg)
}
