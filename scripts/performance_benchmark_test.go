package main

import (
	"testing"
	"time"

	"github.com/opd-ai/whisp/platform/common"
)

func BenchmarkMetricsCollection(b *testing.B) {
	metrics := common.NewMetricsCollector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.IncrementCounter("test_counter")
		metrics.RecordTimer("test_timer", time.Duration(i)*time.Millisecond)
		metrics.SetMetric("test_metric", i)
	}
}

func BenchmarkSecurityMonitoring(b *testing.B) {
	monitor := common.NewSecurityMonitor()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		details := map[string]interface{}{
			"attempt": i,
			"source":  "test",
		}
		monitor.RecordEvent("test_event", "low", "Test security event", details)
	}
}

func BenchmarkInputValidation(b *testing.B) {
	validator := common.NewInputValidator()

	testToxID := "1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890AB"
	testMessage := "This is a test message for validation"
	testFileName := "test_file.txt"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateToxID(testToxID)
		validator.ValidateMessageContent(testMessage)
		validator.ValidateFileName(testFileName)
	}
}

func TestPerformanceMonitorIntegration(t *testing.T) {
	monitor := common.NewPerformanceMonitor()

	// Test basic functionality
	monitor.Metrics.IncrementCounter("startup_events")
	monitor.Metrics.RecordTimer("initialization", 100*time.Millisecond)
	monitor.Metrics.SetMetric("test_version", "1.0.0")

	monitor.Security.RecordEvent("test", "info", "Test event", nil)

	// Collect system metrics
	monitor.Metrics.CollectSystemMetrics()

	// Generate report
	report := monitor.Report()
	if len(report) == 0 {
		t.Error("Expected non-empty report")
	}

	// Verify metrics were collected
	if monitor.Metrics.GetCounter("startup_events") != 1 {
		t.Error("Expected counter to be 1")
	}

	if monitor.Metrics.GetMetric("test_version") != "1.0.0" {
		t.Error("Expected metric to be set")
	}

	events := monitor.Security.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 security event, got %d", len(events))
	}
}
