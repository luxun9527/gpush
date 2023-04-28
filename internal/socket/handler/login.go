package handler

import (
	"github.com/luxun9527/gpush/internal/socket/manager"
	"github.com/luxun9527/gpush/internal/socket/model/request"
	"github.com/luxun9527/gpush/internal/socket/model/response"
)

type Login struct {
}

func (Login) Handle(r request.Message, conn *manager.Connection) {
	//todo 认证
	conn.SetUid("123456")
	manager.CM.Login(conn)
	conn.Send(response.LoginSuccess)
}
