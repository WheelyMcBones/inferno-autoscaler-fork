#!/bin/bash

# Quick Results Viewer
# Shows latest experiment results in a readable format

EXPERIMENT_DIR="$1"

if [ -z "$EXPERIMENT_DIR" ]; then
    echo "Usage: $0 <experiment-directory>"
    echo ""
    echo "Available experiments:"
    ls -1dt ./experiment-data/hpa-experiment-* 2>/dev/null | head -5
    exit 1
fi

if [ ! -d "$EXPERIMENT_DIR" ]; then
    echo "Error: Directory not found: $EXPERIMENT_DIR"
    exit 1
fi

echo "=========================================="
echo "EXPERIMENT QUICK VIEW"
echo "=========================================="
echo ""

# Show metadata
if [ -f "$EXPERIMENT_DIR/metadata.json" ]; then
    echo "📊 METADATA:"
    jq -r '. | "  Name: \(.experiment_name)\n  Start: \(.start_time)\n  End: \(.end_time // "In Progress")\n  Namespace: \(.namespace)"' "$EXPERIMENT_DIR/metadata.json"
    echo ""
fi

# Show metrics summary
if [ -f "$EXPERIMENT_DIR/metrics.csv" ]; then
    echo "📈 METRICS SUMMARY:"
    
    # Get line count (subtract header)
    SAMPLE_COUNT=$(($(wc -l < "$EXPERIMENT_DIR/metrics.csv") - 1))
    echo "  Samples: $SAMPLE_COUNT"
    
    # Show first and last few entries
    echo ""
    echo "  First 3 samples:"
    head -4 "$EXPERIMENT_DIR/metrics.csv" | column -t -s, | sed 's/^/    /'
    
    echo ""
    echo "  Last 3 samples:"
    tail -3 "$EXPERIMENT_DIR/metrics.csv" | column -t -s, | sed 's/^/    /'
    echo ""
fi

# Show scaling events count
if [ -f "$EXPERIMENT_DIR/scaling.log" ]; then
    SCALING_EVENTS=$(grep -c "SCALING EVENT" "$EXPERIMENT_DIR/scaling.log" || echo "0")
    echo "🔄 SCALING EVENTS: $SCALING_EVENTS"
    
    if [ "$SCALING_EVENTS" -gt 0 ]; then
        echo ""
        echo "  Events:"
        grep "SCALING EVENT" "$EXPERIMENT_DIR/scaling.log" | sed 's/^/    /' | head -10
    fi
    echo ""
fi

# Show files
echo "📁 FILES:"
ls -lh "$EXPERIMENT_DIR" | tail -n +2 | awk '{printf "  %-25s %10s\n", $9, $5}'
echo ""

echo "=========================================="
echo "TO ANALYZE:"
echo "=========================================="
echo ""
echo "View full metrics:"
echo "  cat '$EXPERIMENT_DIR/metrics.csv' | column -t -s,"
echo ""
echo "View scaling log:"
echo "  cat '$EXPERIMENT_DIR/scaling.log' | less"
echo ""
echo "Generate plot:"
echo "  python3 config/samples/not_wva/analyze-hpa-experiment.py '$EXPERIMENT_DIR' --plot results.png"
echo ""
