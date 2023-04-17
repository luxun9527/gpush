package api

import (
	"context"
	pb "github.com/mofei1/gpush/proto"
	"go.uber.org/atomic"
	"log"
)

type ProxyApi struct {
	pb.UnimplementedProxyServer
	Data chan *pb.Data
}

var SendCount atomic.Int64

// PushData 后端调用次接口推送数据
func (p *ProxyApi) PushData(c context.Context, d *pb.Data) (*pb.Empty, error) {
	p.Data <- d
	return &pb.Empty{}, nil
}

// PullData api调用此接口获取推送的数据
func (p *ProxyApi) PullData(e *pb.Empty, req pb.Proxy_PullDataServer) error {
	for data := range p.Data {
		//SendCount.Inc()
		if err := req.Send(data); err != nil {
			log.Println("err", err)
		}
	}
	return nil
}
