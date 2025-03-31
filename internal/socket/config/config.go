package config

import (
	"github.com/luxun9527/gpush/internal/pkg"
	"github.com/luxun9527/zlog"
)

type Config struct {
	Bucket     Bucket         `mapstructure:"bucket"`
	Server     Server         `mapstructure:"server"`
	Logger     zlog.Config    `mapstructure:"logger"`
	Proxy      Proxy          `mapstructure:"proxy"`
	Connection Connection     `mapstructure:"connection"`
	ProxyRpc   pkg.EtcdConfig `mapstructure:"ProxyRpc"`
	AuthUrl    string         `mapstructure:"authUrl"`
}
