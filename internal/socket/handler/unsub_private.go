package handler

import (
	"github.com/mofei1/gpush/internal/socket/manager"
	"github.com/mofei1/gpush/internal/socket/model/request"
	"github.com/mofei1/gpush/internal/socket/model/response"
)

type UnSubPrivate struct {
}

func (UnSubPrivate) Handle(r request.Message, conn *manager.Connection) {
	if !conn.IsLogin() {
		conn.Send(response.NotLogin)
		return
	}
	manager.CM.LeavePrivateRoom(r.Topic, conn)
	conn.Send(response.UnSubSuccess)
}
