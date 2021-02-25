package daemon

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
)

var (
	availableCapacityGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "drone",
		Subsystem: "runner",
		Name: "available_capacity",
	}, []string{"name"})

	configuredCapacityGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "drone",
		Subsystem: "runner",
		Name: "configured_capacity",
	}, []string{"name"})
)

func setupMetrics(config Config) {
	http.Handle("/metrics", promhttp.Handler())
	logrus.Fatalf("failed to start metrics handler: %s", http.ListenAndServe(config.Metrics.Address, nil))
}