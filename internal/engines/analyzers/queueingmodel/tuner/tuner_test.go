package tuner

import (
	"math"
	"testing"

	"github.com/llm-d/llm-d-workload-variant-autoscaler/pkg/analyzer"
	"github.com/llm-d/llm-d-workload-variant-autoscaler/pkg/config"
)

func TestNewTuner_NilEnv(t *testing.T) {
	cfg := validTunerConfigData()
	_, err := NewTuner(cfg, nil)
	if err == nil {
		t.Error("expected error for nil environment")
	}
}

func TestNewTuner_InvalidEnv(t *testing.T) {
	cfg := validTunerConfigData()
	env := &Environment{Lambda: 0} // invalid
	_, err := NewTuner(cfg, env)
	if err == nil {
		t.Error("expected error for invalid environment (Lambda=0)")
	}
}

func TestNewTuner_Valid(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, err := NewTuner(cfg, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tuner == nil {
		t.Fatal("expected non-nil Tuner")
	}
}

func TestNewTuner_FilterInitialized(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, err := NewTuner(cfg, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tuner.filter == nil {
		t.Fatal("expected non-nil internal kalman filter")
	}
}

func TestNewTuner_ConfiguratorInitialized(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, err := NewTuner(cfg, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tuner.configurator == nil {
		t.Fatal("expected non-nil configurator")
	}
}

func TestTuner_UpdateEnvironment_Nil(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)

	if err := tuner.UpdateEnvironment(nil); err == nil {
		t.Error("expected error for nil environment")
	}
}

func TestTuner_UpdateEnvironment_Invalid(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)

	invalid := &Environment{Lambda: 0}
	if err := tuner.UpdateEnvironment(invalid); err == nil {
		t.Error("expected error for invalid environment")
	}
}

func TestTuner_UpdateEnvironment_Valid(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)

	newEnv := &Environment{
		Lambda:        20.0,
		AvgInputToks:  100.0,
		AvgOutputToks: 50.0,
		MaxBatchSize:  128,
		AvgTTFT:       300.0,
		AvgITL:        30.0,
	}
	if err := tuner.UpdateEnvironment(newEnv); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tuner.env != newEnv {
		t.Error("expected tuner.env to reference the new environment after UpdateEnvironment")
	}
}

func TestTuner_UpdateEnvironment_ReplacesEnvironment(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)

	newEnv := &Environment{
		Lambda:        5.0,
		AvgInputToks:  50.0,
		AvgOutputToks: 25.0,
		MaxBatchSize:  64,
		AvgTTFT:       200.0,
		AvgITL:        20.0,
	}
	_ = tuner.UpdateEnvironment(newEnv)

	got := tuner.GetEnvironment()
	if got.Lambda != newEnv.Lambda {
		t.Errorf("Lambda = %f, want %f after UpdateEnvironment", got.Lambda, newEnv.Lambda)
	}
}

func TestTuner_GetParms_NonNil(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)

	parms := tuner.GetParms()
	if parms == nil {
		t.Fatal("expected non-nil parameters vector from GetParms()")
	}
}

func TestTuner_GetParms_CorrectLength(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)

	parms := tuner.GetParms()
	if parms.Len() != 3 {
		t.Errorf("GetParms() length = %d, want 3 (alpha, beta, gamma)", parms.Len())
	}
}

func TestTuner_GetParms_MatchesInitState(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)

	parms := tuner.GetParms()
	for i, want := range cfg.ModelData.InitState {
		got := parms.AtVec(i)
		if got != want {
			t.Errorf("parms[%d] = %f, want %f (from InitState)", i, got, want)
		}
	}
}

func TestTuner_GetEnvironment_ReturnsOriginal(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)

	got := tuner.GetEnvironment()
	if got != env {
		t.Error("expected GetEnvironment() to return the original environment")
	}
}

func TestTuner_X_NonNil(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)
	if tuner.X() == nil {
		t.Error("expected non-nil X()")
	}
}

func TestTuner_P_NonNil(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)
	if tuner.P() == nil {
		t.Error("expected non-nil P()")
	}
}

