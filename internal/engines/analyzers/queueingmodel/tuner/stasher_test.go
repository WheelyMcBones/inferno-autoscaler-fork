package tuner

import (
	"testing"

	"gonum.org/v1/gonum/mat"
)

func TestNewStasher_NilFilter(t *testing.T) {
	_, err := NewStasher(nil)
	if err == nil {
		t.Error("expected error for nil filter")
	}
}

func TestNewStasher_Valid(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, err := NewTuner(cfg, env)
	if err != nil {
		t.Fatalf("failed to create tuner: %v", err)
	}

	stasher, err := NewStasher(tuner.filter)
	if err != nil {
		t.Fatalf("unexpected error from NewStasher: %v", err)
	}
	if stasher == nil {
		t.Fatal("expected non-nil stasher")
	}
	if stasher.Filter != tuner.filter {
		t.Error("expected stasher to reference the provided filter")
	}
}

func TestStasher_Stash_Succeeds(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)
	stasher, _ := NewStasher(tuner.filter)

	if err := stasher.Stash(); err != nil {
		t.Fatalf("Stash() returned unexpected error: %v", err)
	}
	if stasher.X == nil {
		t.Error("expected stasher.X to be set after Stash()")
	}
	if stasher.P == nil {
		t.Error("expected stasher.P to be set after Stash()")
	}
}

func TestStasher_Stash_CopiesState(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)
	stasher, _ := NewStasher(tuner.filter)
	if err := stasher.Stash(); err != nil {
		t.Fatalf("Stash() failed: %v", err)
	}

	// Stashed X must match the filter X at time of stash
	n := tuner.filter.X.Len()
	for i := 0; i < n; i++ {
		if stasher.X.AtVec(i) != tuner.filter.X.AtVec(i) {
			t.Errorf("stasher.X[%d] = %f, want %f", i, stasher.X.AtVec(i), tuner.filter.X.AtVec(i))
		}
	}
}

func TestStasher_UnStash_BeforeStash(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)
	stasher, _ := NewStasher(tuner.filter)

	// UnStash before Stash should fail (X and P are nil)
	if err := stasher.UnStash(); err == nil {
		t.Error("expected error when UnStash called before Stash")
	}
}

func TestStasher_UnStash_RestoresState(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)
	stasher, _ := NewStasher(tuner.filter)

	// Capture original state
	originalState := mat.VecDenseCopyOf(tuner.filter.X)

	// Stash
	if err := stasher.Stash(); err != nil {
		t.Fatalf("Stash() failed: %v", err)
	}

	// Modify the filter state
	n := tuner.filter.X.Len()
	for i := 0; i < n; i++ {
		tuner.filter.X.SetVec(i, 9999.0)
	}

	// UnStash — should restore original state
	if err := stasher.UnStash(); err != nil {
		t.Fatalf("UnStash() failed: %v", err)
	}

	for i := 0; i < n; i++ {
		if tuner.filter.X.AtVec(i) != originalState.AtVec(i) {
			t.Errorf("state[%d] after UnStash = %f, want %f",
				i, tuner.filter.X.AtVec(i), originalState.AtVec(i))
		}
	}
}

func TestStasher_StashUnStash_IndependentCopies(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)
	stasher, _ := NewStasher(tuner.filter)

	// Stash current state
	if err := stasher.Stash(); err != nil {
		t.Fatalf("Stash() failed: %v", err)
	}

	// Verify stasher.X is a deep copy: mutating filter.X must not affect stasher.X
	firstStashedValue := stasher.X.AtVec(0)
	tuner.filter.X.SetVec(0, firstStashedValue+1000.0)

	if stasher.X.AtVec(0) != firstStashedValue {
		t.Errorf("stasher.X[0] mutated by filter change: got %f, want %f",
			stasher.X.AtVec(0), firstStashedValue)
	}
}

func TestStasher_MultipleStashRestoresCurrent(t *testing.T) {
	env := validEnv()
	cfg := validTunerConfigData()
	tuner, _ := NewTuner(cfg, env)
	stasher, _ := NewStasher(tuner.filter)

	// Stash, then modify state
	if err := stasher.Stash(); err != nil {
		t.Fatalf("first Stash() failed: %v", err)
	}
	tuner.filter.X.SetVec(0, 111.0)

	// Stash again (overwrites previous stash with modified state)
	if err := stasher.Stash(); err != nil {
		t.Fatalf("second Stash() failed: %v", err)
	}
	// UnStash should now restore the second stash (111.0)
	if err := stasher.UnStash(); err != nil {
		t.Fatalf("UnStash() failed: %v", err)
	}
	if tuner.filter.X.AtVec(0) != 111.0 {
		t.Errorf("state[0] after second stash/unstash = %f, want 111.0", tuner.filter.X.AtVec(0))
	}
}
