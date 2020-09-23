package discovery

import (
	"context"
	"encoding/json"
	"log"

	v3 "go.etcd.io/etcd/clientv3"
)

type Discovery struct {
	Name     string
	Nodes    map[string]*Node
	Client   *v3.Client
	SelfKey  string
	NodeChan chan *Node
}

type Node struct {
	State bool
	Key   string
	Payload
}

func NewDiscovery(ctx context.Context, cli *v3.Client, watchName string) (*Discovery, error) {
	dc := &Discovery{
		Name:     watchName,
		Nodes:    make(map[string]*Node),
		Client:   cli,
		NodeChan: make(chan *Node, 5),
	}

	resp, err := cli.Get(ctx, GenPath(watchName), v3.WithPrefix())
	if err != nil {
		return nil, err
	}

	for _, kv := range resp.Kvs {
		log.Printf("Add %s : %q\n", kv.Key, kv.Value)
		dc.AddNode(string(kv.Key), GetPayload(kv.Value))
	}

	go dc.WatchNodes(ctx)

	return dc, nil
}

func (d *Discovery) AddNode(key string, payload *Payload) {
	d.Nodes[key] = &Node{
		State:   true,
		Key:     key,
		Payload: *payload,
	}
	d.NodeChan <- d.Nodes[key]
}

func GetPayload(value []byte) *Payload {
	pl := &Payload{}
	err := json.Unmarshal(value, pl)
	if err != nil {
		log.Printf("Call GetPayload error: %s\n", err.Error())
	}
	return pl
}

func (d *Discovery) WatchNodes(ctx context.Context) {
	wch := d.Client.Watch(ctx, GenPath(d.Name), v3.WithPrefix())

	for resp := range wch {
		for _, ev := range resp.Events {
			switch ev.Type {
			case v3.EventTypePut:
				log.Printf("Add %s : %q\n", ev.Kv.Key, ev.Kv.Value)
				pl := GetPayload(ev.Kv.Value)
				d.AddNode(string(ev.Kv.Key), pl)
			case v3.EventTypeDelete:
				log.Printf("Delete %s : %q\n", ev.Kv.Key, ev.Kv.Value)
				delete(d.Nodes, string(ev.Kv.Key))
			}
		}
	}
}