func TestTuner_Run_ReturnsResults(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, err := NewTuner(cfg, env)
	if err != nil {
		t.Fatalf("NewTuner failed: %v", err)
	}

	results, err := tuner.Run()
	if err != nil {
		t.Fatalf("Run() returned unexpected error: %v", err)
	}
	if results == nil {
		t.Fatal("expected non-nil TunedResults from Run()")
	}
}

func TestTuner_Run_ServiceParmsNonNil(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)

	results, err := tuner.Run()
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if results.ServiceParms == nil {
		t.Fatal("expected non-nil ServiceParms in TunedResults")
	}
}

func TestTuner_Run_PositiveParamsOrValidationFailed(t *testing.T) {
	// When validation succeeds, all params must be positive.
	// When validation fails, we still get the previous (initial) state — also positive.
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)

	results, err := tuner.Run()
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if results.ServiceParms.Alpha <= 0 {
		t.Errorf("Alpha = %f, want > 0", results.ServiceParms.Alpha)
	}
	if results.ServiceParms.Beta <= 0 {
		t.Errorf("Beta = %f, want > 0", results.ServiceParms.Beta)
	}
	if results.ServiceParms.Gamma <= 0 {
		t.Errorf("Gamma = %f, want > 0", results.ServiceParms.Gamma)
	}
}

func TestTuner_Run_WithInvalidEnvFails(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)

	// Replace env with an invalid one — Run should reject it
	tuner.env = &Environment{Lambda: 0}
	_, err := tuner.Run()
	if err == nil {
		t.Error("expected error when Run() called with invalid environment")
	}
}

func TestTuner_Run_CanRunMultipleTimes(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)

	for i := 0; i < 3; i++ {
		results, err := tuner.Run()
		if err != nil {
			t.Fatalf("Run() iteration %d error: %v", i, err)
		}
		if results == nil {
			t.Fatalf("Run() iteration %d returned nil results", i)
		}
	}
}

func TestTuner_Run_CovarianceNonNilOnSuccess(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)

	results, err := tuner.Run()
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if results.Covariance == nil {
		t.Error("expected non-nil Covariance in TunedResults")
	}
}

func TestTuner_String_NonEmpty(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)

	s := tuner.String()
	if s == "" {
		t.Error("expected non-empty String() output")
	}
}

// =============================================================================
// Semantic / self-consistency tests for the tuner (EKF).
//
// These use SYNTHETIC (α, β, γ) — we pick arbitrary valid parameters,
// generate ground-truth observations via the same queueing model the tuner
// uses internally, and verify properties like convergence, stability, and
// robustness. No external benchmark data is needed.
// =============================================================================

// --- Tolerances and constants ------------------------------------------------
//
// All tolerances are justified below. Where a single tolerance is used across
// multiple tests, it appears here as a named constant.
//
//   - semanticAbsTolObsFunc (1e-3 ms): the observation function and direct
//     Analyze() follow the same code path with the same float32 inputs.
//     The only source of mismatch is float64↔float32 round-trip, which is
//     well under 1e-3 ms for latencies in the 1–1000 ms range.
//
//   - semanticRelTolStability (5%): when initialized at the true state and
//     fed noiseless observations, the EKF may still drift slightly because
//     the process-noise covariance Q is nonzero. 5% covers the maximum drift
//     observed across all param sets over 20 iterations.
//
//   - semanticInitOffsetFraction (15%): convergence tests start the filter
//     15% away from the true parameters. This is the largest offset that
//     reliably stays within the NIS acceptance region (χ² threshold 7.378
//     at 2 DOF, 97.5%), ensuring the filter can accept updates and converge
//     rather than rejecting every step. Larger offsets (e.g. 30%) cause
//     all-rejected runs for some param sets.
//
//   - semanticConvergenceIters (100): with a 5% per-step EKF gain and
//     15% initial offset, 100 iterations provides enough accepted updates
//     for measurable convergence. Empirically, α and β converge within
//     40-60 accepted iterations; 100 adds margin for NIS-rejected steps.
//
//   - semanticOutlierMultiplier (100×): inflates observations by 100× to
//     guarantee NIS rejection. At 100×, the innovation is ~10000× the
//     predicted covariance, far exceeding the χ² threshold of 7.378.

