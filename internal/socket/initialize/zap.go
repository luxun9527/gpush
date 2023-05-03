package initialize

import "github.com/luxun9527/gpush/internal/socket/global"

// InitLogger 初始化日志
func InitLogger() {
	global.L = global.Config.Logger.Build()
}
