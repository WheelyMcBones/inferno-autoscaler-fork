package tuner

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/mat"
)

func TestCheckConfigData_Valid(t *testing.T) {
	cd := validTunerConfigData()
	if !checkConfigData(cd) {
		t.Error("expected checkConfigData = true for valid config")
	}
}

func TestCheckConfigData_Nil(t *testing.T) {
	if checkConfigData(nil) {
		t.Error("expected checkConfigData = false for nil config")
	}
}

func TestCheckConfigData_ZeroGammaFactor(t *testing.T) {
	cd := validTunerConfigData()
	cd.FilterData.GammaFactor = 0
	if checkConfigData(cd) {
		t.Error("expected checkConfigData = false for GammaFactor=0")
	}
}

func TestCheckConfigData_ZeroErrorLevel(t *testing.T) {
	cd := validTunerConfigData()
	cd.FilterData.ErrorLevel = 0
	if checkConfigData(cd) {
		t.Error("expected checkConfigData = false for ErrorLevel=0")
	}
}

func TestCheckConfigData_ZeroTPercentile(t *testing.T) {
	cd := validTunerConfigData()
	cd.FilterData.TPercentile = 0
	if checkConfigData(cd) {
		t.Error("expected checkConfigData = false for TPercentile=0")
	}
}

func TestCheckConfigData_EmptyInitState(t *testing.T) {
	cd := validTunerConfigData()
	cd.ModelData.InitState = []float64{}
	if checkConfigData(cd) {
		t.Error("expected checkConfigData = false for empty InitState")
	}
}

func TestCheckConfigData_NaNInInitState(t *testing.T) {
	cd := validTunerConfigData()
	cd.ModelData.InitState[0] = math.NaN()
	if checkConfigData(cd) {
		t.Error("expected checkConfigData = false for NaN in InitState")
	}
}

func TestCheckConfigData_InfInInitState(t *testing.T) {
	cd := validTunerConfigData()
	cd.ModelData.InitState[1] = math.Inf(1)
	if checkConfigData(cd) {
		t.Error("expected checkConfigData = false for Inf in InitState")
	}
}

func TestCheckConfigData_PercentChangeLengthMismatch(t *testing.T) {
	cd := validTunerConfigData()
	cd.ModelData.PercentChange = []float64{DefaultPercentChange} // wrong length
	if checkConfigData(cd) {
		t.Error("expected checkConfigData = false for PercentChange length mismatch")
	}
}

func TestCheckConfigData_ZeroPercentChange(t *testing.T) {
	cd := validTunerConfigData()
	cd.ModelData.PercentChange[0] = 0 // zero is not allowed
	if checkConfigData(cd) {
		t.Error("expected checkConfigData = false for zero PercentChange")
	}
}

func TestCheckConfigData_NegativePercentChange(t *testing.T) {
	cd := validTunerConfigData()
	cd.ModelData.PercentChange[0] = -0.05
	if checkConfigData(cd) {
		t.Error("expected checkConfigData = false for negative PercentChange")
	}
}

func TestCheckConfigData_BoundedWithWrongMinLen(t *testing.T) {
	cd := validTunerConfigData()
	cd.ModelData.BoundedState = true
	cd.ModelData.MinState = []float64{0.01} // wrong length (1 instead of 3)
	if checkConfigData(cd) {
		t.Error("expected checkConfigData = false for MinState length mismatch")
	}
}

func TestCheckConfigData_BoundedMinGreaterThanMax(t *testing.T) {
	cd := validTunerConfigData()
	cd.ModelData.BoundedState = true
	// Swap min and max so MinState[0] > MaxState[0]
	cd.ModelData.MinState[0], cd.ModelData.MaxState[0] = cd.ModelData.MaxState[0], cd.ModelData.MinState[0]
	if checkConfigData(cd) {
		t.Error("expected checkConfigData = false for MinState >= MaxState")
	}
}

func TestCheckConfigData_EmptyExpectedObservations(t *testing.T) {
	cd := validTunerConfigData()
	cd.ModelData.ExpectedObservations = []float64{}
	if checkConfigData(cd) {
		t.Error("expected checkConfigData = false for empty ExpectedObservations")
	}
}

func TestCheckConfigData_NaNInExpectedObservations(t *testing.T) {
	cd := validTunerConfigData()
	cd.ModelData.ExpectedObservations[0] = math.NaN()
	if checkConfigData(cd) {
		t.Error("expected checkConfigData = false for NaN in ExpectedObservations")
	}
}

