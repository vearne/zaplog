package zaplog

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	myMeter  metric.Meter
	logTotal metric.Int64Counter
)

func init() {
	myMeter = otel.GetMeterProvider().Meter("zaplog")
	logTotal, _ = myMeter.Int64Counter("log_total",
		metric.WithDescription("the count of log event"))
}
