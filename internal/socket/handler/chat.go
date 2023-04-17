package handler

import (
	"github.com/mofei1/gpush/internal/socket/manager"
	"github.com/mofei1/gpush/internal/socket/model/request"
	"github.com/mofei1/gpush/internal/socket/model/response"
)

type Chat struct {
}

func (Chat) Handle(r request.Message, conn *manager.Connection) {
	conn.Send(response.PONG)
}
