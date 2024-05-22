package middlewares

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AwareRO/libaware/golang/geoip"
	"github.com/AwareRO/libaware/golang/metrics"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

type MetricsConfig struct {
	App                string `toml:"app" yaml:"app" env:"METRICS_APP_VALUE"`
	PrometheusHost     string `toml:"prometheus_host" yaml:"prometheus_host" env:"PROMETHEUS_HOST"`
	PrometheusUsername string `toml:"prometheus_username" yaml:"prometheus_username" env:"PROMETHEUS_USERNAME"`
	PrometheusPassword string `toml:"prometheus_password" yaml:"prometheus_password" env:"PROMETHEUS_PASSWORD"`
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

type ResetTick func(wrapper *durationMetricWrapper)

type durationMetricWrapper struct {
	Collector          metrics.Collector
	durations          *prometheus.HistogramVec
	dailyRequests      *prometheus.CounterVec
	monthlyRequests    *prometheus.CounterVec
	resetTick          ResetTick
	app                string
	prometheusHost     string
	prometheusUsername string
	prometheusPassword string
}

func NewDefaultDurationMetricWrapper(conf MetricsConfig) *durationMetricWrapper {
	return (&durationMetricWrapper{}).init(metrics.NewDefaultCollector(), conf)
}

func NewDurationMetricWrapper(collector metrics.Collector, conf MetricsConfig) *durationMetricWrapper {
	return (&durationMetricWrapper{}).init(collector, conf)
}

func (wrapper *durationMetricWrapper) Wrap(nextHandler httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		recorder := &statusRecorder{w, 200}
		start := time.Now()

		nextHandler(recorder, r, params)

		elapsed := time.Since(start)
		endpoint := strings.Split(r.URL.String(), "?")[0]
		method := r.Method
		status := fmt.Sprintf("%d", recorder.status)
		ip := ipv4(r)
		crawler := isCrawler(r)
		lat := ""
		lon := ""
		country := "Unknown"

		find := geoip.NewIPApiFinder()
		loc, err := find(ip)
		if err == nil {
			lat = fmt.Sprintf("%f", loc.Latitude)
			lon = fmt.Sprintf("%f", loc.Longitude)
			country = loc.Country
		} else {
			log.Error().Err(err).Msg(fmt.Sprintf("Failed to get loc[%s]: %v", ip, loc))
		}

		wrapper.durations.
			WithLabelValues(wrapper.app, endpoint, method, status).Observe(float64(elapsed.Milliseconds()))
		wrapper.dailyRequests.
			WithLabelValues(wrapper.app, endpoint, method, status, ip, lat, lon, country, crawler).Inc()
		wrapper.monthlyRequests.
			WithLabelValues(wrapper.app, endpoint, method, status, ip, lat, lon, country, crawler).Inc()
	}
}

func (wrapper *durationMetricWrapper) init(collector metrics.Collector, conf MetricsConfig) *durationMetricWrapper {
	wrapper.resetTick = defaultTick
	wrapper.app = conf.App
	wrapper.prometheusHost = conf.PrometheusHost
	wrapper.prometheusUsername = conf.PrometheusUsername
	wrapper.prometheusPassword = conf.PrometheusPassword
	wrapper.initializeMetrics()
	wrapper.Collector = collector
	wrapper.registerMetrics()
	wrapper.restoreMetrics()

	go func() {
		logger := log.Info().Str("goroutine", "http metrics ticker")
		for {
			now := time.Now()
			nextWait := time.Until(time.Date(
				now.Year(), now.Month(), now.Day(),
				0, 0, 0, 0,
				now.Location(),
			).AddDate(0, 0, 1))
			logger.Msg(fmt.Sprintf("Sleeping %v", nextWait))
			time.Sleep(nextWait)
			logger.Msg("reseting counters")
			wrapper.resetTick(wrapper)
		}
	}()

	return wrapper
}

func defaultTick(wrapper *durationMetricWrapper) {
	wrapper.dailyRequests.Reset()
	if time.Now().Day() == 1 {
		wrapper.monthlyRequests.Reset()
	}
}

func (wrapper *durationMetricWrapper) initializeMetrics() {
	wrapper.durations = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "http_server",
			Name:      "request_duration_milliseconds",
			Help:      "Histogram of response time for handler in milliseconds",
			Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000, 30000},
		},
		[]string{"app", "endpoint", "method", "status"},
	)
	wrapper.dailyRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "http_server",
			Name:      "request_count_daily",
			Help:      "Counts daily http requests",
		},
		[]string{"app", "endpoint", "method", "status", "ip", "lat", "lon", "country", "crawler"},
	)
	wrapper.monthlyRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "http_server",
			Name:      "request_count_monthly",
			Help:      "Counts monthly http requests",
		},
		[]string{"app", "endpoint", "method", "status", "ip", "lat", "lon", "country", "crawler"},
	)
}

func (wrapper *durationMetricWrapper) registerMetrics() {
	wrapper.Collector.RegisterMetric(wrapper.durations)
	log.Info().Str("name", "http_server_request_duration_milliseconds").
		Str("type", "histogram_vec").
		Msg("registered new metric")
	wrapper.Collector.RegisterMetric(wrapper.dailyRequests)
	log.Info().Str("name", "http_server_request_count_daily").
		Str("type", "counter_vec").
		Msg("registered new metric")
	wrapper.Collector.RegisterMetric(wrapper.monthlyRequests)
	log.Info().Str("name", "http_server_request_count_monthly").
		Str("type", "counter_vec").
		Msg("registered new metric")
}

func (wrapper *durationMetricWrapper) restoreMetrics() {
	wrapper.restoreMetric("http_server_request_count_daily", wrapper.dailyRequests)
	wrapper.restoreMetric("http_server_request_count_monthly", wrapper.monthlyRequests)
}

func (wrapper *durationMetricWrapper) restoreMetric(name string, metric *prometheus.CounterVec) {
	logger := log.Error().Str("host", wrapper.prometheusHost)

	resp, err := prometheusRequest(
		wrapper.prometheusHost,
		wrapper.prometheusUsername,
		wrapper.prometheusPassword,
		name,
		wrapper.app,
	)
	if err != nil {
		logger.Err(err).Msg("Failed prometheus request")
		return
	}

	r := metrics.PrometheusResponse{}

	err = json.NewDecoder(resp).Decode(&r)
	if err != nil {
		logger.Err(err).Msg("Failed to decode prometheus response")
		resp.Close()

		return
	}

	resp.Close()

	for _, m := range r.Data.Result {
		value, _ := strconv.ParseInt(m.Value[1].(string), 10, 32)
		metric.WithLabelValues(wrapper.app,
			m.Metric.Endpoint, m.Metric.Method, m.Metric.Status,
			m.Metric.IP, m.Metric.Latitude, m.Metric.Longitude, m.Metric.Country, m.Metric.Crawler,
		).Add(float64(value))
	}
}

func prometheusRequest(host, user, password, metric, app string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(`%s/api/v1/query?query=%s{app="%s"}`, host, metric, app), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req.SetBasicAuth(user, password)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
