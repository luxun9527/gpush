package global

import (
	"github.com/mofei1/gpush/internal/proxy/config"
	"go.uber.org/zap"
)

var (
	Config config.Config
	L      *zap.Logger
)
