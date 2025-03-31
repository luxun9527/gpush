package handler

import (
	"github.com/luxun9527/gpush/internal/socket/manager"
	"github.com/luxun9527/gpush/internal/socket/model/request"
	"github.com/luxun9527/gpush/internal/socket/model/response"
)

type Ping struct {
}

func (Ping) Handle(r request.Message, conn *manager.Connection) {
	conn.KeepAlive()
	conn.Send(response.PONG)
}