const (
	semanticAbsTolObsFunc      = 1e-3  // ms — float32 round-trip tolerance
	semanticRelTolStability    = 0.05  // 5% — max drift at true params over 20 iters
	semanticInitOffsetFraction = 1.15  // start 15% above truth for convergence tests
	semanticConvergenceIters   = 100   // iterations for convergence tests
	semanticOutlierMultiplier  = 100.0 // observation multiplier to trigger NIS rejection
	semanticStabilityIters     = 20    // iterations for stability check
	semanticResidualIters      = 50    // iterations for residual-decrease check
	semanticWarmUpIters        = 5     // iterations to settle covariance before outlier test
)

// --- Synthetic parameter sets ------------------------------------------------

type syntheticTunerParams struct {
	name  string
	alpha float64
	beta  float64
	gamma float64
}

// tunerParamSets spans two hardware profiles. We use two rather than three
// to keep EKF convergence tests fast (~0.1s total).
var tunerParamSets = []syntheticTunerParams{
	{"gpu1", 5.0, 0.03, 0.00005},
	{"gpu2", 12.0, 0.08, 0.00015},
}

// --- Helper: generate a ground-truth Environment from known (α,β,γ) ----------

// syntheticEnv generates an Environment whose TTFT/ITL observations are exactly
// what the queueing model predicts for the given (α,β,γ) at the given operating
// point. Uses the SAME code path as makeObservationFunc() inside the tuner.
func syntheticEnv(
	t *testing.T,
	alpha, beta, gamma float64,
	lambdaReqPerSec float64,
	avgInputToks, avgOutputToks float32,
	maxBatchSize int,
) *Environment {
	t.Helper()

	maxQueue := maxBatchSize * config.MaxQueueToBatchRatio
	qCfg := &analyzer.Configuration{
		MaxBatchSize: maxBatchSize,
		MaxQueueSize: maxQueue,
		ServiceParms: &analyzer.ServiceParms{
			Alpha: float32(alpha),
			Beta:  float32(beta),
			Gamma: float32(gamma),
		},
	}
	rs := &analyzer.RequestSize{
		AvgInputTokens:  avgInputToks,
		AvgOutputTokens: avgOutputToks,
	}
	qa, err := analyzer.NewQueueAnalyzer(qCfg, rs)
	if err != nil {
		t.Fatalf("syntheticEnv: NewQueueAnalyzer failed: %v", err)
	}
	metrics, err := qa.Analyze(float32(lambdaReqPerSec))
	if err != nil {
		t.Fatalf("syntheticEnv: Analyze failed: %v", err)
	}

	// TTFT formula matches makeObservationFunc(): WaitTime + PrefillTime
	ttft := metrics.AvgWaitTime + metrics.AvgPrefillTime
	itl := metrics.AvgTokenTime

	return &Environment{
		Lambda:        float32(lambdaReqPerSec * 60), // req/min
		AvgInputToks:  avgInputToks,
		AvgOutputToks: avgOutputToks,
		MaxBatchSize:  maxBatchSize,
		AvgTTFT:       ttft,
		AvgITL:        itl,
	}
}

// --- Helper: create a tuner initialized at given (α,β,γ) --------------------

func tunerAtParams(t *testing.T, alpha, beta, gamma float64, env *Environment) *Tuner {
	t.Helper()
	cfg := CreateTunerConfigFromData(nil, env)
	cfg.ModelData.InitState = []float64{alpha, beta, gamma}
	cfg.ModelData.MinState = GetFactoredSlice(cfg.ModelData.InitState, DefaultMinStateFactor)
	cfg.ModelData.MaxState = GetFactoredSlice(cfg.ModelData.InitState, DefaultMaxStateFactor)

	tn, err := NewTuner(cfg, env)
	if err != nil {
		t.Fatalf("tunerAtParams: NewTuner failed: %v", err)
	}
	return tn
}

