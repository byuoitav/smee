package smee

import "errors"

var (
	ErrRoomIssueNotFound = errors.New("no active issue found for the given room")
)
