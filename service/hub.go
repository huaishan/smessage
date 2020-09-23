package service

import (
	"log"

	"github.com/huaishan/smessage/message"
)

type Hub struct {
	clients    map[string]map[*Client]bool
	broadcast  chan message.Message
	sBroadcast chan message.Message // 服务间发消息
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan message.Message),
		sBroadcast: make(chan message.Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if h.clients[client.channel] == nil {
				h.clients[client.channel] = make(map[*Client]bool)
			}
			h.clients[client.channel][client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client.channel]; ok {
				delete(h.clients[client.channel], client)
				close(client.send)
			}
		case msg := <-h.broadcast:
			if _, ok := h.clients[msg.GetChannel()]; !ok {
				continue
			}
			for client := range h.clients[msg.GetChannel()] {
				if client.id == msg.GetSender() {
					log.Println("skip sender", client, msg)
					continue
				}
				select {
				case client.send <- msg:
				default:
					close(client.send)
					delete(h.clients[client.channel], client)
				}
			}
		case msg := <-h.sBroadcast:
			for client := range h.clients[Channel] {
				if client.id == msg.GetSender() {
					log.Println("skip sender2", client)
					continue
				}
				select {
				case client.send <- msg:
				default:
					close(client.send)
					delete(h.clients[client.channel], client)
				}
			}
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
	go client.readPump()
	go client.writePump()
}

func (h *Hub) UnRegister(client *Client) {
	h.unregister <- client
}
