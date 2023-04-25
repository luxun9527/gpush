package api

import (
	"context"
	"github.com/luxun9527/gpush/internal/proxy/global"
	pb "github.com/luxun9527/gpush/proto"
	"go.uber.org/zap"
	"sync"
)

type ProxyApi struct {
	pb.UnimplementedProxyServer
	Data    chan *pb.Data
	reqConn sync.Map
}

type SocketConnection struct {
	req    pb.Proxy_PullDataServer
	cancel context.CancelFunc
}

// PushData 后端调用次接口推送数据
func (p *ProxyApi) PushData(c context.Context, d *pb.Data) (*pb.Empty, error) {
	p.Data <- d
	return &pb.Empty{}, nil
}

// PullData api调用此接口获取推送的数据
func (p *ProxyApi) PullData(e *pb.Empty, req pb.Proxy_PullDataServer) error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	s := &SocketConnection{
		req:    req,
		cancel: cancelFunc,
	}
	p.reqConn.Store(s, struct{}{})
	<-ctx.Done()
	return nil
}

// PushSocketData 推送数据给socket
func (p *ProxyApi) PushSocketData() {
	go func() {
		for data := range p.Data {
			p.reqConn.Range(func(req, _ any) bool {
				conn := req.(*SocketConnection)
				if err := conn.req.Send(data); err != nil {
					global.L.Error("send message failed", zap.Error(err))
					conn.cancel()
					p.reqConn.Delete(req)
				}
				return true
			})
		}

	}()

}
