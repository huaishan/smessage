package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "127.0.0.1:8001", "http service address")
var channel = flag.String("channel", "abcc", "channel")
var frequence = flag.Int64("frequence", 1000, "frequence")

func _main() {
	flag.Parse()
	log.SetFlags(0)

	if *frequence <= 0 {
		fmt.Println("frequence should  more than 0")
	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	header := make(http.Header)
	header.Set("Channel", *channel)

	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		fmt.Println("dial:", err)
		return
	}
	defer c.Close()

	done := make(chan struct{})

	ticker := time.NewTicker(time.Duration(*frequence) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			// err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			data, _ := json.Marshal(map[string]string{
				"msg": t.String(),
			})
			err := c.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				return
			} else {
				fmt.Println("send: ", string(data))
			}
		case <-interrupt:
			log.Println("interrupt")
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			os.Exit(0)
		}
	}
}

func main() {
	for {
		_main()
		time.Sleep(time.Second)
	}
}
