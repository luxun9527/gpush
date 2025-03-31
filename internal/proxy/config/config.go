package config

import (
	"github.com/luxun9527/gpush/internal/pkg"
	"github.com/luxun9527/zlog"
)

type Config struct {
	Server     Server
	Logger     zlog.Config
	EtcdConfig pkg.EtcdConfig `mapstructure:"etcd"`
}
