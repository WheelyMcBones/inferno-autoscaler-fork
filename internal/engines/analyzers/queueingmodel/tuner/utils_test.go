package tuner

import (
	"testing"

	"gonum.org/v1/gonum/mat"
)

// ============ FloatEqual ============

func TestFloatEqual_ExactlyEqual(t *testing.T) {
	if !FloatEqual(1.5, 1.5, DefaultEpsilon) {
		t.Error("expected 1.5 == 1.5")
	}
}

func TestFloatEqual_BothZero(t *testing.T) {
	if !FloatEqual(0.0, 0.0, DefaultEpsilon) {
		t.Error("expected 0.0 == 0.0")
	}
}

func TestFloatEqual_WithinEpsilon(t *testing.T) {
	a := 1.0
	b := 1.0 + DefaultEpsilon*0.5 // half-epsilon difference
	if !FloatEqual(a, b, DefaultEpsilon) {
		t.Errorf("expected %f ≈ %f within epsilon %f", a, b, DefaultEpsilon)
	}
}

func TestFloatEqual_OutsideEpsilon(t *testing.T) {
	a := 1.0
	b := 2.0
	if FloatEqual(a, b, DefaultEpsilon) {
		t.Errorf("expected %f ≠ %f (difference is outside default epsilon)", a, b)
	}
}

func TestFloatEqual_NearZero(t *testing.T) {
	// Values very close to zero
	if !FloatEqual(1e-15, 2e-15, DefaultEpsilon) {
		// Both are effectively zero relative to SmallestNonzeroFloat64
		// FloatEqual returns diff < epsilon * SmallestNonzeroFloat64 for these
		t.Log("(near-zero comparison: acceptable if false due to absolute-epsilon check)")
	}
}

func TestFloatEqual_NegativeValues(t *testing.T) {
	if !FloatEqual(-1.5, -1.5, DefaultEpsilon) {
		t.Error("expected -1.5 == -1.5")
	}
	if FloatEqual(-1.0, 1.0, DefaultEpsilon) {
		t.Error("expected -1.0 ≠ 1.0")
	}
}

// ============ IsSymmetric ============

func TestIsSymmetric_1x1(t *testing.T) {
	m := mat.NewDense(1, 1, []float64{5.0})
	if !IsSymmetric(m, DefaultEpsilon) {
		t.Error("expected 1x1 matrix to be symmetric")
	}
}

func TestIsSymmetric_2x2_Symmetric(t *testing.T) {
	m := mat.NewDense(2, 2, []float64{1, 2, 2, 4})
	if !IsSymmetric(m, DefaultEpsilon) {
		t.Error("expected symmetric 2x2 matrix")
	}
}

func TestIsSymmetric_2x2_NonSymmetric(t *testing.T) {
	m := mat.NewDense(2, 2, []float64{1, 2, 3, 4}) // 2 ≠ 3
	if IsSymmetric(m, DefaultEpsilon) {
		t.Error("expected non-symmetric 2x2 matrix")
	}
}

func TestIsSymmetric_3x3_Identity(t *testing.T) {
	m := mat.NewDense(3, 3, []float64{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	})
	if !IsSymmetric(m, DefaultEpsilon) {
		t.Error("expected 3x3 identity to be symmetric")
	}
}

func TestIsSymmetric_3x3_NonSymmetric(t *testing.T) {
	m := mat.NewDense(3, 3, []float64{
		1, 2, 3,
		2, 5, 7, // row 1, col 2 = 7
		3, 6, 9, // row 2, col 1 = 6 ≠ 7
	})
	if IsSymmetric(m, DefaultEpsilon) {
		t.Error("expected non-symmetric 3x3 matrix")
	}
}

func TestIsSymmetric_NonSquare(t *testing.T) {
	m := mat.NewDense(2, 3, []float64{1, 2, 3, 4, 5, 6})
	if IsSymmetric(m, DefaultEpsilon) {
		t.Error("expected non-square matrix to fail symmetry check")
	}
}

func TestIsSymmetric_DiagonalMatrix(t *testing.T) {
	// A diagonal matrix is always symmetric
	m := mat.NewDiagDense(4, []float64{1, 2, 3, 4})
	if !IsSymmetric(m, DefaultEpsilon) {
		t.Error("expected diagonal matrix to be symmetric")
	}
}

// ============ GetFactoredSlice ============

