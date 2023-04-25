package global

import (
	"github.com/luxun9527/gpush/internal/socket/config"
	"go.uber.org/zap"
)

var (
	//普通日志
	L      *zap.Logger
	Config config.Config
)
