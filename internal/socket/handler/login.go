package handler

import (
	"github.com/mofei1/gpush/internal/socket/manager"
	"github.com/mofei1/gpush/internal/socket/model/request"
)

type Login struct {
}

func (Login) Handle(r request.Message, conn *manager.Connection) {
	//todo 认证
	conn.SetLoginStatus(true)
	manager.CM.Login(conn)
}
