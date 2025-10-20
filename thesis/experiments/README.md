# WVA Experiment Results

This directory contains organized results from WVA (Workload-Variant-Autoscaler) performance analysis experiments.

## Directory Structure

Each experiment run creates a timestamped directory with the following structure:

```
experiments/
└── YYYYMMDD_HHMMSS/          # Experiment timestamp
    ├── manifest.txt           # Experiment metadata and configuration
    ├── data/                  # Raw and processed data files
    │   ├── extracted_metrics.csv    # Metrics extracted from logs
    │   ├── processed_metrics.csv    # Processed data with calculations
    │   └── original_log.txt         # Copy of original WVA log file
    ├── plots/                 # Generated visualizations (PNG, 300 DPI)
    │   ├── itl_analysis.png              # ITL performance with warmup gaps
    │   ├── ttft_analysis.png             # TTFT performance with warmup gaps
    │   ├── load_pattern.png              # Arrival rate evolution
    │   ├── itl_replicas_timeline.png     # ITL vs replica scaling
    │   ├── ttft_replicas_timeline.png    # TTFT vs replica scaling
    │   └── combined_itl_ttft.png         # Combined ITL & TTFT analysis
    └── analysis/              # Statistical analysis results
        ├── summary.txt               # Text summary of key findings
        ├── warmup_gaps.csv           # Detected warmup gap statistics
        └── scaling_events.csv        # Scaling event log
```

## File Descriptions

### Metadata
- **manifest.txt**: Experiment configuration, timestamp, duration, and directory structure

### Data Files
- **extracted_metrics.csv**: Raw metrics extracted from WVA controller logs
  - Columns: timestamp, itlAverage, ttftAverage, rate, inTk, outTk, numRep, itl, ttft, slo_itl, slo_ttft
- **processed_metrics.csv**: Enhanced data with time calculations and derived fields
- **original_log.txt**: Backup copy of the original WVA controller log file

### Plots (6 Visualizations)
1. **itl_analysis.png**: Inter-Token Latency (ITL) performance over time
   - Shows actual vs predicted ITL
   - Highlights warmup gaps (SLO violations)
   - Marks scaling events
   
2. **ttft_analysis.png**: Time to First Token (TTFT) performance over time
   - Shows actual vs predicted TTFT
   - Highlights warmup gaps
   - Marks scaling events

3. **load_pattern.png**: Request arrival rate evolution
   - Requests per minute (rpm) over time
   - Scaling event annotations

4. **itl_replicas_timeline.png**: Dual y-axis plot
   - Left axis: ITL metrics
   - Right axis: Number of replicas
   - Shows correlation between scaling and performance

5. **ttft_replicas_timeline.png**: Dual y-axis plot
   - Left axis: TTFT metrics
   - Right axis: Number of replicas

6. **combined_itl_ttft.png**: Stacked subplots
   - Top: ITL performance
   - Bottom: TTFT performance
   - Synchronized time axis for easy comparison

### Analysis Files
- **summary.txt**: Statistical summary including:
  - Experiment duration and data points
  - ITL and TTFT compliance rates
  - Peak load and scaling pattern
  
- **warmup_gaps.csv**: Detected periods where performance exceeded SLO
  - Start/end times, duration
  - Peak violations, average performance
  - Average load during gap

- **scaling_events.csv**: Log of all replica scaling events
  - Timestamp of scaling
  - From/to replica counts
  - Load at time of scaling

## Usage

### Running a New Experiment

1. Open `wva_analysis.ipynb`
2. Update the `LOG_FILE` variable in cell 1 to point to your WVA controller logs
3. Run all cells (Cell → Run All)
4. Results will be automatically saved to a new timestamped directory

### Comparing Experiments

Compare different experiments by examining their respective directories:

```bash
# List all experiments
ls -la experiments/

# Compare summaries
diff experiments/20251020_143022/analysis/summary.txt \
     experiments/20251020_150315/analysis/summary.txt

# View plots side by side
open experiments/20251020_143022/plots/itl_analysis.png \
     experiments/20251020_150315/plots/itl_analysis.png
```

## Metrics Glossary

- **ITL (Inter-Token Latency)**: Time between consecutive output tokens (ms)
- **TTFT (Time to First Token)**: Time from request to first token (ms)
- **SLO (Service Level Objective)**: Target performance threshold
- **Warmup Gap**: Period where performance exceeds SLO during pod initialization
- **rpm**: Requests per minute (arrival rate)

## Notes

- All plots are saved at 300 DPI for publication quality
- Timestamps are in ISO 8601 format (YYYY-MM-DD HH:MM:SS)
- CSV files use comma separators with headers
- Each experiment is self-contained and independent
