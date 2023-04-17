package main

import (
	"flag"
	"github.com/mofei1/gpush/internal/proxy/api"
	"github.com/mofei1/gpush/internal/proxy/global"
	"github.com/mofei1/gpush/internal/proxy/initialize"
	"log"
	"net"
	"time"
)

func main() {
	var addr string
	flag.StringVar(&addr, "config", "config/proxy/config.toml", "配置文件路径")
	flag.Parse()
	initialize.InitConfig(addr)
	server := initialize.InitGrpc()

	listener, err := net.Listen("tcp", global.Config.Server.PullPort)
	if err != nil {
		log.Panicf("net listen err = %v", err.Error())
	}
	go func() {
		for {
			time.Sleep(time.Second * 10)
			log.Println("发送数量", api.SendCount.Load())

		}
	}()
	if err := server.Serve(listener); err != nil {
		log.Panicf("init proxy serve fialed err = %v", err.Error())
	}
}
