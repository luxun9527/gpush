package initialize

import "github.com/mofei1/gpush/internal/proxy/global"

func InitLogger() {
	global.L = global.Config.Logger.Build()
}
