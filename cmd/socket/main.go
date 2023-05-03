package main

import (
	"flag"
	"github.com/luxun9527/gpush/internal/socket/global"
	"github.com/luxun9527/gpush/internal/socket/initialize"
	"github.com/luxun9527/gpush/internal/socket/manager"
	"github.com/luxun9527/gpush/internal/socket/pull"
	"github.com/luxun9527/gpush/internal/socket/server"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go func() {
		http.ListenAndServe("0.0.0.0:8899", nil)
	}()
	var addr string
	flag.StringVar(&addr, "config", "config/socket/config.toml", "配置文件路径")
	flag.Parse()

	initialize.InitConfig(addr)
	initialize.InitLogger()
	manager.InitConnectionManager()
	pull.InitProxyClientManager()
	server.InitHttpServer()
	global.L.Info("load config path success", zap.Any("detail", global.Config))
	select {}
}
