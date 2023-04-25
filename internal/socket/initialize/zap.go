package initialize

import "github.com/luxun9527/gpush/internal/socket/global"

func InitLogger() {
	global.L = global.Config.Logger.Build()
}
