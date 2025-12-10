# WVA Model-Based Experiment Analysis

**Experiment:** model-based-multi-model-parallel-20251210-003455

**Generated:** 2025-12-10 11:07:23

---

## Summary Statistics

- **Total Predictions:** 38
- **Total Observations:** 38
- **Aligned Pairs:** 37
- **Scaling Decisions:** 38

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 10/37 (27.0%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 1.01 ms
- Abs Mean Error: 4.10 ms
- Std Dev: 8.33 ms

### TTFT Predictions

- Mean Error: 0.99 ms
- Abs Mean Error: 7.38 ms
- Std Dev: 13.39 ms

## Scaling Behavior

- no-change: 27 (71.1%)
- scale-up: 8 (21.1%)
- scale-down: 3 (7.9%)

- Replica Range: 1 - 3

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
