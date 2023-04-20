package initialize

import "github.com/mofei1/gpush/internal/socket/global"

func InitLogger() {
	global.L = global.Config.Logger.Build()
}
