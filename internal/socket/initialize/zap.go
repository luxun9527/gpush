package initialize

import "ws/internal/socket/global"

func InitLogger() {
	global.L = global.Config.Logger.Build()
}
