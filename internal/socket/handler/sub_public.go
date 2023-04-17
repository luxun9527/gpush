package handler

import (
	"github.com/mofei1/gpush/internal/socket/manager"
	"github.com/mofei1/gpush/internal/socket/model/request"
	"github.com/mofei1/gpush/internal/socket/model/response"
)

type SubPublic struct {
}

func (SubPublic) Handle(r request.Message, conn *manager.Connection) {
	manager.CM.JoinRoom(r.Topic, conn)
	conn.Send(response.SubSuccess)
}
