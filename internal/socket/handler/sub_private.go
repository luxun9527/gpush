package handler

import (
	"github.com/luxun9527/gpush/internal/socket/manager"
	"github.com/luxun9527/gpush/internal/socket/model/request"
	"github.com/luxun9527/gpush/internal/socket/model/response"
)

type SubPrivate struct {
}

func (SubPrivate) Handle(r request.Message, conn *manager.Connection) {
	if !conn.IsLogin() {
		conn.Send(response.NotLogin)
		return
	}
	manager.CM.JoinPrivateRoom(r.Topic, conn)
	conn.Send(response.SubSuccess)
}
