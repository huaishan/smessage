package message

import "fmt"

var (
	// ErrInvalidMessage should return When Message is invalid
	ErrInvalidMessage = fmt.Errorf("invalid message")
	// ErrInvalidMessageNoContent should return When message content was not in msg
	ErrInvalidMessageNoContent = fmt.Errorf("invalid message no message content")
	// ErrInvalidMessageContent should return When message content was not in msg
	ErrInvalidMessageContent = fmt.Errorf("invalid message content")
)
