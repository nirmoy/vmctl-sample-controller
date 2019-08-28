package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	k8sEventCounter prometheus.Gauge
}

func (m *Metrics) K8sEventUpdate() {
	m.k8sEventCounter.SetToCurrentTime()
}

func InitMetrics(address string) *Metrics {
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":2112", nil)
	return &Metrics{
		k8sEventCounter: promauto.NewGauge(k8sEventCounter),
	}

}

var (
	k8sEventCounter = prometheus.GaugeOpts{
		Name: "k8s_processed_ops_total",
		Help: "The total number of processed events",
	}
)
