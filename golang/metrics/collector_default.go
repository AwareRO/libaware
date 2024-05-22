package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type CollectorDefault struct {
	registry *prometheus.Registry
}

func NewDefaultCollector() Collector {
	return &CollectorDefault{registry: prometheus.NewRegistry()}
}

func (c *CollectorDefault) WithPrefix(prefix string) Collector { return c }

func (c *CollectorDefault) RegisterMetric(metric prometheus.Collector) {
	c.registry.MustRegister(metric)
}

func (c *CollectorDefault) Start() {}

func (c *CollectorDefault) GetHttpHandler() http.Handler {
	return promhttp.HandlerFor(c.registry, promhttp.HandlerOpts{Registry: c.registry})
}
