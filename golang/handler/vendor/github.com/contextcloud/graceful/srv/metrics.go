package srv

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewMetricsServer(metricsAddr string) Startable {
	return NewStandard(metricsAddr, promhttp.Handler())
}
