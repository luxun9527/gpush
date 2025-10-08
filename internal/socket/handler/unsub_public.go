package handler

import (
	"github.com/ikun2021/gpush/internal/socket/manager"
	"github.com/ikun2021/gpush/internal/socket/model/request"
	"github.com/ikun2021/gpush/internal/socket/model/response"
)

type UnSubPublic struct {
}

func (UnSubPublic) Handle(r request.Message, conn *manager.Connection) {

	manager.CM.LeavePublicRoom(r.Topic, conn)

	conn.Send(response.UnSubSuccess)
}
