package api

import (
	"context"
	pb "github.com/mofei1/gpush/proto"
	"go.uber.org/atomic"
	"log"
	"sync"
)

type ProxyApi struct {
	pb.UnimplementedProxyServer
	Data    chan *pb.Data
	reqConn sync.Map
}

var SendCount atomic.Int64

// PushData 后端调用次接口推送数据
func (p *ProxyApi) PushData(c context.Context, d *pb.Data) (*pb.Empty, error) {
	p.Data <- d
	return &pb.Empty{}, nil
}

// PullData api调用此接口获取推送的数据
func (p *ProxyApi) PullData(e *pb.Empty, req pb.Proxy_PullDataServer) error {
	p.reqConn.Store(req, struct{}{})
	return nil
}

// PushSocketData 推送数据给socket
func (p *ProxyApi) PushSocketData() error {
	go func() {
		for data := range p.Data {
			p.reqConn.Range(func(req, _ any) bool {
				r := req.(pb.Proxy_PullDataServer)
				if err := r.Send(data); err != nil {
					log.Println("err", err)
					p.reqConn.Delete(req)
				}
				return true
			})
		}

	}()

	return nil
}
