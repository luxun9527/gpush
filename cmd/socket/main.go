package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/mofei1/gpush/internal/socket/global"
	"github.com/mofei1/gpush/internal/socket/initialize"
	"github.com/mofei1/gpush/internal/socket/pull"
	"github.com/mofei1/gpush/internal/socket/service"
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
	flag.StringVar(&addr, "config", "config/ws/config.toml", "配置文件路径")
	flag.Parse()

	r := gin.New()
	initialize.InitConfig(addr)
	initialize.InitLogger()
	initialize.InitConnectionManager()
	pull.InitProxyConn()
	r.GET("/ws", service.Connect)
	r.GET("/stats", service.Stats)
	//go func() {
	//	for {
	//		time.Sleep(time.Second)
	//		last := manager.LastReceived.Load()
	//		current := manager.Received.Load()
	//		perSecond := current - last
	//		manager.LastReceived = manager.Received
	//		log.Printf("总共收到 %v 平均每秒收到%v 跳过 %v sendcount %v,从proxy收到 %v", current, perSecond, manager.Ship.Load(), manager.SendCount.Load(), pull.Received.Load())
	//	}
	//}()
	global.L.Info("load config path success", zap.Any("detail", global.Config))
	r.Run(global.Config.Server.Port)

}
