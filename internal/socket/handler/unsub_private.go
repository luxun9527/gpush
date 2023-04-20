package handler

import (
	"github.com/mofei1/gpush/internal/socket/manager"
	"github.com/mofei1/gpush/internal/socket/model/request"
)

type UnSubPrivate struct {
}

func (UnSubPrivate) Handle(r request.Message, conn *manager.Connection) {

	//todo
}
