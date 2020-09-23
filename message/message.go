package message

import (
	"encoding/json"
)

// Load unmarsharls a message that was in the storage
func Load(sender, channel, msgt string) (Message, error) {
	m := make(map[string]interface{})

	err := json.Unmarshal([]byte(msgt), &m)
	if err != nil {
		return nil, ErrInvalidMessage
	}

	msg, ok := m["msg"]
	if ok == false {
		return nil, ErrInvalidMessageNoContent
	}

	_msg, ok := msg.(string)
	if ok == false {
		return nil, ErrInvalidMessageContent
	}

	return NewCommonMessage(sender, channel, _msg), nil
}

func ServiceLoad(msg string) (Message, error) {
	m := &serviceMessageUnmarshal{}
	err := json.Unmarshal([]byte(msg), &m)
	if err != nil {
		return nil, ErrInvalidMessage
	}

	return m.Msg, nil
}

// Message provides a dozen of method to operate message
type Message interface {
	// GetSender get sender
	GetSender() string
	// GetChannel get channel
	GetChannel() string
	// Get message body
	GetContent() string
	// Dump message
	Dump() string
}

type commonMessage struct {
	Sender  string `json:"sender"`
	Channel string `json:"channel"`
	Msg     string `json:"msg"`
}

// NewCommonMessage returns a Message
func NewCommonMessage(sender, channel, msg string) Message {
	return &commonMessage{
		Sender:  sender,
		Channel: channel,
		Msg:     msg,
	}
}

func (cm *commonMessage) GetSender() string {
	return cm.Sender
}

func (cm *commonMessage) GetChannel() string {
	return cm.Channel
}

func (cm *commonMessage) GetContent() string {
	return cm.Msg
}

func (cm *commonMessage) Dump() string {
	res, _ := json.Marshal(cm)

	return string(res)
}

type serviceMessage struct {
	Sender  string  `json:"sender"`
	Channel string  `json:"channel"`
	Msg     Message `json:"msg"`
}

type serviceMessageUnmarshal struct {
	Sender  string         `json:"sender"`
	Channel string         `json:"channel"`
	Msg     *commonMessage `json:"msg"`
}

func NewServiceMessage(sender, channel string, msg Message) Message {
	return &serviceMessage{
		Sender:  sender,
		Channel: channel,
		Msg:     msg,
	}
}

func (cm *serviceMessage) GetSender() string {
	return cm.Sender
}

func (cm *serviceMessage) GetChannel() string {
	return cm.Channel
}

func (cm *serviceMessage) GetContent() string {
	return cm.Msg.Dump()
}

func (cm *serviceMessage) Dump() string {
	res, _ := json.Marshal(cm)

	return string(res)
}
