package initialize

import (
	"github.com/luxun9527/gpush/internal/socket/global"
	"github.com/luxun9527/gpush/internal/socket/manager"
	"github.com/luxun9527/gpush/internal/socket/pull"
	"github.com/luxun9527/gpush/internal/socket/server"
	"github.com/luxun9527/zlog"
)

func Init(addr string) {
	InitConfig(addr)
	zlog.InitDefaultLogger(&global.Config.Logger)
	global.Prometheus = NewPrometheusStat()
	manager.InitConnectionManager()
	pull.InitProxyClientManager()
	server.InitHttpServer()

}
