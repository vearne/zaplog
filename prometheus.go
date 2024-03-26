package zaplog

import (
	"github.com/prometheus/client_golang/prometheus"
)

var logTotal *prometheus.CounterVec

const PromLabelLevel = "level" //http method

func init() {
	logTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "log_total",
			Help: "The count of log event",
		},
		[]string{PromLabelLevel},
	)
	prometheus.MustRegister(logTotal)
}