// =============================================================================
// Test: Observation function matches direct QueueAnalyzer.Analyze()
// =============================================================================

func Test_ObservationFunction_MatchesAnalyzer(t *testing.T) {
	for _, sp := range tunerParamSets {
		t.Run(sp.name, func(t *testing.T) {
			lambdaReqPerSec := 3.0
			env := syntheticEnv(t, sp.alpha, sp.beta, sp.gamma,
				lambdaReqPerSec, 256, 128, 64)

			tn := tunerAtParams(t, sp.alpha, sp.beta, sp.gamma, env)

			obsFunc := tn.makeObservationFunc()
			predicted := obsFunc(tn.X())
			if predicted == nil {
				t.Fatal("observation function returned nil")
			}

			obsTTFT := float64(env.AvgTTFT)
			obsITL := float64(env.AvgITL)

			if math.Abs(predicted.AtVec(0)-obsTTFT) > semanticAbsTolObsFunc {
				t.Errorf("TTFT mismatch: predicted=%.6f observed=%.6f", predicted.AtVec(0), obsTTFT)
			}
			if math.Abs(predicted.AtVec(1)-obsITL) > semanticAbsTolObsFunc {
				t.Errorf("ITL mismatch: predicted=%.6f observed=%.6f", predicted.AtVec(1), obsITL)
			}
		})
	}
}

// =============================================================================
// Test: Stability at true parameters — filter does not drift
// =============================================================================

func Test_StabilityAtTrueParams(t *testing.T) {
	for _, sp := range tunerParamSets {
		t.Run(sp.name, func(t *testing.T) {
			env := syntheticEnv(t, sp.alpha, sp.beta, sp.gamma,
				2.0, 256, 128, 64)

			tn := tunerAtParams(t, sp.alpha, sp.beta, sp.gamma, env)

			for i := range semanticStabilityIters {
				res, err := tn.Run()
				if err != nil {
					t.Fatalf("Run() iteration %d failed: %v", i, err)
				}
				if res.ValidationFailed {
					t.Fatalf("iteration %d: unexpected validation failure (NIS=%.4f)", i, res.NIS)
				}
			}

			for _, tc := range []struct {
				name  string
				got   float64
				truth float64
			}{
				{"alpha", tn.X().AtVec(StateIndexAlpha), sp.alpha},
				{"beta", tn.X().AtVec(StateIndexBeta), sp.beta},
				{"gamma", tn.X().AtVec(StateIndexGamma), sp.gamma},
			} {
				relErr := math.Abs(tc.got-tc.truth) / tc.truth
				if relErr > semanticRelTolStability {
					t.Errorf("%s drifted: got=%.6f true=%.6f relErr=%.2f%%",
						tc.name, tc.got, tc.truth, relErr*100)
				}
			}
		})
	}
}

// =============================================================================
// Test: Self-consistency convergence
//
// Start the filter 15% away from the true (α,β,γ) — the largest offset that
// stays within the NIS acceptance region — and verify the error decreases
// over 100 iterations.
// =============================================================================

