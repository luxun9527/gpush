package global

import (
	"github.com/ikun2021/gpush/internal/socket/config"
	"github.com/ikun2021/gpush/internal/socket/stat"
)

var (
	Config     config.Config
	Prometheus *stat.PrometheusStat
)
