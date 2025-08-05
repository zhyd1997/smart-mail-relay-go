package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	PullCount        prometheus.Counter
	MatchCount       prometheus.Counter
	ForwardSuccesses prometheus.Counter
	ForwardFailures  prometheus.Counter
	ProcessingTime   prometheus.Histogram
	ActiveRules      prometheus.Gauge
	TotalRules       prometheus.Gauge
}

// NewMetrics creates new Prometheus metrics
func NewMetrics() *Metrics {
	return &Metrics{
		PullCount: promauto.NewCounter(prometheus.CounterOpts{
			Name: "smart_mail_relay_pull_count",
			Help: "Total number of email fetch operations",
		}),
		MatchCount: promauto.NewCounter(prometheus.CounterOpts{
			Name: "smart_mail_relay_match_count",
			Help: "Total number of emails that matched forwarding rules",
		}),
		ForwardSuccesses: promauto.NewCounter(prometheus.CounterOpts{
			Name: "smart_mail_relay_forward_successes",
			Help: "Total number of successful email forwards",
		}),
		ForwardFailures: promauto.NewCounter(prometheus.CounterOpts{
			Name: "smart_mail_relay_forward_failures",
			Help: "Total number of failed email forwards",
		}),
		ProcessingTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "smart_mail_relay_processing_duration_seconds",
			Help:    "Time spent processing emails",
			Buckets: prometheus.DefBuckets,
		}),
		ActiveRules: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "smart_mail_relay_active_rules",
			Help: "Number of currently active forwarding rules",
		}),
		TotalRules: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "smart_mail_relay_total_rules",
			Help: "Total number of forwarding rules (active and inactive)",
		}),
	}
}
