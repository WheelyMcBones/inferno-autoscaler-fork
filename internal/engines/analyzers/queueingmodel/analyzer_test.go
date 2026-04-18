package queueingmodel

import (
	"context"
	"testing"

	"github.com/llm-d/llm-d-workload-variant-autoscaler/internal/engines/analyzers/queueingmodel/tuner"
	"github.com/llm-d/llm-d-workload-variant-autoscaler/internal/interfaces"
	"github.com/llm-d/llm-d-workload-variant-autoscaler/internal/logging"
	"github.com/llm-d/llm-d-workload-variant-autoscaler/pkg/analyzer"
	"gonum.org/v1/gonum/mat"
)

func newTestContext() context.Context {
	return logging.NewTestLoggerIntoContext(context.Background())
}

func TestNewQueueingModelAnalyzer(t *testing.T) {
	a := NewQueueingModelAnalyzer()
	if a == nil {
		t.Fatal("expected non-nil analyzer")
	}
	if a.modelsParameterStore == nil {
		t.Fatal("expected non-nil modelsParameterStore")
	}
	if len(a.modelsParameterStore) != 0 {
		t.Errorf("expected empty modelsParameterStore, got len=%d", len(a.modelsParameterStore))
	}
}

func TestQueueingModelAnalyzer_Name(t *testing.T) {
	a := NewQueueingModelAnalyzer()
	if a.Name() != interfaces.QueueingModelAnalyzerName {
		t.Errorf("Name() = %q, want %q", a.Name(), interfaces.QueueingModelAnalyzerName)
	}
}

func TestQueueingModelAnalyzer_Update_AddsModels(t *testing.T) {
	a := NewQueueingModelAnalyzer()

	models := map[string]bool{
		"ns1/model-a": true,
		"ns2/model-b": true,
	}
	a.Update(models)

	if len(a.modelsParameterStore) != 2 {
		t.Errorf("expected 2 models in store, got %d", len(a.modelsParameterStore))
	}
	if _, ok := a.modelsParameterStore["ns1/model-a"]; !ok {
		t.Error("expected ns1/model-a to be in store")
	}
	if _, ok := a.modelsParameterStore["ns2/model-b"]; !ok {
		t.Error("expected ns2/model-b to be in store")
	}
}

func TestQueueingModelAnalyzer_Update_RemovesModels(t *testing.T) {
	a := NewQueueingModelAnalyzer()
	a.Update(map[string]bool{"ns1/model-a": true, "ns2/model-b": true})

	// Keep only one
	a.Update(map[string]bool{"ns1/model-a": true})

	if len(a.modelsParameterStore) != 1 {
		t.Errorf("expected 1 model after removal, got %d", len(a.modelsParameterStore))
	}
	if _, ok := a.modelsParameterStore["ns2/model-b"]; ok {
		t.Error("expected ns2/model-b to be removed from store")
	}
}

func TestQueueingModelAnalyzer_Update_EmptySet(t *testing.T) {
	a := NewQueueingModelAnalyzer()
	a.Update(map[string]bool{"ns1/model-a": true})
	a.Update(map[string]bool{}) // clear all

	if len(a.modelsParameterStore) != 0 {
		t.Errorf("expected empty store after clearing all models, got %d", len(a.modelsParameterStore))
	}
}

func TestQueueingModelAnalyzer_Update_PreservesExistingParams(t *testing.T) {
	a := NewQueueingModelAnalyzer()
	a.Update(map[string]bool{"ns/model": true})
	a.setParams("model", "ns", "variant-1", &LearnedParameters{Alpha: 5.0})

	// Call Update again with the same model — must NOT wipe existing parameters
	a.Update(map[string]bool{"ns/model": true})

	params := a.getParams("model", "ns", "variant-1")
	if params == nil {
		t.Fatal("expected parameters to be preserved after Update, got nil")
	}
	if params.Alpha != 5.0 {
		t.Errorf("Alpha = %f, want 5.0 after re-update", params.Alpha)
	}
}

