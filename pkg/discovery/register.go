package discovery

import (
	"context"
	"encoding/json"
	"log"

	v3 "go.etcd.io/etcd/clientv3"
)

type Payload struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

type Service struct {
	Name string
	Payload
	leaseID v3.LeaseID
	client  *v3.Client
	Key     string
}

func NewService(cli *v3.Client, name string, payload Payload) (*Service, error) {
	return &Service{
		Name:    name,
		Payload: payload,
		client:  cli,
	}, nil
}

func (s *Service) Register(ctx context.Context) error {
	val, _ := json.Marshal(&s.Payload)

	resp, err := s.client.Grant(ctx, defaultTTL)
	if err != nil {
		log.Printf("Call client.Grant error: %s\n", err.Error())
		return err
	}
	s.leaseID = resp.ID

	_, err = s.client.Put(ctx, GenKey(s.Name, s.leaseID), string(val), v3.WithLease(s.leaseID))
	if err != nil {
		log.Printf("Call client.Put error: %s\n", err.Error())
		return err
	}
	s.Key = GenKey(s.Name, s.leaseID)

	err = s.keepAlive(ctx)
	if err != nil {
		log.Printf("Call keepAlive error: %s\n", err.Error())
		return err
	}

	return nil
}

func (s *Service) Stop() {
	s.revoke(context.Background())
}

func (s *Service) keepAlive(ctx context.Context) error {
	keepCh, err := s.client.KeepAlive(ctx, s.leaseID)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Fatal("server closed")
			case <-s.client.Ctx().Done():
				log.Fatal("server closed")
			case _, ok := <-keepCh:
				if !ok {
					log.Println("keepalive channel closed")
					s.revoke(ctx)
				}
				//log.Printf("service: %s, lease_id: %d, ttl: %d", s.Name, s.leaseID, resp.TTL)
			}
		}
	}()

	return nil
}

func (s *Service) revoke(ctx context.Context) error {
	s.client.Revoke(ctx, s.leaseID)

	log.Fatalf("revoke service: %s\n", GenKey(s.Name, s.leaseID))
	return nil
}
