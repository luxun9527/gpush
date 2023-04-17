package handler

import (
	"github.com/mofei1/gpush/internal/socket/manager"
	"github.com/mofei1/gpush/internal/socket/model/request"
	"github.com/mofei1/gpush/internal/socket/model/response"
)

type Ping struct {
}

func (Ping) Handle(r request.Message, conn *manager.Connection) {
	conn.KeepAlive()
	conn.Send(response.PONG)
}
