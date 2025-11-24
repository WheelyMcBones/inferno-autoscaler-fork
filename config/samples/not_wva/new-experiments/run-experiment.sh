#!/bin/bash
#
# Configurable HPA Experiment Runner
# Runs experiments based on YAML configuration files
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_DIR="$SCRIPT_DIR/experiment-configs"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check dependencies
command -v kubectl >/dev/null 2>&1 || { print_error "kubectl not found"; exit 1; }
command -v yq >/dev/null 2>&1 || { print_error "yq not found. Install with: brew install yq"; exit 1; }
command -v jq >/dev/null 2>&1 || { print_error "jq not found"; exit 1; }

# Parse arguments
if [[ $# -eq 0 ]]; then
    print_error "Usage: $0 <config-file.yaml>"
    echo ""
    echo "Available configurations:"
    ls -1 "$CONFIG_DIR"/*.yaml 2>/dev/null | xargs -n1 basename || echo "  No configurations found"
    exit 1
fi

CONFIG_FILE="$1"

# Resolve config file path
if [[ ! -f "$CONFIG_FILE" ]]; then
    if [[ -f "$CONFIG_DIR/$CONFIG_FILE" ]]; then
        CONFIG_FILE="$CONFIG_DIR/$CONFIG_FILE"
    else
        print_error "Configuration file not found: $CONFIG_FILE"
        exit 1
    fi
fi

print_header "Loading Experiment Configuration"
print_info "Config file: $CONFIG_FILE"
echo ""

# Parse configuration
EXP_NAME=$(yq e '.name' "$CONFIG_FILE")
EXP_DESC=$(yq e '.description' "$CONFIG_FILE")
NAMESPACE=$(yq e '.namespace' "$CONFIG_FILE")
DEPLOYMENT=$(yq e '.deployment' "$CONFIG_FILE")
MODEL_NAME=$(yq e '.model_name' "$CONFIG_FILE")
HPA_ENABLED=$(yq e '.hpa.enabled' "$CONFIG_FILE")
HPA_MANIFEST=$(yq e '.hpa.manifest' "$CONFIG_FILE")
METRICS_INTERVAL=$(yq e '.metrics.interval' "$CONFIG_FILE")
OUTPUT_BASE_DIR=$(yq e '.output.base_dir' "$CONFIG_FILE")

# Create experiment output directory
EXPERIMENT_DIR="$OUTPUT_BASE_DIR/${EXP_NAME}-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$EXPERIMENT_DIR"

# Copy config to output directory
cp "$CONFIG_FILE" "$EXPERIMENT_DIR/experiment-config.yaml"

print_info "Experiment: $EXP_NAME"
print_info "Description: $EXP_DESC"
print_info "Namespace: $NAMESPACE"
print_info "Deployment: $DEPLOYMENT"
print_info "Model: $MODEL_NAME"
print_info "HPA Enabled: $HPA_ENABLED"
print_info "Output Directory: $EXPERIMENT_DIR"
echo ""

# Deploy HPA if enabled
if [[ "$HPA_ENABLED" == "true" ]]; then
    print_header "Deploying HPA"
    
    # Resolve HPA manifest path relative to config file
    HPA_PATH="$(dirname "$CONFIG_FILE")/$HPA_MANIFEST"
    if [[ ! -f "$HPA_PATH" ]]; then
        print_error "HPA manifest not found: $HPA_PATH"
        exit 1
    fi
    
    print_info "Applying HPA: $HPA_PATH"
    kubectl apply -f "$HPA_PATH"
    
    HPA_NAME=$(yq e '.metadata.name' "$HPA_PATH")
    print_info "HPA deployed: $HPA_NAME"
    sleep 5
    echo ""
fi

# Start monitoring in background
print_header "Starting Metrics Collection"

MONITOR_SCRIPT="$SCRIPT_DIR/monitor-hpa-enhanced.sh"
export NAMESPACE
export DEPLOYMENT
export HPA_NAME="${HPA_NAME:-vllm-hpa-combined}"
export MODEL_NAME
export INTERVAL="$METRICS_INTERVAL"
export OUTPUT_DIR="$EXPERIMENT_DIR"

bash "$MONITOR_SCRIPT" &
MONITOR_PID=$!
print_info "Monitor started (PID: $MONITOR_PID)"
sleep 3
echo ""

# Cleanup function
cleanup() {
    print_header "Experiment Cleanup"
    
    # Stop monitor
    if kill -0 $MONITOR_PID 2>/dev/null; then
        print_info "Stopping monitor (PID: $MONITOR_PID)..."
        kill -SIGTERM $MONITOR_PID 2>/dev/null || true
        wait $MONITOR_PID 2>/dev/null || true
    fi
    
    # Clean up jobs
    print_info "Cleaning up jobs..."
    kubectl delete jobs -n "$NAMESPACE" -l experiment=sharegpt-e2e --ignore-not-found=true
    
    # Generate summary
    print_info "Generating experiment summary..."
    cat > "$EXPERIMENT_DIR/README.md" <<EOF
# Experiment: $EXP_NAME

**Date**: $(date '+%Y-%m-%d %H:%M:%S')
**Description**: $EXP_DESC

## Configuration
- Namespace: $NAMESPACE
- Deployment: $DEPLOYMENT
- Model: $MODEL_NAME
- HPA Enabled: $HPA_ENABLED

## Files
- \`metrics.csv\`: Time-series metrics data (replicas, TTFT, ITL, etc.)
- \`scaling-events.log\`: HPA scaling events log
- \`experiment-config.yaml\`: Full experiment configuration

## Metrics Collected
- Replicas (current and desired)
- Num Requests Waiting
- KV Cache Usage (%)
- TTFT (Time to First Token, ms)
- ITL (Inter-Token Latency, ms)
- Request Rate (req/min)
- Active Jobs

## Analysis
Use the analysis notebook or script to visualize results:
\`\`\`bash
cd ../..
python scripts/analyze-hpa-experiment.py $EXPERIMENT_DIR/metrics.csv
# or
jupyter notebook analyze-hpa-experiment.ipynb
\`\`\`
EOF
    
    print_info "Experiment complete!"
    print_info "Results saved to: $EXPERIMENT_DIR"
    echo ""
}

trap cleanup EXIT INT TERM

# Execute job phases
JOB_COUNT=$(yq e '.jobs | length' "$CONFIG_FILE")
print_header "Executing $JOB_COUNT Job Phases"
echo ""

for i in $(seq 0 $((JOB_COUNT - 1))); do
    JOB_NAME=$(yq e ".jobs[$i].name" "$CONFIG_FILE")
    JOB_MANIFEST=$(yq e ".jobs[$i].manifest" "$CONFIG_FILE")
    JOB_DURATION=$(yq e ".jobs[$i].duration" "$CONFIG_FILE")
    
    # Resolve job manifest path
    JOB_PATH="$(dirname "$CONFIG_FILE")/$JOB_MANIFEST"
    if [[ ! -f "$JOB_PATH" ]]; then
        print_warn "Job manifest not found, skipping: $JOB_PATH"
        continue
    fi
    
    print_info "[$((i+1))/$JOB_COUNT] Starting job phase: $JOB_NAME"
    print_info "  Manifest: $JOB_PATH"
    print_info "  Duration: ${JOB_DURATION}s"
    
    # Deploy job
    kubectl apply -f "$JOB_PATH"
    
    # Wait for job to start
    sleep 10
    
    # Wait for phase duration
    print_info "  Waiting ${JOB_DURATION}s for job phase to complete..."
    
    # Progress bar
    for ((t=0; t<JOB_DURATION; t+=5)); do
        remaining=$((JOB_DURATION - t))
        printf "\r  Progress: %d/%d seconds (remaining: %ds)  " "$t" "$JOB_DURATION" "$remaining"
        sleep 5
    done
    printf "\n"
    
    # Clean up this job before next phase
    kubectl delete job -n "$NAMESPACE" -l experiment=sharegpt-e2e --ignore-not-found=true
    print_info "  Job phase completed"
    echo ""
    
    # Wait between phases
    if [[ $i -lt $((JOB_COUNT - 1)) ]]; then
        print_info "Waiting 30s before next phase..."
        sleep 30
    fi
done

print_header "All Job Phases Complete"
print_info "Experiment finished successfully"
echo ""

# Cleanup will run via trap
exit 0
