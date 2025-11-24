#!/bin/bash

# Cleanup script to stop any running monitor processes

echo "=========================================="
echo "HPA Experiment Cleanup"
echo "=========================================="
echo ""

# Find and kill monitor processes
MONITOR_PIDS=$(pgrep -f "monitor-hpa-experiment.sh" 2>/dev/null)

if [ -z "$MONITOR_PIDS" ]; then
    echo "✓ No monitor processes found"
else
    echo "Found monitor processes:"
    ps -p $MONITOR_PIDS -o pid,command 2>/dev/null || true
    echo ""
    echo "Stopping monitor processes..."
    pkill -SIGINT -f "monitor-hpa-experiment.sh" 2>/dev/null || true
    sleep 2
    
    # Force kill if still running
    if pgrep -f "monitor-hpa-experiment.sh" > /dev/null 2>&1; then
        echo "Force killing remaining processes..."
        pkill -9 -f "monitor-hpa-experiment.sh" 2>/dev/null || true
    fi
    echo "✓ Monitor processes stopped"
fi

echo ""

# Clean up any experiment jobs
echo "Cleaning up experiment jobs..."
kubectl delete jobs -n llm-d-inference-scheduler -l experiment=sharegpt-e2e --ignore-not-found 2>/dev/null || true
echo "✓ Jobs cleaned up"

echo ""
echo "Cleanup complete!"
