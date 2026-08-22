package middleware

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
)

const metricsNamespace = "http"

type Prometheus struct {
	requestsTotal *prometheus.CounterVec
	duration      *prometheus.HistogramVec
}

func NewPrometheus(reg prometheus.Registerer) *Prometheus {
	p := &Prometheus{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: metricsNamespace,
				Name:      "requests_total",
				Help:      "Total HTTP requests processed.",
			},
			[]string{"method", "route", "status"},
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: metricsNamespace,
				Name:      "request_duration_seconds",
				Help:      "HTTP request latency in seconds.",
				Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
			},
			[]string{"method", "route"},
		),
	}
	reg.MustRegister(p.requestsTotal, p.duration)
	return p
}

// Handler는 요청을 계측한다. 라벨에는 파라미터가 아닌 라우트 패턴(/api/v1/tasks/:id)을 사용해
// 카디널리티 폭발을 방지한다.
func (p *Prometheus) Handler() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := timeNow()
		err := c.Next()

		route := c.Route().Path
		if route == "" {
			route = "unmatched"
		}
		status := strconv.Itoa(c.Response().StatusCode())
		p.requestsTotal.WithLabelValues(c.Method(), route, status).Inc()
		p.duration.WithLabelValues(c.Method(), route).Observe(timeSince(start))
		return err
	}
}
