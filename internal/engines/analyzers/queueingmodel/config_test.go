package queueingmodel

import (
	"testing"
)

func TestQMConfig_GetAnalyzerName(t *testing.T) {
	cfg := &QMConfig{}
	if got := cfg.GetAnalyzerName(); got != "queueing-model" {
		t.Errorf("GetAnalyzerName() = %q, want \"queueing-model\"", got)
	}
}

func TestQMConfig_GetSLOForModel_NilTargets(t *testing.T) {
	cfg := &QMConfig{SLOTargets: nil}
	slo := cfg.GetSLOForModel("default", "llama")
	if slo != nil {
		t.Errorf("expected nil SLO when SLOTargets is nil, got %+v", slo)
	}
}

func TestQMConfig_GetSLOForModel_EmptyTargets(t *testing.T) {
	cfg := &QMConfig{SLOTargets: map[string]*SLOTarget{}}
	slo := cfg.GetSLOForModel("default", "llama")
	if slo != nil {
		t.Errorf("expected nil SLO for empty targets, got %+v", slo)
	}
}

func TestQMConfig_GetSLOForModel_MissingKey(t *testing.T) {
	cfg := &QMConfig{
		SLOTargets: map[string]*SLOTarget{
			"other-ns/other-model": {TargetTTFT: 100, TargetITL: 10},
		},
	}
	slo := cfg.GetSLOForModel("default", "llama")
	if slo != nil {
		t.Errorf("expected nil SLO for missing key, got %+v", slo)
	}
}

func TestQMConfig_GetSLOForModel_ExistingKey(t *testing.T) {
	target := &SLOTarget{TargetTTFT: 500.0, TargetITL: 50.0}
	key := MakeModelKey("default", "llama")
	cfg := &QMConfig{
		SLOTargets: map[string]*SLOTarget{
			key: target,
		},
	}
	slo := cfg.GetSLOForModel("default", "llama")
	if slo == nil {
		t.Fatal("expected non-nil SLO for existing key")
	}
	if slo.TargetTTFT != target.TargetTTFT {
		t.Errorf("TargetTTFT = %f, want %f", slo.TargetTTFT, target.TargetTTFT)
	}
	if slo.TargetITL != target.TargetITL {
		t.Errorf("TargetITL = %f, want %f", slo.TargetITL, target.TargetITL)
	}
}

func TestQMConfig_GetSLOForModel_MultipleEntries(t *testing.T) {
	cfg := &QMConfig{
		SLOTargets: map[string]*SLOTarget{
			"ns1/model-a": {TargetTTFT: 100, TargetITL: 10},
			"ns2/model-b": {TargetTTFT: 200, TargetITL: 20},
			"ns1/model-c": {TargetTTFT: 300, TargetITL: 30},
		},
	}

	tests := []struct {
		ns, model string
		wantTTFT  float32
		wantNil   bool
	}{
		{"ns1", "model-a", 100, false},
		{"ns2", "model-b", 200, false},
		{"ns1", "model-c", 300, false},
		{"ns1", "model-b", 0, true}, // wrong namespace
		{"ns3", "model-a", 0, true}, // wrong namespace
	}

	for _, tt := range tests {
		slo := cfg.GetSLOForModel(tt.ns, tt.model)
		if tt.wantNil {
			if slo != nil {
				t.Errorf("GetSLOForModel(%q, %q): expected nil, got %+v", tt.ns, tt.model, slo)
			}
		} else {
			if slo == nil {
				t.Errorf("GetSLOForModel(%q, %q): expected non-nil", tt.ns, tt.model)
			} else if slo.TargetTTFT != tt.wantTTFT {
				t.Errorf("GetSLOForModel(%q, %q): TargetTTFT = %f, want %f",
					tt.ns, tt.model, slo.TargetTTFT, tt.wantTTFT)
			}
		}
	}
}

// ============ SLOTarget.Max ============

func TestSLOTarget_Max_WithNil(t *testing.T) {
	slo := &SLOTarget{TargetTTFT: 500, TargetITL: 50}
	slo.Max(nil)

	if slo.TargetTTFT != 500 {
		t.Errorf("TargetTTFT changed after Max(nil): got %f", slo.TargetTTFT)
	}
	if slo.TargetITL != 50 {
		t.Errorf("TargetITL changed after Max(nil): got %f", slo.TargetITL)
	}
}

func TestSLOTarget_Max_OtherHigher(t *testing.T) {
	slo := &SLOTarget{TargetTTFT: 100, TargetITL: 10}
	other := &SLOTarget{TargetTTFT: 300, TargetITL: 30}
	slo.Max(other)

	if slo.TargetTTFT != 300 {
		t.Errorf("TargetTTFT = %f, want 300 (max)", slo.TargetTTFT)
	}
	if slo.TargetITL != 30 {
		t.Errorf("TargetITL = %f, want 30 (max)", slo.TargetITL)
	}
}

func TestSLOTarget_Max_SelfHigher(t *testing.T) {
	slo := &SLOTarget{TargetTTFT: 600, TargetITL: 60}
	other := &SLOTarget{TargetTTFT: 200, TargetITL: 20}
	slo.Max(other)

	if slo.TargetTTFT != 600 {
		t.Errorf("TargetTTFT = %f, want 600 (self is higher)", slo.TargetTTFT)
	}
	if slo.TargetITL != 60 {
		t.Errorf("TargetITL = %f, want 60 (self is higher)", slo.TargetITL)
	}
}

func TestSLOTarget_Max_MixedHigher(t *testing.T) {
	// self TTFT higher, other ITL higher
	slo := &SLOTarget{TargetTTFT: 500, TargetITL: 10}
	other := &SLOTarget{TargetTTFT: 200, TargetITL: 80}
	slo.Max(other)

	if slo.TargetTTFT != 500 {
		t.Errorf("TargetTTFT = %f, want 500 (self higher)", slo.TargetTTFT)
	}
	if slo.TargetITL != 80 {
		t.Errorf("TargetITL = %f, want 80 (other higher)", slo.TargetITL)
	}
}

func TestSLOTarget_Max_Equal(t *testing.T) {
	slo := &SLOTarget{TargetTTFT: 300, TargetITL: 30}
	other := &SLOTarget{TargetTTFT: 300, TargetITL: 30}
	slo.Max(other)

	if slo.TargetTTFT != 300 {
		t.Errorf("TargetTTFT = %f, want 300 for equal inputs", slo.TargetTTFT)
	}
	if slo.TargetITL != 30 {
		t.Errorf("TargetITL = %f, want 30 for equal inputs", slo.TargetITL)
	}
}

// ============ QMConfig struct defaults ============

func TestQMConfig_DefaultSLOMultiplier(t *testing.T) {
	// DefaultSLOMultiplier constant should be > 1
	if DefaultSLOMultiplier <= 1.0 {
		t.Errorf("DefaultSLOMultiplier = %f must be > 1.0", DefaultSLOMultiplier)
	}
}

func TestQMConfig_SLOTargets_KeyFormat(t *testing.T) {
	// Verify that GetSLOForModel uses MakeModelKey format
	ns := "my-namespace"
	modelID := "my-model"
	key := MakeModelKey(ns, modelID)

	target := &SLOTarget{TargetTTFT: 250, TargetITL: 25}
	cfg := &QMConfig{
		SLOTargets: map[string]*SLOTarget{key: target},
	}

	// Both lookup and storage must use the same key format
	got := cfg.GetSLOForModel(ns, modelID)
	if got == nil {
		t.Errorf("GetSLOForModel: expected to find SLO stored under MakeModelKey(%q, %q) = %q", ns, modelID, key)
	}
}