func TestQueueingModelAnalyzer_GetParams_NonExistent(t *testing.T) {
	a := NewQueueingModelAnalyzer()

	// No model registered
	if p := a.getParams("missing-model", "ns", "variant-1"); p != nil {
		t.Errorf("expected nil for unregistered model, got %+v", p)
	}
}

func TestQueueingModelAnalyzer_GetParams_AfterUpdate_Missing(t *testing.T) {
	a := NewQueueingModelAnalyzer()
	a.Update(map[string]bool{"ns/model": true})

	// Model exists, but variant not yet stored
	if p := a.getParams("model", "ns", "no-such-variant"); p != nil {
		t.Errorf("expected nil for unregistered variant, got %+v", p)
	}
}

func TestQueueingModelAnalyzer_SetAndGetParams(t *testing.T) {
	a := NewQueueingModelAnalyzer()
	a.Update(map[string]bool{"ns/model-a": true})

	expected := &LearnedParameters{Alpha: 1.0, Beta: 2.0, Gamma: 3.0, NIS: 0.5}
	a.setParams("model-a", "ns", "variant-1", expected)

	got := a.getParams("model-a", "ns", "variant-1")
	if got == nil {
		t.Fatal("expected non-nil params after setParams")
	}
	if got.Alpha != expected.Alpha {
		t.Errorf("Alpha = %f, want %f", got.Alpha, expected.Alpha)
	}
	if got.Beta != expected.Beta {
		t.Errorf("Beta = %f, want %f", got.Beta, expected.Beta)
	}
	if got.Gamma != expected.Gamma {
		t.Errorf("Gamma = %f, want %f", got.Gamma, expected.Gamma)
	}
	if got.NIS != expected.NIS {
		t.Errorf("NIS = %f, want %f", got.NIS, expected.NIS)
	}
}

func TestQueueingModelAnalyzer_SetParams_Overrides(t *testing.T) {
	a := NewQueueingModelAnalyzer()
	a.Update(map[string]bool{"ns/model": true})

	a.setParams("model", "ns", "v1", &LearnedParameters{Alpha: 1.0})
	a.setParams("model", "ns", "v1", &LearnedParameters{Alpha: 9.0})

	got := a.getParams("model", "ns", "v1")
	if got.Alpha != 9.0 {
		t.Errorf("Alpha = %f, want 9.0 after override", got.Alpha)
	}
}

// ============ guessInitState ============