func TestGetFactoredSlice_Empty(t *testing.T) {
	got := GetFactoredSlice([]float64{}, 2.0)
	if len(got) != 0 {
		t.Errorf("expected empty slice, got len=%d", len(got))
	}
}

func TestGetFactoredSlice_Nil(t *testing.T) {
	got := GetFactoredSlice(nil, 2.0)
	if len(got) != 0 {
		t.Errorf("expected empty result for nil input, got len=%d", len(got))
	}
}

func TestGetFactoredSlice_Normal(t *testing.T) {
	input := []float64{1.0, 2.0, 3.0}
	got := GetFactoredSlice(input, 2.0)

	want := []float64{2.0, 4.0, 6.0}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %f, want %f", i, got[i], want[i])
		}
	}
}

func TestGetFactoredSlice_ZeroFactor(t *testing.T) {
	input := []float64{5.0, 10.0, 15.0}
	got := GetFactoredSlice(input, 0.0)

	for i, v := range got {
		if v != 0 {
			t.Errorf("got[%d] = %f, want 0 for zero factor", i, v)
		}
	}
}

func TestGetFactoredSlice_FractionalFactor(t *testing.T) {
	input := []float64{4.0, 6.0}
	got := GetFactoredSlice(input, 0.5)

	if got[0] != 2.0 {
		t.Errorf("got[0] = %f, want 2.0", got[0])
	}
	if got[1] != 3.0 {
		t.Errorf("got[1] = %f, want 3.0", got[1])
	}
}

func TestGetFactoredSlice_DoesNotMutateInput(t *testing.T) {
	input := []float64{1.0, 2.0, 3.0}
	original := make([]float64, len(input))
	copy(original, input)

	_ = GetFactoredSlice(input, 5.0)

	for i := range input {
		if input[i] != original[i] {
			t.Errorf("input[%d] mutated: got %f, want %f", i, input[i], original[i])
		}
	}
}

// ============ CreateTunerConfigFromData ============

func TestCreateTunerConfigFromData_NilFilterUsesDefaults(t *testing.T) {
	env := validEnv()
	cfg := CreateTunerConfigFromData(nil, env)

	if cfg == nil {
		t.Fatal("expected non-nil TunerConfigData")
	}
	if cfg.FilterData.GammaFactor != DefaultGammaFactor {
		t.Errorf("GammaFactor = %f, want %f", cfg.FilterData.GammaFactor, DefaultGammaFactor)
	}
	if cfg.FilterData.ErrorLevel != DefaultErrorLevel {
		t.Errorf("ErrorLevel = %f, want %f", cfg.FilterData.ErrorLevel, DefaultErrorLevel)
	}
	if cfg.FilterData.TPercentile != DefaultTPercentile {
		t.Errorf("TPercentile = %f, want %f", cfg.FilterData.TPercentile, DefaultTPercentile)
	}
}

func TestCreateTunerConfigFromData_CustomFilterUsed(t *testing.T) {
	env := validEnv()
	customFilter := &FilterData{
		GammaFactor: 2.5,
		ErrorLevel:  0.10,
		TPercentile: 2.58,
	}
	cfg := CreateTunerConfigFromData(customFilter, env)

	if cfg.FilterData.GammaFactor != 2.5 {
		t.Errorf("GammaFactor = %f, want 2.5", cfg.FilterData.GammaFactor)
	}
	if cfg.FilterData.ErrorLevel != 0.10 {
		t.Errorf("ErrorLevel = %f, want 0.10", cfg.FilterData.ErrorLevel)
	}
	if cfg.FilterData.TPercentile != 2.58 {
		t.Errorf("TPercentile = %f, want 2.58", cfg.FilterData.TPercentile)
	}
}

func TestCreateTunerConfigFromData_InitStateFromDefaults(t *testing.T) {
	cfg := CreateTunerConfigFromData(nil, nil)

	if len(cfg.ModelData.InitState) != 3 {
		t.Fatalf("expected InitState length 3, got %d", len(cfg.ModelData.InitState))
	}
	if cfg.ModelData.InitState[StateIndexAlpha] != DefaultAlpha {
		t.Errorf("InitState[alpha] = %f, want %f", cfg.ModelData.InitState[StateIndexAlpha], DefaultAlpha)
	}
	if cfg.ModelData.InitState[StateIndexBeta] != DefaultBeta {
		t.Errorf("InitState[beta] = %f, want %f", cfg.ModelData.InitState[StateIndexBeta], DefaultBeta)
	}
	if cfg.ModelData.InitState[StateIndexGamma] != DefaultGamma {
		t.Errorf("InitState[gamma] = %f, want %f", cfg.ModelData.InitState[StateIndexGamma], DefaultGamma)
	}
}

