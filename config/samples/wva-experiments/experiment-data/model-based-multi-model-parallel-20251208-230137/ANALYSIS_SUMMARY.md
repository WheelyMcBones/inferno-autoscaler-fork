# WVA Model-Based Experiment Analysis

**Experiment:** model-based-multi-model-parallel-20251208-230137

**Generated:** 2025-12-09 10:20:47

---

## Summary Statistics

- **Total Predictions:** 33
- **Total Observations:** 33
- **Aligned Pairs:** 32
- **Scaling Decisions:** 33

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 3/32 (9.4%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 0.11 ms
- Abs Mean Error: 0.84 ms
- Std Dev: 2.07 ms

### TTFT Predictions

- Mean Error: -0.33 ms
- Abs Mean Error: 2.11 ms
- Std Dev: 4.04 ms

## Scaling Behavior

- no-change: 28 (84.8%)
- scale-up: 3 (9.1%)
- scale-down: 2 (6.1%)

- Replica Range: 1 - 4

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
