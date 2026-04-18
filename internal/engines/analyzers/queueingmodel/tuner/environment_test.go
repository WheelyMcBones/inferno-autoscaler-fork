package tuner

import (
	"math"
	"testing"
)

func TestEnvironment_Valid_AllPositive(t *testing.T) {
	env := &Environment{
		Lambda:        10.0,
		AvgInputToks:  200.0,
		AvgOutputToks: 100.0,
		MaxBatchSize:  256,
		AvgTTFT:       500.0,
		AvgITL:        50.0,
	}
	if !env.Valid() {
		t.Error("expected Valid() = true for fully-populated environment")
	}
}

func TestEnvironment_Valid_ZeroLambda(t *testing.T) {
	env := &Environment{
		Lambda:        0,
		AvgInputToks:  200.0,
		AvgOutputToks: 100.0,
		MaxBatchSize:  256,
		AvgTTFT:       500.0,
		AvgITL:        50.0,
	}
	if env.Valid() {
		t.Error("expected Valid() = false for Lambda=0")
	}
}

func TestEnvironment_Valid_NegativeLambda(t *testing.T) {
	env := &Environment{
		Lambda:        -1.0,
		AvgInputToks:  200.0,
		AvgOutputToks: 100.0,
		MaxBatchSize:  256,
		AvgTTFT:       500.0,
		AvgITL:        50.0,
	}
	if env.Valid() {
		t.Error("expected Valid() = false for negative Lambda")
	}
}

func TestEnvironment_Valid_InfiniteLambda(t *testing.T) {
	env := &Environment{
		Lambda:        float32(math.Inf(1)),
		AvgInputToks:  200.0,
		AvgOutputToks: 100.0,
		MaxBatchSize:  256,
		AvgTTFT:       500.0,
		AvgITL:        50.0,
	}
	if env.Valid() {
		t.Error("expected Valid() = false for Lambda=+Inf")
	}
}

func TestEnvironment_Valid_NaNLambda(t *testing.T) {
	env := &Environment{
		Lambda:        float32(math.NaN()),
		AvgInputToks:  200.0,
		AvgOutputToks: 100.0,
		MaxBatchSize:  256,
		AvgTTFT:       500.0,
		AvgITL:        50.0,
	}
	if env.Valid() {
		t.Error("expected Valid() = false for Lambda=NaN")
	}
}

func TestEnvironment_Valid_ZeroAvgInputToks(t *testing.T) {
	env := &Environment{
		Lambda:        10.0,
		AvgInputToks:  0,
		AvgOutputToks: 100.0,
		MaxBatchSize:  256,
		AvgTTFT:       500.0,
		AvgITL:        50.0,
	}
	if env.Valid() {
		t.Error("expected Valid() = false for AvgInputToks=0")
	}
}

func TestEnvironment_Valid_ZeroAvgOutputToks(t *testing.T) {
	env := &Environment{
		Lambda:        10.0,
		AvgInputToks:  200.0,
		AvgOutputToks: 0,
		MaxBatchSize:  256,
		AvgTTFT:       500.0,
		AvgITL:        50.0,
	}
	if env.Valid() {
		t.Error("expected Valid() = false for AvgOutputToks=0")
	}
}

func TestEnvironment_Valid_ZeroMaxBatchSize(t *testing.T) {
	env := &Environment{
		Lambda:        10.0,
		AvgInputToks:  200.0,
		AvgOutputToks: 100.0,
		MaxBatchSize:  0,
		AvgTTFT:       500.0,
		AvgITL:        50.0,
	}
	if env.Valid() {
		t.Error("expected Valid() = false for MaxBatchSize=0")
	}
}

func TestEnvironment_Valid_ZeroAvgTTFT(t *testing.T) {
	env := &Environment{
		Lambda:        10.0,
		AvgInputToks:  200.0,
		AvgOutputToks: 100.0,
		MaxBatchSize:  256,
		AvgTTFT:       0,
		AvgITL:        50.0,
	}
	if env.Valid() {
		t.Error("expected Valid() = false for AvgTTFT=0")
	}
}

func TestEnvironment_Valid_ZeroAvgITL(t *testing.T) {
	env := &Environment{
		Lambda:        10.0,
		AvgInputToks:  200.0,
		AvgOutputToks: 100.0,
		MaxBatchSize:  256,
		AvgTTFT:       500.0,
		AvgITL:        0,
	}
	if env.Valid() {
		t.Error("expected Valid() = false for AvgITL=0")
	}
}

func TestEnvironment_Valid_NilEnvironment(t *testing.T) {
	var env *Environment
	if env.Valid() {
		t.Error("expected Valid() = false for nil environment")
	}
}

func TestEnvironment_GetObservations(t *testing.T) {
	env := &Environment{
		Lambda:        10.0,
		AvgInputToks:  200.0,
		AvgOutputToks: 100.0,
		MaxBatchSize:  256,
		AvgTTFT:       500.0,
		AvgITL:        50.0,
	}

	obs := env.GetObservations()
	if obs == nil {
		t.Fatal("expected non-nil observations vector")
	}
	if obs.Len() != 2 {
		t.Fatalf("expected observation vector of length 2, got %d", obs.Len())
	}
	if obs.AtVec(0) != float64(env.AvgTTFT) {
		t.Errorf("obs[0] = %f, want %f (AvgTTFT)", obs.AtVec(0), float64(env.AvgTTFT))
	}
	if obs.AtVec(1) != float64(env.AvgITL) {
		t.Errorf("obs[1] = %f, want %f (AvgITL)", obs.AtVec(1), float64(env.AvgITL))
	}
}

func TestEnvironment_GetObservations_DifferentValues(t *testing.T) {
	tests := []struct {
		name    string
		avgTTFT float32
		avgITL  float32
	}{
		{"typical", 500.0, 50.0},
		{"low latency", 10.0, 1.0},
		{"high latency", 9000.0, 450.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &Environment{
				Lambda:        5.0,
				AvgInputToks:  100.0,
				AvgOutputToks: 50.0,
				MaxBatchSize:  128,
				AvgTTFT:       tt.avgTTFT,
				AvgITL:        tt.avgITL,
			}
			obs := env.GetObservations()
			if obs.AtVec(0) != float64(tt.avgTTFT) {
				t.Errorf("obs[0] = %f, want %f", obs.AtVec(0), float64(tt.avgTTFT))
			}
			if obs.AtVec(1) != float64(tt.avgITL) {
				t.Errorf("obs[1] = %f, want %f", obs.AtVec(1), float64(tt.avgITL))
			}
		})
	}
}
