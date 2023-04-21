package initialize

import (
	"github.com/mofei1/gpush/internal/proxy/api"
	pb "github.com/mofei1/gpush/proto"
	"google.golang.org/grpc"
)

func InitGrpc() *grpc.Server {
	s := grpc.NewServer()
	t := &api.ProxyApi{
		Data: make(chan *pb.Data, 10000),
	}
	t.PushSocketData()
	pb.RegisterProxyServer(s, t)
	return s
}
