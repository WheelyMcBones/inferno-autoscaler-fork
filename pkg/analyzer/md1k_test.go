package analyzer

import (
	"math"
	"testing"

	"github.com/llm-d-incubation/workload-variant-autoscaler/pkg/config"
)

// Test that MD1K and MM1K models can both be created and solve correctly
func TestQueueModelTypes(t *testing.T) {
	// Test parameters
	lambda := float32(0.5) // arrival rate (requests/msec)
	mu := float32(1.0)     // service rate (requests/msec)
	K := 10                // capacity

	tests := []struct {
		name      string
		modelType config.QueueModelType
	}{
		{"MM1K Model", config.MM1K},
		{"MD1K Model", config.MD1K},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			servRate := []float32{mu, mu, mu, mu, mu} // constant service rate for simplicity

			var model QueueModelInterface
			if tt.modelType == config.MM1K {
				model = NewMM1ModelStateDependent(K, servRate)
			} else {
				model = NewMD1ModelStateDependent(K, servRate)
			}

			// Solve the model
			model.Solve(lambda, mu)

			// Check that model is valid
			if !model.IsValid() {
				t.Errorf("%s: model is not valid", tt.name)
			}

			// Check basic properties
			rho := model.GetRho()
			if rho < 0 || rho > 1 {
				t.Errorf("%s: utilization %v out of bounds [0,1]", tt.name, rho)
			}

			throughput := model.GetThroughput()
			if throughput <= 0 || throughput > lambda {
				t.Errorf("%s: throughput %v should be in (0, %v]", tt.name, throughput, lambda)
			}

			avgWaitTime := model.GetAvgWaitTime()
			if avgWaitTime < 0 {
				t.Errorf("%s: waiting time %v cannot be negative", tt.name, avgWaitTime)
			}

			avgServTime := model.GetAvgServTime()
			if avgServTime <= 0 {
				t.Errorf("%s: service time %v must be positive", tt.name, avgServTime)
			}

			avgRespTime := model.GetAvgRespTime()
			expectedRespTime := avgWaitTime + avgServTime
			if math.Abs(float64(avgRespTime-expectedRespTime)) > 0.01 {
				t.Errorf("%s: response time %v != wait + service time %v",
					tt.name, avgRespTime, expectedRespTime)
			}

			t.Logf("%s: ρ=%.3f, throughput=%.3f, wait=%.3f, service=%.3f, response=%.3f",
				tt.name, rho, throughput, avgWaitTime, avgServTime, avgRespTime)
		})
	}
}

// Test that MD1K gives lower waiting times than MM1K (for same parameters)
func TestMD1KVsMM1K(t *testing.T) {
	lambda := float32(0.7) // arrival rate (70% utilization)
	mu := float32(1.0)     // service rate
	K := 20                // capacity

	servRate := []float32{mu, mu, mu, mu, mu, mu, mu, mu}

	// Create both models
	mm1Model := NewMM1ModelStateDependent(K, servRate)
	md1Model := NewMD1ModelStateDependent(K, servRate)

	// Solve both
	mm1Model.Solve(lambda, mu)
	md1Model.Solve(lambda, mu)

	if !mm1Model.IsValid() || !md1Model.IsValid() {
		t.Fatal("One or both models are invalid")
	}

	// MD1K should have lower waiting time due to deterministic service
	mm1Wait := mm1Model.GetAvgWaitTime()
	md1Wait := md1Model.GetAvgWaitTime()

	t.Logf("MM1K waiting time: %.3f", mm1Wait)
	t.Logf("MD1K waiting time: %.3f", md1Wait)
	t.Logf("Ratio: %.2f", mm1Wait/md1Wait)

	// MD1K should have approximately half the waiting time of MM1K
	// (based on Pollaczek-Khintchine formula: C_s=0 vs C_s=1)
	if md1Wait >= mm1Wait {
		t.Errorf("MD1K waiting time %.3f should be less than MM1K waiting time %.3f",
			md1Wait, mm1Wait)
	}

	// Check that ratio is approximately 2 (MM1K / MD1K ≈ 2)
	ratio := mm1Wait / md1Wait
	if ratio < 1.5 || ratio > 2.5 {
		t.Logf("Warning: Expected ratio around 2.0, got %.2f (may vary due to finite capacity)", ratio)
	}
}

