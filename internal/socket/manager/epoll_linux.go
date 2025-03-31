//go:build linux
// +build linux

package manager

import (
	"errors"
	"github.com/luxun9527/zlog"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"syscall"
)

type Epoller struct {
	fd int
}

func NewEpoller() *Epoller {
	var err error
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		zlog.Panic("init epoll failed", zap.Any("err", err))
	}
	epoller := &Epoller{fd: fd}
	go epoller.run()
	return epoller
}

func (e *Epoller) run() {
	defer syscall.Close(e.fd)
	for {
		events := make([]unix.EpollEvent, 1024)
		n, err := unix.EpollWait(e.fd, events, -1)
		if errors.Is(err, syscall.EINTR) {
			continue
		}
		if errors.Is(err, syscall.EAGAIN) {
			break
		}
		for i := 0; i < n; i++ {
			switch events[i].Events {
			//case unix.EPOLLIN | unix.EPOLLRDHUP:

			case unix.EPOLLIN:
				CM.NotifyRead(int(events[i].Fd))
				//其他的操作视为关闭
			default:
				CM.CloseConnection(int(events[i].Fd))
			}

		}

	}
}

func (e *Epoller) Add(fd int) error {
	//使用et模式 一次触发
	return unix.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: unix.EPOLLERR | unix.EPOLLHUP | unix.EPOLLET | unix.EPOLLRDHUP | unix.EPOLLPRI | unix.EPOLLIN, Fd: int32(fd)})

}

func (e *Epoller) Remove(fd int) error {
	return unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
}
