package telemetry

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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

func (t *Telemetry) HTTPMiddleware() func(http.Handler) http.Handler {
	otelMiddleware := otelhttp.NewMiddleware("api",
		otelhttp.WithMeterProvider(t.provider),
		otelhttp.WithFilter(func(r *http.Request) bool { return r.URL.Path != "/metrics" && r.URL.Path != "/health" }),
	)
	return func(next http.Handler) http.Handler {
		return otelMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			// otelhttp never sees chi's route, so read it from the chi context after the handler runs
			if rctx := chi.RouteContext(r.Context()); rctx != nil && rctx.RoutePattern() != "" {
				labeler, _ := otelhttp.LabelerFromContext(r.Context())
				labeler.Add(attribute.String("http.route", rctx.RoutePattern()))
			}
		}))
	}
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	return t.provider.Shutdown(ctx)
}
