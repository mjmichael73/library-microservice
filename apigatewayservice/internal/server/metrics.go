package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	inFlightGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_inflight_requests",
			Help: "Number of in-flight requests being handled",
		},
		[]string{"path"},
	)

	requestSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "http_request_size_bytes",
			Help:       "Size of HTTP requests",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"method", "path"},
	)

	responseSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "http_response_size_bytes",
			Help:       "Size of HTTP responses",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"method", "path"},
	)

	appErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_errors_total",
			Help: "Number of internal application errors",
		},
		[]string{"path"},
	)
)

func init() {
	prometheus.MustRegister(httpRequests, httpDuration, inFlightGauge, requestSize, responseSize, appErrors)
}

func MetricsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		res := c.Response()

		path := c.Path()
		method := req.Method

		// Increment in-flight
		inFlightGauge.WithLabelValues(path).Inc()
		defer inFlightGauge.WithLabelValues(path).Dec()

		// Track request size
		if req.ContentLength > 0 {
			requestSize.WithLabelValues(method, path).Observe(float64(req.ContentLength))
		}

		// Duration tracking
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(method, path))
		defer timer.ObserveDuration()

		// Execute actual handler
		err := next(c)

		// Track response size
		responseSize.WithLabelValues(method, path).Observe(float64(res.Size))

		// Track status
		status := res.Status
		statusText := http.StatusText(status)
		httpRequests.WithLabelValues(method, path, statusText).Inc()

		// Track app errors (only 5xx)
		if status >= 500 {
			appErrors.WithLabelValues(path).Inc()
		}

		return err
	}
}
