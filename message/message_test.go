package message

import (
	"testing"
)

func TestNewCommonMessageShouldReturnsAMessage(t *testing.T) {
	NewCommonMessage("body")
}

func TestCommonMessageGetContentShouldReturnsContent(t *testing.T) {
	m := NewCommonMessage("content")

	if m.GetContent() != "content" {
		t.Error("should returns content, got: ", m.GetContent())
	}
}

func TestCommonMessageDumpShouldReturnsDumpedString(t *testing.T) {
	m := NewCommonMessage("content")

	if m.Dump() != `{"msg":"content"}` {
		t.Error(`should returns '{"msg":"content"}', got: `, m.Dump())
	}
}

func TestLoadShouldReturnErrInvalidMessageWhenUnmarshalError(t *testing.T) {
	_, err := Load("")

	if err != ErrInvalidMessage {
		t.Error("err should be ErrInvalidMessage, got: ", err)
	}
}

func TestLoadShouldReturnErrInvalidMessageNoContentWhenNoMessageContent(t *testing.T) {
	_, err := Load(`{}`)

	if err != ErrInvalidMessageNoContent {
		t.Error("err should be ErrInvalidMessageNoContent, got: ", err)
	}
}

func TestLoadShouldReturnErrInvalidMessageContentWhenMessageContentInvalid(t *testing.T) {
	_, err := Load(`{"msg":1}`)

	if err != ErrInvalidMessageContent {
		t.Error("err should be ErrInvalidMessageContent, got: ", err)
	}
}

func TestLoadShouldReturnNoErrorWhenSucceed(t *testing.T) {
	_, err := Load(`{"msg":""}`)

	if err != nil {
		t.Error("err should be nil, got: ", err)
	}
}