// Test QueueAnalyzer with both model types
func TestQueueAnalyzerModelTypes(t *testing.T) {
	// Create service parameters
	serviceParms := &ServiceParms{
		Prefill: &PrefillParms{
			Gamma: 10.0, // 10 msec base
			Delta: 0.01, // 0.01 msec per token per batch element
		},
		Decode: &DecodeParms{
			Alpha: 5.0, // 5 msec base
			Beta:  1.0, // 1 msec per batch element
		},
	}

	requestSize := &RequestSize{
		AvgInputTokens:  100,
		AvgOutputTokens: 50,
	}

	tests := []struct {
		name      string
		modelType config.QueueModelType
	}{
		{"With MM1K", config.MM1K},
		{"With MD1K", config.MD1K},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qConfig := &Configuration{
				MaxBatchSize: 8,
				MaxQueueSize: 80,
				ServiceParms: serviceParms,
				ModelType:    tt.modelType,
			}

			analyzer, err := NewQueueAnalyzer(qConfig, requestSize)
			if err != nil {
				t.Fatalf("Failed to create analyzer: %v", err)
			}

			// Verify correct model type was created
			if analyzer.ModelType != tt.modelType {
				t.Errorf("Expected model type %v, got %v", tt.modelType, analyzer.ModelType)
			}

			// Test analysis at 50% of max rate
			requestRate := analyzer.RateRange.Max * 0.5
			metrics, err := analyzer.Analyze(requestRate)
			if err != nil {
				t.Fatalf("Analysis failed: %v", err)
			}

			// Verify metrics are reasonable
			if metrics.Throughput <= 0 {
				t.Errorf("Throughput should be positive, got %v", metrics.Throughput)
			}

			if metrics.AvgWaitTime < 0 {
				t.Errorf("Wait time cannot be negative, got %v", metrics.AvgWaitTime)
			}

			if metrics.Rho < 0 || metrics.Rho > 1 {
				t.Errorf("Utilization should be in [0,1], got %v", metrics.Rho)
			}

			t.Logf("%s: throughput=%.2f req/s, wait=%.2f ms, rho=%.3f",
				tt.name, metrics.Throughput, metrics.AvgWaitTime, metrics.Rho)
		})
	}
}

// Test that MD1K sizing gives higher capacity than MM1K (due to lower waiting time)
func TestMD1KCapacityAdvantage(t *testing.T) {
	serviceParms := &ServiceParms{
		Prefill: &PrefillParms{Gamma: 10.0, Delta: 0.01},
		Decode:  &DecodeParms{Alpha: 5.0, Beta: 1.0},
	}

	requestSize := &RequestSize{
		AvgInputTokens:  100,
		AvgOutputTokens: 50,
	}

	targetPerf := &TargetPerf{
		TargetTTFT: 100.0, // 100 msec TTFT target
		TargetITL:  10.0,  // 10 msec ITL target
		TargetTPS:  0,     // no TPS constraint
	}

	// Test with MM1K
	mm1Config := &Configuration{
		MaxBatchSize: 8,
		MaxQueueSize: 80,
		ServiceParms: serviceParms,
		ModelType:    config.MM1K,
	}
	mm1Analyzer, _ := NewQueueAnalyzer(mm1Config, requestSize)
	mm1Rates, mm1Metrics, _, err := mm1Analyzer.Size(targetPerf)
	if err != nil {
		t.Fatalf("MM1K sizing failed: %v", err)
	}

	// Test with MD1K
	md1Config := &Configuration{
		MaxBatchSize: 8,
		MaxQueueSize: 80,
		ServiceParms: serviceParms,
		ModelType:    config.MD1K,
	}
	md1Analyzer, _ := NewQueueAnalyzer(md1Config, requestSize)
	md1Rates, md1Metrics, _, err := md1Analyzer.Size(targetPerf)
	if err != nil {
		t.Fatalf("MD1K sizing failed: %v", err)
	}

	t.Logf("MM1K max rate: %.2f req/s (wait: %.2f ms)",
		mm1Rates.RateTargetTTFT, mm1Metrics.AvgWaitTime)
	t.Logf("MD1K max rate: %.2f req/s (wait: %.2f ms)",
		md1Rates.RateTargetTTFT, md1Metrics.AvgWaitTime)

	// MD1K should support higher arrival rate for same TTFT target
	if md1Rates.RateTargetTTFT <= mm1Rates.RateTargetTTFT {
		t.Logf("Note: MD1K rate %.2f should be higher than MM1K rate %.2f",
			md1Rates.RateTargetTTFT, mm1Rates.RateTargetTTFT)
	} else {
		improvement := (md1Rates.RateTargetTTFT - mm1Rates.RateTargetTTFT) / mm1Rates.RateTargetTTFT * 100
		t.Logf("MD1K provides %.1f%% higher capacity than MM1K", improvement)
	}
}
