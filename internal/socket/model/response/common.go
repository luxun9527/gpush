package response

import (
	"encoding/json"
	gws "github.com/gobwas/ws"
	"ws/internal/socket/global"
	"ws/internal/socket/model"
)

type Response struct {
	Code    int32       `json:"code"`
	Message string      `json:"message"`
	Detail  interface{} `json:"detail,omitempty"`
}

func (resp Response) EncodeMessage() []byte {
	r, _ := json.Marshal(resp)
	m := model.NewMessage(gws.OpText, r)
	var data []byte
	if global.Config.Connection.IsCompress {
		data, _ = m.ToCompressBytes()
	} else {
		data, _ = m.ToBytes()
	}

	return data
}

func EncodePONG() []byte {
	msg := model.NewMessage(gws.OpText, []byte("pong"))
	var data []byte
	if global.Config.Connection.IsCompress {
		data, _ = msg.ToCompressBytes()
	} else {
		data, _ = msg.ToBytes()
	}
	return data
}

var (
	PONG           = EncodePONG()
	ConnectSuccess = Response{
		Code:    200,
		Message: "连接成功",
	}.EncodeMessage()
	SubSuccess = Response{
		Code:    201,
		Message: "订阅成功",
	}.EncodeMessage()
	UnSubSuccess = Response{
		Code:    202,
		Message: "取消订阅成功",
	}.EncodeMessage()
	Failed = Response{
		Code:    203,
		Message: "操作有误",
		Detail:  "数据格式不正确",
	}.EncodeMessage()
	SubWithoutLogin = Response{
		Code:    204,
		Message: "操作有误",
		Detail:  "未登录",
	}.EncodeMessage()
)
