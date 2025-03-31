package global

import (
	"github.com/luxun9527/gpush/internal/socket/config"
	"github.com/luxun9527/gpush/internal/socket/stat"
)

var (
	Config     config.Config
	Prometheus *stat.PrometheusStat
)
