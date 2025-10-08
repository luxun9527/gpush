package handler

import (
	"github.com/ikun2021/gpush/internal/socket/manager"
	"github.com/ikun2021/gpush/internal/socket/model/request"
	"github.com/ikun2021/gpush/internal/socket/model/response"
)

type SubPublic struct {
}

func (SubPublic) Handle(r request.Message, conn *manager.Connection) {
	manager.CM.JoinPublicRoom(r.Topic, conn)
	conn.Send(response.SubSuccess)
}
