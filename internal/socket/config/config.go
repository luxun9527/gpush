package config

import "github.com/luxun9527/gpush/internal/pkg"

type Config struct {
	Bucket     Bucket           `mapstructure:"bucket"`
	Server     Server           `mapstructure:"server"`
	Logger     pkg.LoggerConfig `mapstructure:"logger"`
	Proxy      Proxy            `mapstructure:"proxy"`
	Connection Connection       `mapstructure:"connection"`
	EtcdConfig pkg.EtcdConfig   `mapstructure:"etcd"`
}