func Test_SelfConsistencyConvergence(t *testing.T) {
	for _, sp := range tunerParamSets {
		t.Run(sp.name, func(t *testing.T) {
			type opPoint struct {
				lambda    float64
				inputTok  float32
				outputTok float32
			}
			ops := []opPoint{
				{2.0, 256, 128},
				{1.5, 512, 64},
				{3.0, 128, 256},
				{1.0, 1024, 128},
			}

			initAlpha := sp.alpha * semanticInitOffsetFraction
			initBeta := sp.beta * semanticInitOffsetFraction
			initGamma := sp.gamma * semanticInitOffsetFraction

			firstEnv := syntheticEnv(t, sp.alpha, sp.beta, sp.gamma,
				ops[0].lambda, ops[0].inputTok, ops[0].outputTok, 64)

			tn := tunerAtParams(t, initAlpha, initBeta, initGamma, firstEnv)

			initialErrAlpha := math.Abs(tn.X().AtVec(StateIndexAlpha) - sp.alpha)
			initialErrBeta := math.Abs(tn.X().AtVec(StateIndexBeta) - sp.beta)

			accepted := 0
			for i := range semanticConvergenceIters {
				op := ops[i%len(ops)]
				env := syntheticEnv(t, sp.alpha, sp.beta, sp.gamma,
					op.lambda, op.inputTok, op.outputTok, 64)

				if err := tn.UpdateEnvironment(env); err != nil {
					t.Fatalf("UpdateEnvironment iteration %d failed: %v", i, err)
				}
				res, err := tn.Run()
				if err != nil {
					t.Fatalf("Run iteration %d failed: %v", i, err)
				}
				if !res.ValidationFailed {
					accepted++
				}
			}

			if accepted == 0 {
				t.Fatal("all iterations were NIS-rejected; no convergence possible")
			}

			finalErrAlpha := math.Abs(tn.X().AtVec(StateIndexAlpha) - sp.alpha)
			finalErrBeta := math.Abs(tn.X().AtVec(StateIndexBeta) - sp.beta)

			if finalErrAlpha >= initialErrAlpha {
				t.Errorf("alpha did not converge toward truth: initialErr=%.6f finalErr=%.6f",
					initialErrAlpha, finalErrAlpha)
			}
			if finalErrBeta >= initialErrBeta {
				t.Errorf("beta did not converge toward truth: initialErr=%.6f finalErr=%.6f",
					initialErrBeta, finalErrBeta)
			}
		})
	}
}

// =============================================================================
// Test: Multi-operating-point convergence
//
// Wider variety of traffic patterns; verifies the filter converges without
// diverging under workload variation.
// =============================================================================

func Test_MultiOperatingPointConvergence(t *testing.T) {
	sp := syntheticTunerParams{"gpu1", 5.0, 0.03, 0.00005}

	type opPoint struct {
		lambda    float64
		inputTok  float32
		outputTok float32
	}
	ops := []opPoint{
		{1.0, 64, 64},
		{2.0, 256, 128},
		{0.5, 1024, 256},
		{3.0, 128, 512},
		{1.5, 512, 64},
		{2.5, 256, 256},
	}

	firstEnv := syntheticEnv(t, sp.alpha, sp.beta, sp.gamma,
		ops[0].lambda, ops[0].inputTok, ops[0].outputTok, 128)

	tn := tunerAtParams(t,
		sp.alpha*semanticInitOffsetFraction,
		sp.beta*semanticInitOffsetFraction,
		sp.gamma*semanticInitOffsetFraction,
		firstEnv)

	initialErrAlpha := math.Abs(tn.X().AtVec(StateIndexAlpha) - sp.alpha)

	accepted := 0
	for i := range semanticConvergenceIters {
		op := ops[i%len(ops)]
		env := syntheticEnv(t, sp.alpha, sp.beta, sp.gamma,
			op.lambda, op.inputTok, op.outputTok, 128)

		if err := tn.UpdateEnvironment(env); err != nil {
			t.Fatalf("UpdateEnvironment iteration %d failed: %v", i, err)
		}
		res, err := tn.Run()
		if err != nil {
			t.Fatalf("Run iteration %d failed: %v", i, err)
		}
		if !res.ValidationFailed {
			accepted++
		}
	}

	if accepted == 0 {
		t.Fatal("all iterations NIS-rejected")
	}

	finalErrAlpha := math.Abs(tn.X().AtVec(StateIndexAlpha) - sp.alpha)
	if finalErrAlpha >= initialErrAlpha {
		t.Errorf("alpha did not converge: initialErr=%.6f finalErr=%.6f",
			initialErrAlpha, finalErrAlpha)
	}
}

// =============================================================================
// Test: NIS rejection on outlier observation
//
// An outlier produces an innovation exceeding the NIS threshold
// so the filter must reject it and restore the previous state.
// =============================================================================

