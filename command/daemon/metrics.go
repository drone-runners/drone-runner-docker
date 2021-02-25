package daemon

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
)

var (
	runnerCapacityGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "drone",
		Subsystem: "runner",
		Name: "capacity",
	}, []string{"hostname"})
)

func setupMetrics(config Config) {
	http.Handle("/metrics", promhttp.Handler())
	logrus.Fatalf("failed to start metrics handler: %s", http.ListenAndServe(config.Metrics.Address, nil))
}