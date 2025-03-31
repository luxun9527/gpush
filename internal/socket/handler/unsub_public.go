package handler

import (
	"github.com/luxun9527/gpush/internal/socket/manager"
	"github.com/luxun9527/gpush/internal/socket/model/request"
	"github.com/luxun9527/gpush/internal/socket/model/response"
)

type UnSubPublic struct {
}

func (UnSubPublic) Handle(r request.Message, conn *manager.Connection) {

	manager.CM.LeavePublicRoom(r.Topic, conn)

	conn.Send(response.UnSubSuccess)
}
