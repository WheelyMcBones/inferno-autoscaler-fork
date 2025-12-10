# WVA Model-Based Experiment Analysis

**Experiment:** model-based-high-parallel-20251210-105831

**Generated:** 2025-12-10 11:51:29

---

## Summary Statistics

- **Total Predictions:** 28
- **Total Observations:** 28
- **Aligned Pairs:** 27
- **Scaling Decisions:** 28

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 13/27 (48.1%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 14.15 ms
- Abs Mean Error: 17.80 ms
- Std Dev: 25.22 ms

### TTFT Predictions

- Mean Error: 738.30 ms
- Abs Mean Error: 745.99 ms
- Std Dev: 2442.98 ms

## Scaling Behavior

- scale-up: 13 (46.4%)
- no-change: 10 (35.7%)
- scale-down: 5 (17.9%)

- Replica Range: 1 - 6

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
