package service

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/huaishan/smessage/message"
	"github.com/rs/xid"
)

const Channel = "c2VydmljZV9jaGFubmVsOjE3OTQ3.43b73abb4eebda36216cb58d79a9e217adc37149a6a8ee069eadfedd837c59c2"

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	send    chan message.Message
	id      string
	channel string
	addr    string
}

func NewClient(addr, channel string, hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		id:      xid.New().String(),
		addr:    addr,
		hub:     hub,
		conn:    conn,
		send:    make(chan message.Message, 256),
		channel: channel,
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			} else {
				log.Println(err)
			}
			break
		}
		msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))
		var _msg message.Message
		if c.channel == Channel {
			_msg, err = message.ServiceLoad(string(msg))
		} else {
			_msg, err = message.Load(c.id, c.channel, string(msg))
		}
		if err != nil {
			log.Println(err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		c.hub.broadcast <- _msg

		if c.channel != Channel {
			c.hub.sBroadcast <- message.NewServiceMessage(c.id, Channel, _msg)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Println(err)
				return
			}
			w.Write([]byte(msg.Dump()))

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				_msg, ok := <-c.send
				if ok != true {
					return
				}
				w.Write([]byte(_msg.Dump()))
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, channel string, addr string, isReg bool) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		id:      xid.New().String(),
		addr:    addr,
		hub:     hub,
		conn:    conn,
		send:    make(chan message.Message, 256),
		channel: channel,
	}

	if isReg {
		client.hub.register <- client
	}
	go client.writePump()
	go client.readPump()
}
