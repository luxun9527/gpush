package pull

import (
	"context"
	gws "github.com/gobwas/ws"
	"github.com/luxun9527/gpush/internal/socket/global"
	"github.com/luxun9527/gpush/internal/socket/manager"
	"github.com/luxun9527/gpush/internal/socket/model"
	pb "github.com/luxun9527/gpush/proto"
	"github.com/luxun9527/zlog"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

// ProxyClientManager ProxyClient proxy的客户端，维护proxy连接
type ProxyClientManager struct {
	cli          *clientv3.Client
	proxyClients map[string]*proxyClient
}

func newProxyClientManager() ProxyClientManager {
	cli, err := global.Config.ProxyRpc.BuildClient()
	if err != nil {
		zlog.Panic("init etcd client failed", zap.Error(err))
	}
	return ProxyClientManager{
		cli:          cli,
		proxyClients: make(map[string]*proxyClient, 2),
	}
}

func (c *ProxyClientManager) initProxyClientManager() {
	resp, err := c.cli.Get(context.Background(), global.Config.ProxyRpc.KeyPrefix, clientv3.WithPrefix())
	zlog.Info("get proxy", zap.Int("proxy num", len(resp.Kvs)))
	if err != nil {
		zlog.Panic("get proxy failed", zap.Error(err))
	}
	for _, v := range resp.Kvs {
		zlog.Info("get proxy success", zap.Int("proxy num", len(resp.Kvs)), zap.String("proxy addr", string(v.Value)), zap.String("key", string(v.Key)))
		ctx, cancel := context.WithCancel(context.Background())
		pc := &proxyClient{
			cancel: cancel,
			ctx:    ctx,
			addr:   string(v.Value),
		}
		if err := pc.pullDataFromProxy(); err != nil {
			zlog.Error("init proxy failed ", zap.Error(err))
			continue
		}
		c.proxyClients[string(v.Key)] = pc
	}

	go c.Watch()

}

// InitProxyClientManager 初始化连接proxy客户端管理
func InitProxyClientManager() {
	client := newProxyClientManager()
	client.initProxyClientManager()
}

func (c *ProxyClientManager) Watch() {

	for resp := range c.cli.Watch(context.Background(), global.Config.ProxyRpc.KeyPrefix, clientv3.WithPrefix()) {
		for _, ev := range resp.Events {
			switch ev.Type {
			case mvccpb.PUT: //修改或者新增
				zlog.Info("add or update etcd", zap.Any("data", string(ev.Kv.Value)))
				ctx, cancel := context.WithCancel(context.Background())
				pc := &proxyClient{
					cancel: cancel,
					ctx:    ctx,
					addr:   string(ev.Kv.Value),
				}
				if err := pc.pullDataFromProxy(); err != nil {
					zlog.Error("init proxy failed ", zap.Error(err))
					continue
				}
				c.proxyClients[string(ev.Kv.Key)] = pc
			case mvccpb.DELETE: //删除
				zlog.Info("delete etcd client", zap.Any("value", string(ev.Kv.Value)), zap.Any("key", string(ev.Kv.Key)))
				pc, ok := c.proxyClients[string(ev.Kv.Key)]
				if ok {
					pc.cancel()
					delete(c.proxyClients, string(ev.Kv.Key))
				}
			}
		}
	}
}

type proxyClient struct {
	cancel context.CancelFunc
	ctx    context.Context
	addr   string
}

func (pc *proxyClient) pullDataFromProxy() error {

	conn, err := grpc.DialContext(pc.ctx, pc.addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return err
	}
	client := pb.NewProxyClient(conn)
	stream, err := client.PullData(pc.ctx, &pb.Empty{})
	if err != nil {
		zlog.Error("pull data failed err ", zap.Any("err", err))
		return err
	}
	zlog.Info("connect to proxy", zap.Any("data", conn.Target()))
	go func() {
		for {
			select {
			case <-pc.ctx.Done():
				zlog.Info("disconnect proxy ", zap.Any("addr", pc.addr))
				return

			default:
				data, err := stream.Recv()
				if err != nil {
					zlog.Error("pull data failed err ", zap.Any("err", err))
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
					zlog.Error("init message failed", zap.Error(err))
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

		}
	}()
	return nil
}
