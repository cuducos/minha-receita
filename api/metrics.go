package api

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricLabels = []string{"method", "status_code", "endpoint"}
	requestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "total_requests",
		Help: "The total number of requests served",
	}, metricLabels)
	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "request_duration",
		Help: "The duration of requests in milliseconds",
	}, metricLabels)
)

func registerMetric(e, m string, s int, i int64) {
	requestCount.WithLabelValues(m, fmt.Sprintf("%d", s), e).Inc()
	requestDuration.WithLabelValues(m, fmt.Sprintf("%d", s), e).Observe(float64(time.Now().UnixMilli() - i))
}
