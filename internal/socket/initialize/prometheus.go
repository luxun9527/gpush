package initialize

import (
	"github.com/luxun9527/gpush/internal/socket/stat"
	"github.com/prometheus/client_golang/prometheus"
)

func NewPrometheusStat() *stat.PrometheusStat {
	var (
		// WebSocket 当前连接数（按路径分组）
		wsConnections = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "websocket_current_connections",
				Help: "Current number of WebSocket connections",
			},
			nil,
		)

		// WebSocket 总连接数（累计）
		wsTotalConnections = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "websocket_total_connections",
				Help: "Total number of WebSocket connections since startup",
			},
			nil,
		)

		// WebSocket 推送字节数
		wsBytesSent = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "websocket_bytes_sent_total",
				Help: "Total number of bytes sent via WebSocket",
			},
			nil,
		)

		// WebSocket 推送消息数
		wsMessagesSent = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "websocket_messages_sent_total",
				Help: "Total number of messages sent via WebSocket",
			},
			nil,
		)
	)
	prometheus.Register(wsConnections)
	prometheus.Register(wsTotalConnections)
	prometheus.Register(wsBytesSent)
	prometheus.Register(wsMessagesSent)
	return &stat.PrometheusStat{
		WsConnections:      wsConnections,
		WsTotalConnections: wsTotalConnections,
		WsBytesSent:        wsBytesSent,
		WsMessagesSent:     wsMessagesSent,
	}
}
