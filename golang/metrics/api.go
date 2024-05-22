package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

type Collector interface {
	GetHttpHandler() http.Handler
	RegisterMetric(metric prometheus.Collector)
	Start()
	// todo: add stop
	WithPrefix(prefix string) Collector
}
