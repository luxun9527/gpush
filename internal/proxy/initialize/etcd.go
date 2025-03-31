package initialize

import (
	"context"
	"github.com/luxun9527/gpush/internal/proxy/global"
	"github.com/luxun9527/gpush/tools"
	"github.com/luxun9527/zlog"
	"github.com/spf13/cast"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type EtcdClient struct {
	cli     *clientv3.Client
	leaseID clientv3.LeaseID
}

func InitEtcd() {
	cli, err := global.Config.EtcdConfig.BuildClient()
	if err != nil {
		zlog.Panic("init etcd client failed", zap.Error(err))
	}
	ec := &EtcdClient{
		cli: cli,
	}
	if err := ec.resister(); err != nil {
		zlog.Panic("init etcd client failed", zap.Error(err))
	}
}
func (ec EtcdClient) resister() error {
	resp, err := ec.cli.Grant(context.Background(), 5)
	if err != nil {
		return err
	}
	ec.leaseID = resp.ID
	ip, _ := tools.GetLocalIP()
	socket := ip + global.Config.Server.PullPort
	key := global.Config.EtcdConfig.KeyPrefix + "/" + cast.ToString(int64(resp.ID))
	if _, err := ec.cli.Put(context.Background(), key, socket, clientv3.WithLease(resp.ID)); err != nil {
		return err
	}
	kc, err := ec.cli.KeepAlive(context.Background(), resp.ID)
	zlog.Debug("register to etcd ", zap.Any("key", key))
	if err != nil {
		return err
	}
	go func() {
		if err := ec.listenLease(kc); err != nil {
			zlog.Warn("close client fail", zap.Error(err))
		}
	}()
	return nil
}
func (ec EtcdClient) listenLease(response <-chan *clientv3.LeaseKeepAliveResponse) error {
	for _ = range response {
		//		zlog.Info("receive", zap.Any("r", r.ID))
	}
	zlog.Warn("listen lease chan finish", zap.Any("leaseID", ec.leaseID))
	return ec.close()
}
func (ec EtcdClient) close() error {
	if _, err := ec.cli.Revoke(context.Background(), ec.leaseID); err != nil {
		return err
	}
	return ec.cli.Close()

}
