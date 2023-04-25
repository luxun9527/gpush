package handler

import (
	"github.com/luxun9527/gpush/internal/socket/manager"
	"github.com/luxun9527/gpush/internal/socket/model/request"
	"github.com/luxun9527/gpush/internal/socket/model/response"
)

type LoginOut struct {
}

func (LoginOut) Handle(r request.Message, conn *manager.Connection) {
	conn.SetLoginStatus(false)
	if !conn.IsLogin() {
		conn.Send(response.NotLogin)
		return
	}
	manager.CM.Logout(conn)
	conn.Send(response.LogoutSuccess)
}
