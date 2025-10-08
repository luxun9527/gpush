package handler

import (
	"github.com/ikun2021/gpush/internal/socket/manager"
	"github.com/ikun2021/gpush/internal/socket/model/request"
	"github.com/ikun2021/gpush/internal/socket/model/response"
)

type Ping struct {
}

func (Ping) Handle(r request.Message, conn *manager.Connection) {
	conn.KeepAlive()
	conn.Send(response.PONG)
}
