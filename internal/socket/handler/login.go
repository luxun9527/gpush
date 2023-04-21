package handler

import (
	"github.com/mofei1/gpush/internal/socket/manager"
	"github.com/mofei1/gpush/internal/socket/model/request"
	"github.com/mofei1/gpush/internal/socket/model/response"
)

type Login struct {
}

func (Login) Handle(r request.Message, conn *manager.Connection) {
	//todo 认证
	conn.SetLoginStatus(true)
	conn.Uid = "123456"
	manager.CM.Login(conn)
	conn.Send(response.LoginSuccess)
}
