package initialize

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/mofei1/gpush/internal/proxy/api"
	"github.com/mofei1/gpush/internal/proxy/global"
	pb "github.com/mofei1/gpush/proto"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
)

func InitGrpc() {
	s := grpc.NewServer()
	t := &api.ProxyApi{
		Data: make(chan *pb.Data, 10000),
	}
	t.PushSocketData()
	pb.RegisterProxyServer(s, t)
	listener, err := net.Listen("tcp", global.Config.Server.PullPort)
	if err != nil {
		log.Panicf("init proxy grpc  failed %v", err)
	}
	go func() {
		if err := s.Serve(listener); err != nil {
			log.Panicf("init proxy serve fialed err = %v", err.Error())
		}
	}()

}
func InitHttpServer() {
	conn, err := grpc.Dial(
		"127.0.0.1"+global.Config.Server.PullPort,
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Panicf("dail proxy grpc serve fialed err = %v", err)
	}

	gwmux := runtime.NewServeMux()
	if err = pb.RegisterProxyHandler(context.Background(), gwmux, conn); err != nil {
		log.Panicf("Failed to register gateway %v", err)
	}

	gwServer := &http.Server{
		Addr:    global.Config.Server.HttpPort,
		Handler: gwmux,
	}
	go func() {
		if err := gwServer.ListenAndServe(); err != nil {
			log.Panicf("init proxy http serve fialed err = %v", err)
		}
	}()

}
