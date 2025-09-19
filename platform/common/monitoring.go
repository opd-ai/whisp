package common

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// MetricsCollector collects performance and security metrics
type MetricsCollector struct {
	mu        sync.RWMutex
	startTime time.Time
	metrics   map[string]interface{}
	counters  map[string]int64
	timers    map[string][]time.Duration
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime: time.Now(),
		metrics:   make(map[string]interface{}),
		counters:  make(map[string]int64),
		timers:    make(map[string][]time.Duration),
	}
}

// IncrementCounter increments a named counter
func (m *MetricsCollector) IncrementCounter(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name]++
}

// GetCounter returns the value of a counter
func (m *MetricsCollector) GetCounter(name string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.counters[name]
}

// RecordTimer records a timing measurement
func (m *MetricsCollector) RecordTimer(name string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timers[name] = append(m.timers[name], duration)
}

// GetAverageTimer returns the average duration for a timer
func (m *MetricsCollector) GetAverageTimer(name string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	times := m.timers[name]
	if len(times) == 0 {
		return 0
	}

	var total time.Duration
	for _, t := range times {
		total += t
	}
	return total / time.Duration(len(times))
}

// SetMetric sets a custom metric value
func (m *MetricsCollector) SetMetric(name string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics[name] = value
}

// GetMetric retrieves a metric value
func (m *MetricsCollector) GetMetric(name string) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metrics[name]
}

// CollectSystemMetrics collects current system performance metrics
func (m *MetricsCollector) CollectSystemMetrics() {
	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)

	m.SetMetric("memory_allocated", memStats.Alloc)
	m.SetMetric("memory_system", memStats.Sys)
	m.SetMetric("memory_gc_cycles", memStats.NumGC)
	m.SetMetric("goroutines", runtime.NumGoroutine())
	m.SetMetric("uptime", time.Since(m.startTime))
}

// GetSummary returns a summary of all metrics
func (m *MetricsCollector) GetSummary() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := make(map[string]interface{})

	// Copy metrics
	for k, v := range m.metrics {
		summary[k] = v
	}

	// Copy counters
	for k, v := range m.counters {
		summary[k] = v
	}

	// Calculate timer averages
	for k := range m.timers {
		summary[k+"_avg"] = m.GetAverageTimer(k)
	}

	return summary
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	Severity    string                 `json:"severity"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// SecurityMonitor monitors security-related events
type SecurityMonitor struct {
	mu     sync.RWMutex
	events []SecurityEvent
}

// NewSecurityMonitor creates a new security monitor
func NewSecurityMonitor() *SecurityMonitor {
	return &SecurityMonitor{
		events: make([]SecurityEvent, 0),
	}
}

// RecordEvent records a security event
func (sm *SecurityMonitor) RecordEvent(eventType, severity, description string, details map[string]interface{}) {
	event := SecurityEvent{
		Timestamp:   time.Now(),
		EventType:   eventType,
		Severity:    severity,
		Description: description,
		Details:     details,
	}

	sm.mu.Lock()
	sm.events = append(sm.events, event)
	sm.mu.Unlock()
}

// GetEvents returns all recorded security events
func (sm *SecurityMonitor) GetEvents() []SecurityEvent {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	events := make([]SecurityEvent, len(sm.events))
	copy(events, sm.events)
	return events
}

// GetEventsBySeverity returns events filtered by severity
func (sm *SecurityMonitor) GetEventsBySeverity(severity string) []SecurityEvent {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var filtered []SecurityEvent
	for _, event := range sm.events {
		if event.Severity == severity {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// PerformanceMonitor combines metrics collection and security monitoring
type PerformanceMonitor struct {
	Metrics  *MetricsCollector
	Security *SecurityMonitor
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		Metrics:  NewMetricsCollector(),
		Security: NewSecurityMonitor(),
	}
}

// Start starts the monitoring system
func (pm *PerformanceMonitor) Start() {
	// Start background collection of system metrics
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			pm.Metrics.CollectSystemMetrics()
		}
	}()
}

// Report generates a comprehensive performance and security report
func (pm *PerformanceMonitor) Report() string {
	var report strings.Builder

	report.WriteString("=== Performance & Security Report ===\n\n")

	// Performance metrics
	report.WriteString("Performance Metrics:\n")
	metrics := pm.Metrics.GetSummary()
	for key, value := range metrics {
		report.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
	}

	report.WriteString("\nSecurity Events:\n")
	events := pm.Security.GetEvents()
	if len(events) == 0 {
		report.WriteString("  No security events recorded\n")
	} else {
		for _, event := range events {
			report.WriteString(fmt.Sprintf("  [%s] %s: %s\n",
				event.Severity, event.EventType, event.Description))
		}
	}

	return report.String()
}