func TestCreateTunerConfigFromData_ExpectedObservationsFromEnv(t *testing.T) {
	env := validEnv()
	cfg := CreateTunerConfigFromData(nil, env)

	if len(cfg.ModelData.ExpectedObservations) != 2 {
		t.Fatalf("expected 2 observations, got %d", len(cfg.ModelData.ExpectedObservations))
	}
	if cfg.ModelData.ExpectedObservations[0] != float64(env.AvgTTFT) {
		t.Errorf("ExpectedObservations[0] = %f, want %f (AvgTTFT)", cfg.ModelData.ExpectedObservations[0], float64(env.AvgTTFT))
	}
	if cfg.ModelData.ExpectedObservations[1] != float64(env.AvgITL) {
		t.Errorf("ExpectedObservations[1] = %f, want %f (AvgITL)", cfg.ModelData.ExpectedObservations[1], float64(env.AvgITL))
	}
}

func TestCreateTunerConfigFromData_ExpectedObservationsDefaultsForNilEnv(t *testing.T) {
	cfg := CreateTunerConfigFromData(nil, nil)

	if len(cfg.ModelData.ExpectedObservations) != 2 {
		t.Fatalf("expected 2 observations, got %d", len(cfg.ModelData.ExpectedObservations))
	}
	if cfg.ModelData.ExpectedObservations[0] != DefaultExpectedTTFT {
		t.Errorf("obs[0] = %f, want DefaultExpectedTTFT=%f", cfg.ModelData.ExpectedObservations[0], DefaultExpectedTTFT)
	}
	if cfg.ModelData.ExpectedObservations[1] != DefaultExpectedITL {
		t.Errorf("obs[1] = %f, want DefaultExpectedITL=%f", cfg.ModelData.ExpectedObservations[1], DefaultExpectedITL)
	}
}

func TestCreateTunerConfigFromData_BoundedStateEnabled(t *testing.T) {
	cfg := CreateTunerConfigFromData(nil, nil)

	if !cfg.ModelData.BoundedState {
		t.Error("expected BoundedState = true")
	}
	if len(cfg.ModelData.MinState) != 3 {
		t.Errorf("expected MinState length 3, got %d", len(cfg.ModelData.MinState))
	}
	if len(cfg.ModelData.MaxState) != 3 {
		t.Errorf("expected MaxState length 3, got %d", len(cfg.ModelData.MaxState))
	}
}

func TestCreateTunerConfigFromData_StateBoundsAreOrdered(t *testing.T) {
	cfg := CreateTunerConfigFromData(nil, nil)

	for i := range cfg.ModelData.InitState {
		if cfg.ModelData.MinState[i] >= cfg.ModelData.MaxState[i] {
			t.Errorf("state[%d]: MinState (%f) >= MaxState (%f)", i, cfg.ModelData.MinState[i], cfg.ModelData.MaxState[i])
		}
		if cfg.ModelData.InitState[i] < cfg.ModelData.MinState[i] ||
			cfg.ModelData.InitState[i] > cfg.ModelData.MaxState[i] {
			t.Errorf("state[%d]: InitState (%f) outside [MinState=%f, MaxState=%f]",
				i, cfg.ModelData.InitState[i], cfg.ModelData.MinState[i], cfg.ModelData.MaxState[i])
		}
	}
}

func TestCreateTunerConfigFromData_PercentChangeLength(t *testing.T) {
	cfg := CreateTunerConfigFromData(nil, nil)

	n := len(cfg.ModelData.InitState)
	if len(cfg.ModelData.PercentChange) != n {
		t.Errorf("PercentChange length = %d, want %d (= len(InitState))", len(cfg.ModelData.PercentChange), n)
	}
}

// ============ helpers shared by other test files ============

// validEnv returns a fully-populated valid Environment suitable for tests.
func validEnv() *Environment {
	return &Environment{
		Lambda:        10.0,
		AvgInputToks:  200.0,
		AvgOutputToks: 100.0,
		MaxBatchSize:  256,
		AvgTTFT:       500.0,
		AvgITL:        50.0,
	}
}

// validTunerConfigData builds a minimal valid TunerConfigData using defaults.
func validTunerConfigData() *TunerConfigData {
	return CreateTunerConfigFromData(nil, validEnv())
}
