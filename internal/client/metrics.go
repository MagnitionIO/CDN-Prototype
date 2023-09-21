package client

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	requests *prometheus.CounterVec
	latency  prometheus.Histogram
}

func newMetrics() *metrics {
	return &metrics{
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "client_requests_total",
				Help: "Number of HTTP requests",
			},
			[]string{"reqType"},
		),
		latency: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name: "request_latency_ms",
				Help: "Histogram of latencies for HTTP requests",
				// Buckets: prometheus.DefBuckets,
				Buckets: []float64{0.5, 1, 2.5, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
			},
		),
	}
}
