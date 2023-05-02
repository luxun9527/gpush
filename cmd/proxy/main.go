package main

import (
	"flag"
	"github.com/luxun9527/gpush/internal/proxy/initialize"
)

func main() {
	var addr string
	flag.StringVar(&addr, "config", "config/proxy/config.toml", "配置文件路径")
	flag.Parse()
	initialize.InitConfig(addr)
	initialize.InitLogger()
	initialize.InitGrpc()
	initialize.InitHttpServer()
	initialize.InitEtcd()
	select {}
}
