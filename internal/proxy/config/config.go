package config

import (
	"github.com/luxun9527/gpush/internal/pkg"
	"github.com/luxun9527/zaplog"
)

type Config struct {
	Server     Server
	Logger     zaplog.Config
	EtcdConfig pkg.EtcdConfig `mapstructure:"etcd"`
}
