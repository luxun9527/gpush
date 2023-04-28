package global

import (
	"github.com/luxun9527/gpush/internal/socket/config"
	"go.uber.org/zap"
)

var (
	L      *zap.Logger
	Config config.Config
)
