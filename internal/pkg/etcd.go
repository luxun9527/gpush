package pkg

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type EtcdConfig struct {
	// Endpoints is a list of URLs.
	Endpoints []string `mapstructure:"endpoints"`

	// Username is a user name for authentication.
	Username string `json:"username"`
	// Password is a password for authentication.
	Password  string `json:"password"`
	KeyPrefix string `mapstructure:"keyPrefix"`
}

func (etcdConfig EtcdConfig) BuildClient() (*clientv3.Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdConfig.Endpoints,
		DialTimeout: 5 * time.Second,
	})
	return cli, err
}