func TestCheckConfigData_ProvidedCovarianceWrongSize(t *testing.T) {
	cd := validTunerConfigData()
	// 3 states, so covariance must be 3x3 = 9 elements; provide 4 instead
	cd.ModelData.InitCovarianceMatrix = []float64{1, 0, 0, 1}
	if checkConfigData(cd) {
		t.Error("expected checkConfigData = false for covariance matrix wrong size")
	}
}

func TestCheckConfigData_ProvidedCovarianceNonSymmetric(t *testing.T) {
	cd := validTunerConfigData()
	// 3x3 but not symmetric
	cd.ModelData.InitCovarianceMatrix = []float64{
		1, 2, 3,
		0, 1, 0,
		0, 0, 1,
	}
	if checkConfigData(cd) {
		t.Error("expected checkConfigData = false for non-symmetric covariance matrix")
	}
}

func TestCheckConfigData_ProvidedCovarianceSymmetric(t *testing.T) {
	cd := validTunerConfigData()
	// Valid symmetric 3x3 covariance matrix
	cd.ModelData.InitCovarianceMatrix = []float64{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}
	if !checkConfigData(cd) {
		t.Error("expected checkConfigData = true for valid symmetric covariance matrix")
	}
}

func TestNewConfigurator_Valid(t *testing.T) {
	cd := validTunerConfigData()
	c, err := NewConfigurator(cd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil Configurator")
	}
}

func TestNewConfigurator_NilConfig(t *testing.T) {
	_, err := NewConfigurator(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestNewConfigurator_InvalidConfig(t *testing.T) {
	cd := &TunerConfigData{} // empty/invalid
	_, err := NewConfigurator(cd)
	if err == nil {
		t.Error("expected error for invalid config")
	}
}

func TestNewConfigurator_NumStates(t *testing.T) {
	cd := validTunerConfigData()
	c, err := NewConfigurator(cd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.NumStates() != len(cd.ModelData.InitState) {
		t.Errorf("NumStates() = %d, want %d", c.NumStates(), len(cd.ModelData.InitState))
	}
}

func TestNewConfigurator_NumObservations(t *testing.T) {
	cd := validTunerConfigData()
	c, err := NewConfigurator(cd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.NumObservations() != len(cd.ModelData.ExpectedObservations) {
		t.Errorf("NumObservations() = %d, want %d", c.NumObservations(), len(cd.ModelData.ExpectedObservations))
	}
}

func TestNewConfigurator_WithProvidedCovariance(t *testing.T) {
	cd := validTunerConfigData()
	cd.ModelData.InitCovarianceMatrix = []float64{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}
	c, err := NewConfigurator(cd)
	if err != nil {
		t.Fatalf("unexpected error with provided covariance: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil Configurator with provided covariance")
	}
}

func TestGetStateCov_MatchesStateLength(t *testing.T) {
	cd := validTunerConfigData()
	c, err := NewConfigurator(cd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	x := mat.NewVecDense(c.NumStates(), cd.ModelData.InitState)
	cov, err := c.GetStateCov(x)
	if err != nil {
		t.Fatalf("unexpected error from GetStateCov: %v", err)
	}

	rows, cols := cov.Dims()
	if rows != c.NumStates() || cols != c.NumStates() {
		t.Errorf("covariance dims = (%d, %d), want (%d, %d)", rows, cols, c.NumStates(), c.NumStates())
	}
}

func TestGetStateCov_WrongLengthStateVector(t *testing.T) {
	cd := validTunerConfigData()
	c, err := NewConfigurator(cd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Pass a state vector with wrong length
	x := mat.NewVecDense(1, []float64{1.0})
	_, err = c.GetStateCov(x)
	if err == nil {
		t.Error("expected error for state vector with wrong length")
	}
}

func TestGetStateCov_IsDiagonal(t *testing.T) {
	cd := validTunerConfigData()
	c, err := NewConfigurator(cd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	x := mat.NewVecDense(c.NumStates(), cd.ModelData.InitState)
	cov, err := c.GetStateCov(x)
	if err != nil {
		t.Fatalf("unexpected error from GetStateCov: %v", err)
	}

	// Off-diagonal elements should be zero (diagonal matrix)
	n := c.NumStates()
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i != j && cov.At(i, j) != 0 {
				t.Errorf("cov[%d][%d] = %f, want 0 (expected diagonal)", i, j, cov.At(i, j))
			}
		}
	}
}

func TestGetStateCov_IsSymmetric(t *testing.T) {
	cd := validTunerConfigData()
	c, err := NewConfigurator(cd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	x := mat.NewVecDense(c.NumStates(), cd.ModelData.InitState)
	cov, err := c.GetStateCov(x)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !IsSymmetric(cov, DefaultEpsilon) {
		t.Error("expected state covariance to be symmetric")
	}
}
