package handler

import (
	"encoding/json"
	"github.com/mofei1/gpush/internal/socket/manager"
	"github.com/mofei1/gpush/internal/socket/model/request"
	"github.com/mofei1/gpush/internal/socket/model/response"
	"strings"
)

//不同的操作对应不同的策略
var _strategy = map[request.Code]HandleStrategy{
	request.SubPublic:       SubPublic{},
	request.UnsubPublic:     UnSubPublic{},
	request.Login:           Login{},
	request.Logout:          LoginOut{},
	request.SubPrivate:      SubPrivate{},
	request.UnSubSubPrivate: UnSubPrivate{},
}

type HandleStrategy interface {
	Handle(request.Message, *manager.Connection)
}
type Handler struct {
	message  request.Message
	conn     *manager.Connection
	strategy HandleStrategy
}

func (p *Handler) handle() {
	p.strategy.Handle(p.message, p.conn)
}
func NewHandler(strategy HandleStrategy, message request.Message, conn *manager.Connection) Handler {
	return Handler{
		message:  message,
		conn:     conn,
		strategy: strategy,
	}
}

func Handle(data []byte, conn *manager.Connection) {

	if strings.ToLower(string(data)) == "ping" {
		h := NewHandler(Ping{}, request.Message{}, conn)
		h.handle()
		return
	}
	var message request.Message
	if err := json.Unmarshal(data, &message); err != nil {
		conn.Send(response.Failed)
		return
	}
	if _, ok := _strategy[message.Code]; !ok {
		conn.Send(response.Failed)
		return
	}
	h := NewHandler(_strategy[message.Code], message, conn)
	h.handle()

}
