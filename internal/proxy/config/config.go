package config

import "github.com/luxun9527/gpush/internal/pkg"

type Config struct {
	Server Server
	Logger pkg.LoggerConfig
}
