# WVA Model-Based Experiment Analysis

**Experiment:** model-based-high-parallel-20251209-101938

**Generated:** 2025-12-09 10:56:13

---

## Summary Statistics

- **Total Predictions:** 24
- **Total Observations:** 24
- **Aligned Pairs:** 23
- **Scaling Decisions:** 24

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 10/23 (43.5%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 8.60 ms
- Abs Mean Error: 11.40 ms
- Std Dev: 17.79 ms

### TTFT Predictions

- Mean Error: 26.95 ms
- Abs Mean Error: 34.27 ms
- Std Dev: 84.12 ms

## Scaling Behavior

- no-change: 10 (41.7%)
- scale-up: 10 (41.7%)
- scale-down: 4 (16.7%)

- Replica Range: 1 - 6

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
