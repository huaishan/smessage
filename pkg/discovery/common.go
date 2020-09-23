package discovery

import (
	"fmt"

	v3 "go.etcd.io/etcd/clientv3"
)

const (
	keyPfx     = "services"
	defaultTTL = 1 // s
)

func GenPath(name string) string {
	return fmt.Sprintf("%s/%s", keyPfx, name)
}

func GenKey(name string, id v3.LeaseID) string {
	return fmt.Sprintf("%s/%d", GenPath(name), id)
}