func Test_NISRejectionOnOutlier(t *testing.T) {
	sp := tunerParamSets[0] // gpu1

	env := syntheticEnv(t, sp.alpha, sp.beta, sp.gamma,
		2.0, 256, 128, 64)
	tn := tunerAtParams(t, sp.alpha, sp.beta, sp.gamma, env)

	// Warm up to settle covariance
	for i := range semanticWarmUpIters {
		res, err := tn.Run()
		if err != nil {
			t.Fatalf("warm-up Run %d failed: %v", i, err)
		}
		if res.ValidationFailed {
			t.Fatalf("warm-up iteration %d: unexpected validation failure", i)
		}
	}

	alphaBefore := tn.X().AtVec(StateIndexAlpha)
	betaBefore := tn.X().AtVec(StateIndexBeta)
	gammaBefore := tn.X().AtVec(StateIndexGamma)

	outlierEnv := &Environment{
		Lambda:        env.Lambda,
		AvgInputToks:  env.AvgInputToks,
		AvgOutputToks: env.AvgOutputToks,
		MaxBatchSize:  env.MaxBatchSize,
		AvgTTFT:       env.AvgTTFT * semanticOutlierMultiplier,
		AvgITL:        env.AvgITL * semanticOutlierMultiplier,
	}
	if err := tn.UpdateEnvironment(outlierEnv); err != nil {
		t.Fatalf("UpdateEnvironment with outlier failed: %v", err)
	}

	res, err := tn.Run()
	if err != nil {
		t.Fatalf("Run with outlier failed: %v", err)
	}

	if !res.ValidationFailed {
		t.Errorf("expected ValidationFailed=true for %.0f× outlier, got NIS=%.4f",
			semanticOutlierMultiplier, res.NIS)
	}

	// Parameters must be unchanged (unstashed)
	if tn.X().AtVec(StateIndexAlpha) != alphaBefore {
		t.Errorf("alpha changed after outlier rejection: before=%.6f after=%.6f",
			alphaBefore, tn.X().AtVec(StateIndexAlpha))
	}
	if tn.X().AtVec(StateIndexBeta) != betaBefore {
		t.Errorf("beta changed after outlier rejection: before=%.6f after=%.6f",
			betaBefore, tn.X().AtVec(StateIndexBeta))
	}
	if tn.X().AtVec(StateIndexGamma) != gammaBefore {
		t.Errorf("gamma changed after outlier rejection: before=%.6f after=%.6f",
			gammaBefore, tn.X().AtVec(StateIndexGamma))
	}
}

// =============================================================================
// Test: Convergence reduces observation residual
//
// The filter should improve its predictions: the L2 residual between predicted
// and observed (TTFT, ITL) should decrease over iterations.
// =============================================================================

func Test_ConvergenceReducesResidual(t *testing.T) {
	sp := syntheticTunerParams{"gpu2", 12.0, 0.08, 0.00015}

	env := syntheticEnv(t, sp.alpha, sp.beta, sp.gamma,
		1.5, 512, 128, 64)

	tn := tunerAtParams(t,
		sp.alpha*semanticInitOffsetFraction,
		sp.beta*semanticInitOffsetFraction,
		sp.gamma*semanticInitOffsetFraction,
		env)

	obsFunc := tn.makeObservationFunc()
	initialPred := obsFunc(tn.X())
	if initialPred == nil {
		t.Fatal("initial observation function returned nil")
	}
	initialResidual := math.Sqrt(
		math.Pow(initialPred.AtVec(0)-float64(env.AvgTTFT), 2) +
			math.Pow(initialPred.AtVec(1)-float64(env.AvgITL), 2),
	)

	for i := range semanticResidualIters {
		if _, err := tn.Run(); err != nil {
			t.Fatalf("Run iteration %d failed: %v", i, err)
		}
	}

	finalPred := obsFunc(tn.X())
	if finalPred == nil {
		t.Fatal("final observation function returned nil")
	}
	finalResidual := math.Sqrt(
		math.Pow(finalPred.AtVec(0)-float64(env.AvgTTFT), 2) +
			math.Pow(finalPred.AtVec(1)-float64(env.AvgITL), 2),
	)

	if finalResidual >= initialResidual {
		t.Errorf("residual did not decrease: initial=%.4f final=%.4f",
			initialResidual, finalResidual)
	}
}
