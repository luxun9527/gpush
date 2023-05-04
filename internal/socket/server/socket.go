package server

import (
	"github.com/gin-gonic/gin"
	gws "github.com/gobwas/ws"
	"github.com/gobwas/ws/wsflate"
	"github.com/luxun9527/gpush/internal/socket/global"
	"github.com/luxun9527/gpush/internal/socket/handler"
	"github.com/luxun9527/gpush/internal/socket/manager"
	"github.com/luxun9527/gpush/internal/socket/model/response"
	"go.uber.org/zap"
)

func InitHttpServer() {
	r := gin.New()
	r.GET("/ws", connect)
	r.GET("/stats", stats)
	go func() {
		if err := r.Run("0.0.0.0:" + global.Config.Server.Port); err != nil {
			global.L.Panic("init http server failed", zap.Error(err))
		}
	}()

}
func connect(c *gin.Context) {
	httpUpgrade := gws.DefaultHTTPUpgrader
	if global.Config.Connection.IsCompress {
		e := wsflate.Extension{
			Parameters: wsflate.DefaultParameters,
		}
		httpUpgrade.Negotiate = e.Negotiate
	}
	conn, _, _, err := httpUpgrade.Upgrade(c.Request, c.Writer)
	if err != nil {
		global.L.Error("upgrade failed", zap.Error(err))
		return
	}
	_, err = conn.Write(response.ConnectSuccess)
	if err != nil {
		global.L.Error("write data to connect failed", zap.Error(err))
		return
	}
	connection, err := manager.NewConnection(conn, handler.Handle)
	if err != nil {
		connection.Close()
	}

}

func stats(c *gin.Context) {
	count := manager.GetConnectionInfo()
	c.JSON(200, map[string]interface{}{"count": count})
}
