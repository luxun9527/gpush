package handler

import (
	"github.com/ikun2021/gpush/internal/socket/manager"
	"github.com/ikun2021/gpush/internal/socket/model/request"
	"github.com/ikun2021/gpush/internal/socket/model/response"
)

type LoginOut struct {
}

func (LoginOut) Handle(r request.Message, conn *manager.Connection) {
	conn.SetUid("")
	if !conn.IsLogin() {
		conn.Send(response.NotLogin)
		return
	}
	manager.CM.Logout(conn)
	conn.Send(response.LogoutSuccess)
}
