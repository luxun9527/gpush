package initialize

import "github.com/mofei1/gpush/internal/socket/manager"

func InitConnectionManager() {
	manager.NewConnectionManager()
}
