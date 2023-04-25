package global

import (
	"github.com/luxun9527/gpush/internal/proxy/config"
	"go.uber.org/zap"
)

var (
	Config config.Config
	L      *zap.Logger
)
