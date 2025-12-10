# WVA Model-Based Experiment Analysis

**Experiment:** model-based-multi-model-parallel-20251209-001450

**Generated:** 2025-12-10 01:02:10

---

## Summary Statistics

- **Total Predictions:** 33
- **Total Observations:** 33
- **Aligned Pairs:** 32
- **Scaling Decisions:** 33

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 10/32 (31.2%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 0.94 ms
- Abs Mean Error: 1.71 ms
- Std Dev: 3.47 ms

### TTFT Predictions

- Mean Error: 1.11 ms
- Abs Mean Error: 2.87 ms
- Std Dev: 5.81 ms

## Scaling Behavior

- no-change: 21 (63.6%)
- scale-up: 10 (30.3%)
- scale-down: 2 (6.1%)

- Replica Range: 1 - 3

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
