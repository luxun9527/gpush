package handler

import (
	"github.com/mofei1/gpush/internal/socket/manager"
	"github.com/mofei1/gpush/internal/socket/model/request"
)

type LoginOut struct {
}

func (LoginOut) Handle(r request.Message, conn *manager.Connection) {
	//fixme 退出
	conn.SetLoginStatus(false)

}
