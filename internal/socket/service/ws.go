package service

import (
	"github.com/gin-gonic/gin"
	gws "github.com/gobwas/ws"
	"github.com/gobwas/ws/wsflate"
	"go.uber.org/zap"
	"ws/internal/socket/global"
	"ws/internal/socket/handler"
	"ws/internal/socket/manager"
	"ws/internal/socket/model/response"
)

func Connect(c *gin.Context) {
	httpUpgrade := gws.DefaultHTTPUpgrader
	if global.Config.Connection.IsCompress {
		e := wsflate.Extension{
			Parameters: wsflate.DefaultParameters,
		}
		httpUpgrade.Negotiate = e.Negotiate
	}
	conn, _, _, err := httpUpgrade.Upgrade(c.Request, c.Writer)
	if err != nil {
		global.L.Error("upgrade 连接失败", zap.Error(err))
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

func Stats(c *gin.Context) {
	count := manager.GetConnectionInfo()
	c.JSON(200, map[string]interface{}{"count": count})
}
func Push(c *gin.Context) {
	topic, _ := c.GetPostForm("topic")
	data, _ := c.GetPostForm("data")
	manager.CM.PushRoom(topic, []byte(data))
}
