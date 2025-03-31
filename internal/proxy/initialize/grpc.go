package initialize

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/luxun9527/gpush/internal/proxy/api"
	"github.com/luxun9527/gpush/internal/proxy/global"
	pb "github.com/luxun9527/gpush/proto"
	"github.com/luxun9527/zlog"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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
		zlog.Panic("tcp listen failed", zap.Error(err))
	}
	go func() {
		if err := s.Serve(listener); err != nil {
			zlog.Panic("init grpc server failed ", zap.Error(err))
		}
	}()

}
func InitHttpServer() {
	conn, err := grpc.Dial(
		"127.0.0.1"+global.Config.Server.PullPort,
		grpc.WithInsecure(),
	)
	if err != nil {
		zlog.Panic("dail proxy grpc serve failed ", zap.Error(err))
	}

	gwmux := runtime.NewServeMux()
	if err = pb.RegisterProxyHandler(context.Background(), gwmux, conn); err != nil {
		zlog.Panic("Failed to register gateway ", zap.Error(err))
	}

	gwServer := &http.Server{
		Addr:    global.Config.Server.HttpPort,
		Handler: gwmux,
	}
	go func() {
		if err := gwServer.ListenAndServe(); err != nil {
			zlog.Panic("init proxy http serve failed err", zap.Error(err))

		}
	}()

}
