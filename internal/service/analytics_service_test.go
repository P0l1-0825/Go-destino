package service

import (
	"testing"
)

func TestAnalyticsService_GetSLOMetrics(t *testing.T) {
	svc := NewAnalyticsService(nil)

	metrics := svc.GetSLOMetrics()

	if len(metrics) == 0 {
		t.Fatal("expected SLO metrics, got none")
	}

	// Verify all required services are present
	requiredServices := map[string]bool{
		"auth-service":    false,
		"booking-service": false,
		"payment-service": false,
		"fleet-service":   false,
		"ai-service":      false,
		"kiosk-service":   false,
	}

	for _, m := range metrics {
		if _, ok := requiredServices[m.Service]; ok {
			requiredServices[m.Service] = true
		}

		// Verify valid ranges
		if m.UptimePercent < 0 || m.UptimePercent > 100 {
			t.Errorf("service %s: invalid uptime %.2f", m.Service, m.UptimePercent)
		}
		if m.P50LatencyMS <= 0 {
			t.Errorf("service %s: P50 should be positive, got %d", m.Service, m.P50LatencyMS)
		}
		if m.P99LatencyMS < m.P50LatencyMS {
			t.Errorf("service %s: P99 (%d) should be >= P50 (%d)", m.Service, m.P99LatencyMS, m.P50LatencyMS)
		}
		if m.ErrorRate < 0 || m.ErrorRate > 1 {
			t.Errorf("service %s: invalid error rate %.4f", m.Service, m.ErrorRate)
		}
		if m.ErrorBudget < 0 || m.ErrorBudget > 100 {
			t.Errorf("service %s: invalid error budget %.2f", m.Service, m.ErrorBudget)
		}
	}

	for svc, found := range requiredServices {
		if !found {
			t.Errorf("missing SLO metrics for service: %s", svc)
		}
	}
}

func TestAnalyticsService_SLOMetrics_Count(t *testing.T) {
	svc := NewAnalyticsService(nil)
	metrics := svc.GetSLOMetrics()

	if len(metrics) != 6 {
		t.Errorf("expected 6 SLO metrics, got %d", len(metrics))
	}
}
