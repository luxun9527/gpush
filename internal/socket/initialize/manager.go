package initialize

import "ws/internal/socket/manager"

func InitConnectionManager() {
	manager.NewConnectionManager()
}