func TestGuessInitState_Valid(t *testing.T) {
	env := &tuner.Environment{
		Lambda:        10.0,
		AvgInputToks:  200.0,
		AvgOutputToks: 100.0,
		MaxBatchSize:  256,
		AvgTTFT:       500.0, // ms
		AvgITL:        50.0,  // ms
	}

	state, err := guessInitState(env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(state) != 3 {
		t.Fatalf("expected state vector of length 3, got %d", len(state))
	}

	alpha, beta, gamma := StateVectorToParams(state)
	if alpha <= 0 {
		t.Errorf("expected alpha > 0, got %f", alpha)
	}
	if beta <= 0 {
		t.Errorf("expected beta > 0, got %f", beta)
	}
	if gamma <= 0 {
		t.Errorf("expected gamma > 0, got %f", gamma)
	}
}

func TestGuessInitState_NilEnv(t *testing.T) {
	_, err := guessInitState(nil)
	if err == nil {
		t.Error("expected error for nil environment")
	}
}

func TestGuessInitState_InvalidEnv(t *testing.T) {
	env := &tuner.Environment{
		Lambda:        0, // invalid: Lambda must be > 0
		AvgInputToks:  100.0,
		AvgOutputToks: 50.0,
		MaxBatchSize:  256,
		AvgTTFT:       100.0,
		AvgITL:        10.0,
	}
	_, err := guessInitState(env)
	if err == nil {
		t.Error("expected error for invalid environment (Lambda=0)")
	}
}

func TestGuessInitState_AlphaBasedOnITL(t *testing.T) {
	// Verify that alpha = BaseFactor * ITL
	env := &tuner.Environment{
		Lambda:        10.0,
		AvgInputToks:  200.0,
		AvgOutputToks: 100.0,
		MaxBatchSize:  256,
		AvgTTFT:       500.0,
		AvgITL:        50.0,
	}

	state, err := guessInitState(env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	alpha, _, _ := StateVectorToParams(state)

	expectedAlpha := tuner.BaseFactor * float64(env.AvgITL)
	if alpha != expectedAlpha {
		t.Errorf("alpha = %f, want %f (BaseFactor*ITL)", alpha, expectedAlpha)
	}
}

func TestGuessInitState_RoundTrip_StateVector(t *testing.T) {
	env := &tuner.Environment{
		Lambda:        5.0,
		AvgInputToks:  150.0,
		AvgOutputToks: 80.0,
		MaxBatchSize:  128,
		AvgTTFT:       400.0,
		AvgITL:        40.0,
	}

	state, err := guessInitState(env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	alpha, beta, gamma := StateVectorToParams(state)
	rebuilt := ParamsToStateVector(alpha, beta, gamma)

	if len(rebuilt) != len(state) {
		t.Fatalf("round-trip vector length mismatch: got %d, want %d", len(rebuilt), len(state))
	}
	for i := range state {
		if rebuilt[i] != state[i] {
			t.Errorf("state[%d]: got %f, want %f", i, rebuilt[i], state[i])
		}
	}
}

// ============ aggregateCapacities ============

func TestAggregateCapacities_Empty(t *testing.T) {
	supply, demand := aggregateCapacities(nil)
	if supply != 0 || demand != 0 {
		t.Errorf("expected (0, 0) for nil input, got (%f, %f)", supply, demand)
	}
}

func TestAggregateCapacities_EmptySlice(t *testing.T) {
	supply, demand := aggregateCapacities([]interfaces.VariantCapacity{})
	if supply != 0 || demand != 0 {
		t.Errorf("expected (0, 0) for empty slice, got (%f, %f)", supply, demand)
	}
}

func TestAggregateCapacities_Single(t *testing.T) {
	caps := []interfaces.VariantCapacity{
		{TotalCapacity: 10.0, TotalDemand: 7.5},
	}
	supply, demand := aggregateCapacities(caps)
	if supply != 10.0 {
		t.Errorf("supply = %f, want 10.0", supply)
	}
	if demand != 7.5 {
		t.Errorf("demand = %f, want 7.5", demand)
	}
}

func TestAggregateCapacities_Multiple(t *testing.T) {
	caps := []interfaces.VariantCapacity{
		{TotalCapacity: 10.0, TotalDemand: 5.0},
		{TotalCapacity: 20.0, TotalDemand: 15.0},
		{TotalCapacity: 5.0, TotalDemand: 8.0},
	}
	supply, demand := aggregateCapacities(caps)
	if supply != 35.0 {
		t.Errorf("supply = %f, want 35.0", supply)
	}
	if demand != 28.0 {
		t.Errorf("demand = %f, want 28.0", demand)
	}
}

func TestAggregateCapacities_ZeroValues(t *testing.T) {
	caps := []interfaces.VariantCapacity{
		{TotalCapacity: 0.0, TotalDemand: 0.0},
		{TotalCapacity: 0.0, TotalDemand: 0.0},
	}
	supply, demand := aggregateCapacities(caps)
	if supply != 0 || demand != 0 {
		t.Errorf("expected (0, 0) for all-zero capacities, got (%f, %f)", supply, demand)
	}
}

// ============ groupMetricsByVariant ============

func TestGroupMetricsByVariant_Nil(t *testing.T) {
	got := groupMetricsByVariant(nil)
	if len(got) != 0 {
		t.Errorf("expected empty map for nil input, got len=%d", len(got))
	}
}

func TestGroupMetricsByVariant_Empty(t *testing.T) {
	got := groupMetricsByVariant([]interfaces.ReplicaMetrics{})
	if len(got) != 0 {
		t.Errorf("expected empty map for empty input, got len=%d", len(got))
	}
}

func TestGroupMetricsByVariant_SingleVariant(t *testing.T) {
	metrics := []interfaces.ReplicaMetrics{
		{PodName: "pod-1", VariantName: "v1"},
		{PodName: "pod-2", VariantName: "v1"},
	}
	got := groupMetricsByVariant(metrics)

	if len(got) != 1 {
		t.Fatalf("expected 1 group, got %d", len(got))
	}
	if len(got["v1"]) != 2 {
		t.Errorf("expected 2 replicas for v1, got %d", len(got["v1"]))
	}
}

func TestGroupMetricsByVariant_MultipleVariants(t *testing.T) {
	metrics := []interfaces.ReplicaMetrics{
		{PodName: "pod-1", VariantName: "v1"},
		{PodName: "pod-2", VariantName: "v2"},
		{PodName: "pod-3", VariantName: "v1"},
		{PodName: "pod-4", VariantName: "v3"},
	}
	got := groupMetricsByVariant(metrics)

	if len(got) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(got))
	}
	if len(got["v1"]) != 2 {
		t.Errorf("expected 2 replicas for v1, got %d", len(got["v1"]))
	}
	if len(got["v2"]) != 1 {
		t.Errorf("expected 1 replica for v2, got %d", len(got["v2"]))
	}
	if len(got["v3"]) != 1 {
		t.Errorf("expected 1 replica for v3, got %d", len(got["v3"]))
	}
}

func TestGroupMetricsByVariant_PodsBelongToCorrectVariant(t *testing.T) {
	metrics := []interfaces.ReplicaMetrics{
		{PodName: "pod-a", VariantName: "x"},
		{PodName: "pod-b", VariantName: "y"},
	}
	got := groupMetricsByVariant(metrics)

	if got["x"][0].PodName != "pod-a" {
		t.Errorf("expected pod-a in variant x, got %q", got["x"][0].PodName)
	}
	if got["y"][0].PodName != "pod-b" {
		t.Errorf("expected pod-b in variant y, got %q", got["y"][0].PodName)
	}
}

// ============ getVariantNames ============

func TestGetVariantNames_Nil(t *testing.T) {
	got := getVariantNames(nil)
	if len(got) != 0 {
		t.Errorf("expected empty for nil input, got %d names", len(got))
	}
}

func TestGetVariantNames_Empty(t *testing.T) {
	got := getVariantNames([]interfaces.ReplicaMetrics{})
	if len(got) != 0 {
		t.Errorf("expected empty for empty input, got %d names", len(got))
	}
}

func TestGetVariantNames_Deduplicates(t *testing.T) {
	metrics := []interfaces.ReplicaMetrics{
		{PodName: "pod-1", VariantName: "v1"},
		{PodName: "pod-2", VariantName: "v2"},
		{PodName: "pod-3", VariantName: "v1"}, // duplicate
	}
	got := getVariantNames(metrics)
	if len(got) != 2 {
		t.Errorf("expected 2 unique names, got %d: %v", len(got), got)
	}
}

func TestGetVariantNames_PreservesOrder(t *testing.T) {
	metrics := []interfaces.ReplicaMetrics{
		{PodName: "pod-1", VariantName: "v1"},
		{PodName: "pod-2", VariantName: "v2"},
		{PodName: "pod-3", VariantName: "v3"},
	}
	got := getVariantNames(metrics)

	if len(got) != 3 {
		t.Fatalf("expected 3 names, got %d", len(got))
	}
	want := []string{"v1", "v2", "v3"}
	for i, name := range want {
		if got[i] != name {
			t.Errorf("got[%d] = %q, want %q", i, got[i], name)
		}
	}
}

func TestGetVariantNames_DuplicatesPreserveFirstOccurrenceOrder(t *testing.T) {
	metrics := []interfaces.ReplicaMetrics{
		{VariantName: "v1"},
		{VariantName: "v3"},
		{VariantName: "v2"},
		{VariantName: "v1"}, // duplicate
		{VariantName: "v3"}, // duplicate
	}
	got := getVariantNames(metrics)

	if len(got) != 3 {
		t.Fatalf("expected 3 names, got %d: %v", len(got), got)
	}
	want := []string{"v1", "v3", "v2"}
	for i, name := range want {
		if got[i] != name {
			t.Errorf("got[%d] = %q, want %q", i, got[i], name)
		}
	}
}

// ============ aggregateWorkloadMetrics ============

func TestAggregateWorkloadMetrics_NilInput(t *testing.T) {
	wm := aggregateWorkloadMetrics(nil)
	if wm.busyPods != 0 {
		t.Errorf("expected 0 busy pods, got %d", wm.busyPods)
	}
	if wm.avgArrivalRate != 0 {
		t.Errorf("expected 0 arrival rate, got %f", wm.avgArrivalRate)
	}
}

func TestAggregateWorkloadMetrics_AllZeroArrivalRate(t *testing.T) {
	metrics := []interfaces.ReplicaMetrics{
		{PodName: "pod-1", ArrivalRate: 0},
		{PodName: "pod-2", ArrivalRate: 0},
	}
	wm := aggregateWorkloadMetrics(metrics)
	if wm.busyPods != 0 {
		t.Errorf("expected 0 busy pods for zero-rate pods, got %d", wm.busyPods)
	}
}

func TestAggregateWorkloadMetrics_SinglePod(t *testing.T) {
	metrics := []interfaces.ReplicaMetrics{
		{
			PodName:         "pod-1",
			ArrivalRate:     10.0,
			AvgInputTokens:  200.0,
			AvgOutputTokens: 100.0,
			AvgTTFT:         0.5,
			AvgITL:          0.05,
		},
	}
	wm := aggregateWorkloadMetrics(metrics)

	if wm.busyPods != 1 {
		t.Errorf("expected 1 busy pod, got %d", wm.busyPods)
	}
	if wm.avgArrivalRate != 10.0 {
		t.Errorf("avgArrivalRate = %f, want 10.0", wm.avgArrivalRate)
	}
	if wm.avgInputTokens != 200.0 {
		t.Errorf("avgInputTokens = %f, want 200.0", wm.avgInputTokens)
	}
	if wm.avgOutputTokens != 100.0 {
		t.Errorf("avgOutputTokens = %f, want 100.0", wm.avgOutputTokens)
	}
}

func TestAggregateWorkloadMetrics_EqualRatePods(t *testing.T) {
	// Two pods with identical stats and equal arrival rates
	metrics := []interfaces.ReplicaMetrics{
		{ArrivalRate: 10.0, AvgInputTokens: 200.0, AvgOutputTokens: 100.0, AvgTTFT: 0.5, AvgITL: 0.05},
		{ArrivalRate: 10.0, AvgInputTokens: 200.0, AvgOutputTokens: 100.0, AvgTTFT: 0.5, AvgITL: 0.05},
	}
	wm := aggregateWorkloadMetrics(metrics)

	if wm.busyPods != 2 {
		t.Errorf("expected 2 busy pods, got %d", wm.busyPods)
	}
	// avgArrivalRate = totalArrival / busyPods = 20 / 2 = 10
	if wm.avgArrivalRate != 10.0 {
		t.Errorf("avgArrivalRate = %f, want 10.0", wm.avgArrivalRate)
	}
	if wm.avgInputTokens != 200.0 {
		t.Errorf("avgInputTokens = %f, want 200.0", wm.avgInputTokens)
	}
}

func TestAggregateWorkloadMetrics_SkipsZeroArrivalRate(t *testing.T) {
	metrics := []interfaces.ReplicaMetrics{
		{ArrivalRate: 10.0, AvgInputTokens: 100.0, AvgOutputTokens: 50.0, AvgTTFT: 0.5, AvgITL: 0.05},
		{ArrivalRate: 0.0, AvgInputTokens: 999.0}, // should be skipped
	}
	wm := aggregateWorkloadMetrics(metrics)

	if wm.busyPods != 1 {
		t.Errorf("expected 1 busy pod (zero-rate pod skipped), got %d", wm.busyPods)
	}
	if wm.avgInputTokens != 100.0 {
		t.Errorf("avgInputTokens = %f, want 100.0 (zero-rate pod ignored)", wm.avgInputTokens)
	}
}

func TestAggregateWorkloadMetrics_WeightedByArrivalRate(t *testing.T) {
	// Pod-1 has 4x the arrival rate of pod-2, so its metrics dominate
	metrics := []interfaces.ReplicaMetrics{
		{ArrivalRate: 8.0, AvgInputTokens: 100.0, AvgOutputTokens: 50.0, AvgTTFT: 0.4, AvgITL: 0.04},
		{ArrivalRate: 2.0, AvgInputTokens: 300.0, AvgOutputTokens: 150.0, AvgTTFT: 1.0, AvgITL: 0.10},
	}
	wm := aggregateWorkloadMetrics(metrics)

	if wm.busyPods != 2 {
		t.Errorf("expected 2 busy pods, got %d", wm.busyPods)
	}
	// avgArrivalRate = (8+2) / 2 = 5.0
	if wm.avgArrivalRate != 5.0 {
		t.Errorf("avgArrivalRate = %f, want 5.0", wm.avgArrivalRate)
	}
	// avgInputTokens = (8*100 + 2*300) / (8+2) = (800+600)/10 = 140
	if wm.avgInputTokens != 140.0 {
		t.Errorf("avgInputTokens = %f, want 140.0", wm.avgInputTokens)
	}
}

// ============ buildEnvironmentsFromMetrics ============

func TestBuildEnvironmentsFromMetrics_NilMetrics(t *testing.T) {
	envs, err := buildEnvironmentsFromMetrics("v1", nil)
	if err == nil {
		t.Error("expected error for nil metrics")
	}
	if len(envs) != 0 {
		t.Errorf("expected 0 environments for nil input, got %d", len(envs))
	}
}

func TestBuildEnvironmentsFromMetrics_NoTrafficPods(t *testing.T) {
	metrics := []interfaces.ReplicaMetrics{
		{PodName: "pod-1", ArrivalRate: 0},
	}
	envs, err := buildEnvironmentsFromMetrics("v1", metrics)
	if err == nil {
		t.Error("expected error for pods with no traffic")
	}
	if len(envs) != 0 {
		t.Errorf("expected 0 environments for zero-rate pods, got %d", len(envs))
	}
}

func TestBuildEnvironmentsFromMetrics_ValidPod(t *testing.T) {
	metrics := []interfaces.ReplicaMetrics{
		{
			PodName:         "pod-1",
			VariantName:     "v1",
			ArrivalRate:     10.0,
			AvgInputTokens:  200.0,
			AvgOutputTokens: 100.0,
			MaxBatchSize:    256,
			AvgTTFT:         0.5,  // seconds -> 500 ms
			AvgITL:          0.05, // seconds -> 50 ms
		},
	}
	envs, err := buildEnvironmentsFromMetrics("v1", metrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(envs) != 1 {
		t.Fatalf("expected 1 environment, got %d", len(envs))
	}

	env := envs[0]
	// Lambda = ArrivalRate * 60 (convert req/sec to req/min)
	if env.Lambda != float32(10.0*60) {
		t.Errorf("Lambda = %f, want %f", env.Lambda, float32(10.0*60))
	}
	// AvgTTFT = 0.5 * 1000 = 500 ms
	if env.AvgTTFT != float32(500.0) {
		t.Errorf("AvgTTFT = %f, want 500.0 (ms)", env.AvgTTFT)
	}
	// AvgITL = 0.05 * 1000 = 50 ms
	if env.AvgITL != float32(50.0) {
		t.Errorf("AvgITL = %f, want 50.0 (ms)", env.AvgITL)
	}
	if env.MaxBatchSize != 256 {
		t.Errorf("MaxBatchSize = %d, want 256", env.MaxBatchSize)
	}
}

func TestBuildEnvironmentsFromMetrics_UsesDefaultMaxBatchSize(t *testing.T) {
	metrics := []interfaces.ReplicaMetrics{
		{
			PodName:         "pod-1",
			ArrivalRate:     5.0,
			AvgInputTokens:  100.0,
			AvgOutputTokens: 50.0,
			MaxBatchSize:    0, // not set
			AvgTTFT:         0.5,
			AvgITL:          0.05,
		},
	}
	envs, err := buildEnvironmentsFromMetrics("v1", metrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(envs) != 1 {
		t.Fatalf("expected 1 environment, got %d", len(envs))
	}
	if envs[0].MaxBatchSize != DefaultMaxBatchSize {
		t.Errorf("MaxBatchSize = %d, want DefaultMaxBatchSize=%d", envs[0].MaxBatchSize, DefaultMaxBatchSize)
	}
}

func TestBuildEnvironmentsFromMetrics_MultiplePodsWithTraffic(t *testing.T) {
	metrics := []interfaces.ReplicaMetrics{
		{PodName: "pod-1", ArrivalRate: 5.0, AvgInputTokens: 100.0, AvgOutputTokens: 50.0, AvgTTFT: 0.5, AvgITL: 0.05},
		{PodName: "pod-2", ArrivalRate: 0.0}, // no traffic, skipped
		{PodName: "pod-3", ArrivalRate: 8.0, AvgInputTokens: 200.0, AvgOutputTokens: 100.0, AvgTTFT: 0.3, AvgITL: 0.03},
	}
	envs, err := buildEnvironmentsFromMetrics("v1", metrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only 2 pods have traffic (pod-2 skipped due to zero arrival rate)
	if len(envs) != 2 {
		t.Errorf("expected 2 environments (one per trafficked pod), got %d", len(envs))
	}
}

// ============ fallbackSLOFromObservations ============

func TestFallbackSLOFromObservations_ZeroMetrics(t *testing.T) {
	ctx := newTestContext()
	wm := &workloadMetrics{avgTTFT: 0, avgITL: 0, busyPods: 1}
	slo := fallbackSLOFromObservations(ctx, wm)
	if slo != nil {
		t.Errorf("expected nil SLO for zero TTFT/ITL, got %+v", slo)
	}
}

func TestFallbackSLOFromObservations_NegativeTTFT(t *testing.T) {
	ctx := newTestContext()
	wm := &workloadMetrics{avgTTFT: -0.5, avgITL: 0.05, busyPods: 1}
	slo := fallbackSLOFromObservations(ctx, wm)
	if slo != nil {
		t.Errorf("expected nil SLO for negative TTFT, got %+v", slo)
	}
}

func TestFallbackSLOFromObservations_ValidMetrics(t *testing.T) {
	ctx := newTestContext()
	// avgTTFT=0.5s, avgITL=0.05s (seconds)
	wm := &workloadMetrics{avgTTFT: 0.5, avgITL: 0.05, busyPods: 1}
	slo := fallbackSLOFromObservations(ctx, wm)
	if slo == nil {
		t.Fatal("expected non-nil SLO for valid metrics")
	}

	// ttft = min(0.5 * 1000 * 1.5, 10000) = min(750, 10000) = 750
	// itl  = min(0.05 * 1000 * 1.5, 500) = min(75, 500) = 75
	if slo.TargetTTFT != float32(750.0) {
		t.Errorf("TargetTTFT = %f, want 750.0", slo.TargetTTFT)
	}
	if slo.TargetITL != float32(75.0) {
		t.Errorf("TargetITL = %f, want 75.0", slo.TargetITL)
	}
}

func TestFallbackSLOFromObservations_CapsAtMaximum(t *testing.T) {
	ctx := newTestContext()
	// Very high TTFT (100s) and ITL (10s) should be capped at defaults
	wm := &workloadMetrics{avgTTFT: 100.0, avgITL: 10.0, busyPods: 1}
	slo := fallbackSLOFromObservations(ctx, wm)
	if slo == nil {
		t.Fatal("expected non-nil SLO")
	}

	if float64(slo.TargetTTFT) > DefaultMaxFallbackTTFT {
		t.Errorf("TargetTTFT %f exceeds cap %f", slo.TargetTTFT, DefaultMaxFallbackTTFT)
	}
	if float64(slo.TargetITL) > DefaultMaxFallbackITL {
		t.Errorf("TargetITL %f exceeds cap %f", slo.TargetITL, DefaultMaxFallbackITL)
	}
}

func TestFallbackSLOFromObservations_HeadroomApplied(t *testing.T) {
	ctx := newTestContext()
	// avgTTFT=0.1s, avgITL=0.01s -> both safely below caps
	wm := &workloadMetrics{avgTTFT: 0.1, avgITL: 0.01, busyPods: 1}
	slo := fallbackSLOFromObservations(ctx, wm)
	if slo == nil {
		t.Fatal("expected non-nil SLO")
	}

	expectedTTFT := float32(0.1 * 1000.0 * DefaultFallbackHeadroom) // 150 ms
	expectedITL := float32(0.01 * 1000.0 * DefaultFallbackHeadroom) // 15 ms
	if slo.TargetTTFT != expectedTTFT {
		t.Errorf("TargetTTFT = %f, want %f", slo.TargetTTFT, expectedTTFT)
	}
	if slo.TargetITL != expectedITL {
		t.Errorf("TargetITL = %f, want %f", slo.TargetITL, expectedITL)
	}
}

// ============ storeParametersFromResults ============

func TestStoreParametersFromResults(t *testing.T) {
	a := NewQueueingModelAnalyzer()
	a.Update(map[string]bool{"ns/model": true})

	results := &tuner.TunedResults{
		ServiceParms: &analyzer.ServiceParms{
			Alpha: 5.0,
			Beta:  0.05,
			Gamma: 0.001,
		},
		Covariance: mat.NewDense(3, 3, []float64{
			1, 0, 0,
			0, 1, 0,
			0, 0, 1,
		}),
		NIS: 1.5,
	}

	a.storeParametersFromResults("ns", "model", "variant-1", results)

	params := a.getParams("model", "ns", "variant-1")
	if params == nil {
		t.Fatal("expected non-nil params after storeParametersFromResults")
	}
	if params.Alpha != 5.0 {
		t.Errorf("Alpha = %f, want 5.0", params.Alpha)
	}
	if params.Beta != 0.05 {
		t.Errorf("Beta = %f, want 0.05", params.Beta)
	}
	if params.Gamma != 0.001 {
		t.Errorf("Gamma = %f, want 0.001", params.Gamma)
	}
	if params.NIS != 1.5 {
		t.Errorf("NIS = %f, want 1.5", params.NIS)
	}
	if params.LastUpdated.IsZero() {
		t.Error("expected LastUpdated to be set after storing results")
	}
}

func TestStoreParametersFromResults_CovarianceExtracted(t *testing.T) {
	a := NewQueueingModelAnalyzer()
	a.Update(map[string]bool{"ns/model": true})

	cov := mat.NewDense(3, 3, []float64{
		2, 0, 0,
		0, 3, 0,
		0, 0, 4,
	})
	results := &tuner.TunedResults{
		ServiceParms: &analyzer.ServiceParms{Alpha: 1.0, Beta: 0.01, Gamma: 0.0001},
		Covariance:   cov,
		NIS:          0.5,
	}

	a.storeParametersFromResults("ns", "model", "v1", results)

	params := a.getParams("model", "ns", "v1")
	if params == nil {
		t.Fatal("expected non-nil params")
	}
	if len(params.Covariance) != 3 {
		t.Fatalf("expected 3x3 covariance, got %d rows", len(params.Covariance))
	}
	if params.Covariance[0][0] != 2.0 {
		t.Errorf("Covariance[0][0] = %f, want 2.0", params.Covariance[0][0])
	}
	if params.Covariance[1][1] != 3.0 {
		t.Errorf("Covariance[1][1] = %f, want 3.0", params.Covariance[1][1])
	}
	if params.Covariance[2][2] != 4.0 {
		t.Errorf("Covariance[2][2] = %f, want 4.0", params.Covariance[2][2])
	}
}
