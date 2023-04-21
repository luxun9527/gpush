package handler

import (
	"github.com/mofei1/gpush/internal/socket/manager"
	"github.com/mofei1/gpush/internal/socket/model/request"
	"github.com/mofei1/gpush/internal/socket/model/response"
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
