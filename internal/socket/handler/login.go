package handler

import (
	"encoding/json"
	"github.com/luxun9527/zlog"
	"github.com/spf13/cast"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/luxun9527/gpush/internal/socket/global"
	"github.com/luxun9527/gpush/internal/socket/manager"
	"github.com/luxun9527/gpush/internal/socket/model/request"
	"github.com/luxun9527/gpush/internal/socket/model/response"
	"go.uber.org/zap"
)

type Login struct {
	HttpClient *resty.Client
}
type AuthSuccess struct {
	Uid      string `json:"uid"`
	Username string `json:"username"`
}

func (l Login) Handle(r request.Message, conn *manager.Connection) {
	zlog.Debug("receive login req", zap.Any("data", r))
	resp, err := l.HttpClient.R().
		SetBody(gin.H{"token": r.Data}).
		Post(global.Config.AuthUrl)
	if err != nil {
		zlog.Error("http auth client failed", zap.Error(err))
		conn.Send(response.TokenValidateFailed)
		return
	}
	var data gin.H
	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		zlog.Error("json unmarshal failed", zap.Error(err))
		conn.Send(response.TokenValidateFailed)
		return
	}
	zlog.Debug("receive login req", zap.Any("data", data))
	if cast.ToInt32(data["code"]) == 0 {
		userInfo := data["data"].(map[string]interface{})["user_info"].(map[string]interface{})
		uid := userInfo["uid"].(string)
		conn.SetUid(uid)
		manager.CM.Login(conn)
		conn.Send(response.LoginSuccess)
	} else {
		conn.Send(response.TokenValidateFailed)
	}

}
