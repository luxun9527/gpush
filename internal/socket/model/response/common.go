package response

import (
	"encoding/json"

	gws "github.com/gobwas/ws"
	"github.com/luxun9527/gpush/internal/socket/global"
	"github.com/luxun9527/gpush/internal/socket/model"
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
		Code:    20000,
		Message: "连接成功",
	}.EncodeMessage()
	SubSuccess = Response{
		Code:    20001,
		Message: "订阅成功",
	}.EncodeMessage()
	UnSubSuccess = Response{
		Code:    20002,
		Message: "取消订阅成功",
	}.EncodeMessage()
	LoginSuccess = Response{
		Code:    20003,
		Message: "登录成功",
	}.EncodeMessage()
	LogoutSuccess = Response{
		Code:    20003,
		Message: "退出成功",
	}.EncodeMessage()
	Failed = Response{
		Code:    30001,
		Message: "操作有误",
		Detail:  "数据格式不正确",
	}.EncodeMessage()
	NotLogin = Response{
		Code:    30002,
		Message: "操作有误",
		Detail:  "未登录",
	}.EncodeMessage()
	TokenValidateFailed = Response{
		Code:    30003,
		Message: "token验证失败",
		Detail:  "token验证失败",
	}.EncodeMessage()
)
