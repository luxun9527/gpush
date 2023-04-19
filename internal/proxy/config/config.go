package config

import "github.com/mofei1/gpush/internal/pkg"

type Config struct {
	Server Server
	Logger pkg.LoggerConfig
}