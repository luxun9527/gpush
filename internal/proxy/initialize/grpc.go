package initialize

import (
	"google.golang.org/grpc"
	"ws/internal/proxy/api"
	pb "ws/proto"
)

func InitGrpc() *grpc.Server {
	s := grpc.NewServer()
	t:=&api.ProxyApi{
		Data:                     make(chan *pb.Data,10000),
	}
	pb.RegisterProxyServer(s,t)
	return s
}