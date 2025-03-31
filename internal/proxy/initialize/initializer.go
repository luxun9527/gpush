package initialize

func Init(addr string) {
	InitConfig(addr)
	InitGrpc()
	InitHttpServer()
	InitEtcd()
}
