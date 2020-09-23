package message

import (
	"testing"
)

func TestNewCommonMessageShouldReturnsAMessage(t *testing.T) {
	NewCommonMessage("body", "test", "msg")
}

func TestCommonMessageGetContentShouldReturnsContent(t *testing.T) {
	m := NewCommonMessage("body", "test", "content")

	if m.GetSender() != "body" {
		t.Error("should returns 'body', got: ", m.GetSender())
	}

	if m.GetChannel() != "test" {
		t.Error("should returns 'test', got: ", m.GetChannel())
	}

	if m.GetContent() != "content" {
		t.Error("should returns 'content', got: ", m.GetContent())
	}

}

func TestCommonMessageDumpShouldReturnsDumpedString(t *testing.T) {
	m := NewCommonMessage("body", "test", "content")

	if m.Dump() != `{"sender":"body","channel":"test","msg":"content"}` {
		t.Error(`should returns '{"sender":"body","channel":"test","msg":"content"}', got: `, m.Dump())
	}
}

func TestLoadShouldReturnErrInvalidMessageWhenUnmarshalError(t *testing.T) {
	_, err := Load("body", "test", "")

	if err != ErrInvalidMessage {
		t.Error("err should be ErrInvalidMessage, got: ", err)
	}
}

func TestLoadShouldReturnErrInvalidMessageNoContentWhenNoMessageContent(t *testing.T) {
	_, err := Load("body", "test",`{}`)

	if err != ErrInvalidMessageNoContent {
		t.Error("err should be ErrInvalidMessageNoContent, got: ", err)
	}
}

func TestLoadShouldReturnErrInvalidMessageContentWhenMessageContentInvalid(t *testing.T) {
	_, err := Load("body", "test",`{"msg":1}`)

	if err != ErrInvalidMessageContent {
		t.Error("err should be ErrInvalidMessageContent, got: ", err)
	}
}

func TestLoadShouldReturnNoErrorWhenSucceed(t *testing.T) {
	_, err := Load("body", "test",`{"msg":""}`)

	if err != nil {
		t.Error("err should be nil, got: ", err)
	}
}
