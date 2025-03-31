package stat

import "github.com/prometheus/client_golang/prometheus"

type PrometheusStat struct {
	WsConnections      *prometheus.GaugeVec
	WsTotalConnections *prometheus.CounterVec
	WsBytesSent        *prometheus.CounterVec
	WsMessagesSent     *prometheus.CounterVec
}
