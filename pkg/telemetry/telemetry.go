package telemetry

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
)

type Telemetry struct {
	MetricsHandler http.Handler
	provider       *metricsdk.MeterProvider
}

func New() (*Telemetry, error) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	exporter, err := otelprom.New(otelprom.WithRegisterer(registry))
	if err != nil {
		return nil, err
	}

	provider := metricsdk.NewMeterProvider(metricsdk.WithReader(exporter))
	otel.SetMeterProvider(provider)

	return &Telemetry{
		MetricsHandler: promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
		provider:       provider,
	}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	return t.provider.Shutdown(ctx)
}
