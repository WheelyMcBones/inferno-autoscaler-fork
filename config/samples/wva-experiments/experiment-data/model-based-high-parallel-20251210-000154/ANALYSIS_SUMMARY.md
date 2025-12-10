# WVA Model-Based Experiment Analysis

**Experiment:** model-based-high-parallel-20251210-000154

**Generated:** 2025-12-10 00:26:12

---

## Summary Statistics

- **Total Predictions:** 26
- **Total Observations:** 26
- **Aligned Pairs:** 25
- **Scaling Decisions:** 26

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 13/25 (52.0%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 8.62 ms
- Abs Mean Error: 12.03 ms
- Std Dev: 17.74 ms

### TTFT Predictions

- Mean Error: 317.88 ms
- Abs Mean Error: 325.07 ms
- Std Dev: 1539.32 ms

## Scaling Behavior

- scale-up: 12 (46.2%)
- no-change: 8 (30.8%)
- scale-down: 6 (23.1%)

- Replica Range: 1 - 7

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
