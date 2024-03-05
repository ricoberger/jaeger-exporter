package exporter

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/proto-gen/api_v2/metrics"
	"github.com/jaegertracing/jaeger/storage/spanstore"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

var (
	spansReceivedTotalMetric = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "jaeger_exporter",
		Name:      "spans_received_total",
	})

	callsTotalMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "",
		Name:      "calls_total",
	}, []string{"service_name", "span_name", "span_kind", "status_code"})

	durationMillisecondsMetric = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "",
		Name:      "duration_milliseconds",
		Buckets:   []float64{0.1, 1, 5, 10, 50, 100, 250, 500, 1000, 5000, 10000},
	}, []string{"service_name", "span_name", "span_kind", "status_code"})

	jaegerToOtelSpanKind = map[string]string{
		"unspecified": metrics.SpanKind_SPAN_KIND_UNSPECIFIED.String(),
		"internal":    metrics.SpanKind_SPAN_KIND_INTERNAL.String(),
		"server":      metrics.SpanKind_SPAN_KIND_SERVER.String(),
		"client":      metrics.SpanKind_SPAN_KIND_CLIENT.String(),
		"producer":    metrics.SpanKind_SPAN_KIND_PRODUCER.String(),
		"consumer":    metrics.SpanKind_SPAN_KIND_CONSUMER.String(),
	}
)

type Exporter interface {
	spanstore.Writer
	Start()
	Stop() error
}

type exporter struct {
	server   *http.Server
	logger   *zap.Logger
	services []string
}

func (e *exporter) Start() {
	e.logger.Info("Exporter server started", zap.String("address", e.server.Addr))

	if err := e.server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			e.logger.Error("Exporter server died unexpected", zap.Error(err))
		}
	}
}

func (e *exporter) Stop() error {
	e.logger.Debug("Start shutdown of the exporter server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return e.server.Shutdown(ctx)
}

func (e *exporter) WriteSpan(ctx context.Context, span *model.Span) error {
	spansReceivedTotalMetric.Inc()

	serviceName := span.Process.GetServiceName()
	operationName := span.GetOperationName()

	if e.services != nil && !slices.Contains(e.services, serviceName) {
		return nil
	}

	jaegerSpanKind, _ := span.GetSpanKind()
	otelSpanKind := metrics.SpanKind_SPAN_KIND_UNSPECIFIED.String()
	if val, ok := jaegerToOtelSpanKind[jaegerSpanKind.String()]; ok {
		otelSpanKind = val
	}

	statusCode := ""
	for _, tag := range span.Tags {
		if tag.Key == "error" {
			statusCode = "STATUS_CODE_ERROR"
		}
	}

	callsTotalMetric.WithLabelValues(serviceName, operationName, otelSpanKind, statusCode).Inc()
	durationMillisecondsMetric.WithLabelValues(serviceName, operationName, otelSpanKind, statusCode).Observe(durationToMillis(span.Duration))

	return nil
}

func New(logger *zap.Logger, address, services string) (Exporter, error) {
	router := http.NewServeMux()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	router.Handle("/metrics", promhttp.Handler())

	var servicesSlice []string
	if len(services) > 0 {
		servicesSlice = strings.Split(services, ",")
	}

	return &exporter{
		server: &http.Server{
			Addr:    address,
			Handler: router,
		},
		logger:   logger,
		services: servicesSlice,
	}, nil
}
