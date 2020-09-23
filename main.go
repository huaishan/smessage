package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	v3 "go.etcd.io/etcd/clientv3"

	"github.com/gorilla/websocket"
	"github.com/huaishan/smessage/pkg/discovery"
	"github.com/huaishan/smessage/service"
	"github.com/huaishan/smessage/utils"
)

func GetPort(addr string) int {
	port, err := strconv.Atoi(strings.Split(addr, ":")[1])
	if err != nil {
		panic(err)
	}

	return port
}

const serviceName = "smessage_service"

func InitCli(endpoints []string) *v3.Client {
	cli, err := v3.New(v3.Config{
		Endpoints:   endpoints,
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	return cli
}

func Register(ctx context.Context, cli *v3.Client, ip string, port int) *discovery.Service {
	ser, err := discovery.NewService(
		cli,
		serviceName,
		discovery.Payload{IP: ip, Port: port},
	)
	if err != nil {
		panic(err)
	}

	err = ser.Register(ctx)
	if err != nil {
		panic(err)
	}

	return ser
}

func InitDiscovery(ctx context.Context, cli *v3.Client) *discovery.Discovery {
	dc, err := discovery.NewDiscovery(ctx, cli, serviceName)
	if err != nil {
		panic(err)
	}
	return dc
}

func InitServiceChan(dc *discovery.Discovery, ser *discovery.Service, hub *service.Hub) {
	for node := range dc.NodeChan {
		if node.Key == ser.Key {
			continue
		}
		u := url.URL{
			Scheme: "ws",
			Host:   fmt.Sprintf("%s:%d", node.IP, node.Port),
			Path:   "/ws",
		}

		log.Printf("connect %s ...", u.String())
		var conn *websocket.Conn
		var resp *http.Response
		var err error
		header := make(http.Header)
		header.Set("Channel", service.Channel)
		for i := 0; i < 10; i++ {
			conn, resp, err = websocket.DefaultDialer.Dial(u.String(), header)
			if err != nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}
		}
		if err != nil {
			log.Println("dial:", err, resp)
			continue
		}
		log.Printf("connect %s success.", u.String())

		client := service.NewClient(*addr, service.Channel, hub, conn)

		hub.Register(client)
	}
}

var addr = flag.String("addr", ":8000", "http service address")
var endpoints = flag.String("endpoints", "127.0.0.1:2379", "etcd endpoints. '127.0.0.1:2379,127.0.0.1:2379,127.0.0.1:2379'")

func main() {
	flag.Parse()

	eps := strings.Split(*endpoints, ",")
	cli := InitCli(eps)

	ip := utils.GetLocalIP()
	port := GetPort(*addr)

	dc := InitDiscovery(context.TODO(), cli)

	hub := service.NewHub()
	go hub.Run()

	srv := &http.Server{
		Addr:         *addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		channel := r.Header.Get("Channel")
		if channel == "" {
			return
		}

		addr := utils.GetRealAddr(r)
		service.ServeWs(hub, w, r, channel, addr, channel != service.Channel)
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err.Error())
		}
	}()
	log.Printf("Listen %s...", *addr)

	ser := Register(context.TODO(), cli, ip, port)
	go InitServiceChan(dc, ser, hub)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server...")
	ser.Stop()

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err.Error())
	}
}
