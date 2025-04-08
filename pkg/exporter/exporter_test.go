package exporter

import (
	"context"
	"testing"
	"time"

	"github.com/jaegertracing/jaeger/model"
	"go.uber.org/zap/zaptest"
)

func TestExporter_WriteSpan(t *testing.T) {
	logger := zaptest.NewLogger(t)
	exporter, err := New(logger, ":0", "service1,service2")
	if err != nil {
		t.Fatalf("Failed to create exporter: %v", err)
	}

	span := &model.Span{
		Process: &model.Process{
			ServiceName: "service1",
		},
		OperationName: "test-operation",
		Duration:      time.Millisecond * 500,
		Tags: []model.KeyValue{
			{Key: "error", VType: model.ValueType_STRING, VStr: "true"},
		},
	}

	if err := exporter.WriteSpan(context.Background(), span); err != nil {
		t.Errorf("Failed to write span: %v", err)
	}
}
