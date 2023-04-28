package pull

import (
	"context"
	gws "github.com/gobwas/ws"
	"github.com/luxun9527/gpush/internal/socket/global"
	"github.com/luxun9527/gpush/internal/socket/manager"
	"github.com/luxun9527/gpush/internal/socket/model"
	pb "github.com/luxun9527/gpush/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"sync"
	"time"
)

// ProxyClient proxy的客户端，维护proxy连接
type ProxyClient struct {
	proxyConn sync.Map
}

func (pc *ProxyClient) PullData() {
	for _, addr := range global.Config.Proxy.Addrs {
		if err := pc.pullData(addr); err != nil {
			global.L.Error("pull data failed", zap.Error(err))
		}
	}
}
func newProxyClient() ProxyClient {
	return ProxyClient{}
}
func InitProxyConn() {
	client := newProxyClient()
	client.PullData()
}
func (pc *ProxyClient) pullData(addr string) error {
	c, ok := pc.proxyConn.Load(addr)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if !ok {
		conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			return err
		}
		c = conn
		pc.proxyConn.Store(addr, conn)
	}
	conn := c.(*grpc.ClientConn)

	client := pb.NewProxyClient(conn)
	stream, err := client.PullData(context.Background(), &pb.Empty{})
	if err != nil {
		global.L.Error("pull data failed err ", zap.Any("err", err))
		return err
	}
	global.L.Debug("connect to proxy", zap.Any("data", conn.Target()))
	go func() {
		for {
			data, err := stream.Recv()
			if err != nil {
				global.L.Error("pull data failed err ", zap.Any("err", err))
				time.Sleep(time.Second * 3)
				continue
			}
			message := model.NewMessage(gws.OpText, data.Data)
			var messageData []byte
			if global.Config.Connection.IsCompress {
				messageData, err = message.ToCompressBytes()
			} else {
				messageData, err = message.ToBytes()
			}
			if err != nil {
				global.L.Error("init message failed", zap.Error(err))
				continue
			}
			if data.Uid != "" {
				manager.CM.PushPerson(data.Uid, data.Topic, messageData)
			} else {
				if data.Topic == "" {
					manager.CM.PushAll(messageData)
				} else {
					manager.CM.PushRoom(data.Topic, messageData)
				}
			}
		}
	}()
	return nil
}
