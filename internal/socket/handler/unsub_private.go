package handler

import (
	"github.com/luxun9527/gpush/internal/socket/manager"
	"github.com/luxun9527/gpush/internal/socket/model/request"
	"github.com/luxun9527/gpush/internal/socket/model/response"
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
