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
EXPERIMENT_START_TIME=$(date '+%Y-%m-%d %H:%M:%S')

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
JOBS_DELETED=false  # Track if jobs were already deleted during cooldown

cleanup() {
    print_header "Experiment Cleanup"
    
    # Stop monitor (only if not already stopped)
    if [[ -n "$MONITOR_PID" ]] && kill -0 $MONITOR_PID 2>/dev/null; then
        print_info "Stopping monitor (PID: $MONITOR_PID)..."
        kill -SIGTERM $MONITOR_PID 2>/dev/null || true
        wait $MONITOR_PID 2>/dev/null || true
    fi
    
    # Clean up jobs (only if not already deleted during cooldown)
    if [[ "$JOBS_DELETED" != "true" ]]; then
        print_info "Cleaning up jobs..."
        kubectl delete jobs -n "$NAMESPACE" -l 'experiment in (sharegpt-e2e,sharegpt-high-load)' --ignore-not-found=true 2>/dev/null || \
        kubectl delete jobs -n "$NAMESPACE" --selector='job-name' --field-selector='status.successful=1' --ignore-not-found=true 2>/dev/null || \
        kubectl delete jobs -n "$NAMESPACE" --all --ignore-not-found=true 2>/dev/null || true
        print_info "Jobs cleanup complete"
    fi
    
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
    
    # Generate metadata.json for programmatic access
    EXPERIMENT_END_TIME=$(date '+%Y-%m-%d %H:%M:%S')
    cat > "$EXPERIMENT_DIR/metadata.json" <<EOF
{
  "experiment_name": "$EXP_NAME",
  "description": "$EXP_DESC",
  "start_time": "$EXPERIMENT_START_TIME",
  "end_time": "$EXPERIMENT_END_TIME",
  "namespace": "$NAMESPACE",
  "deployment_name": "$DEPLOYMENT",
  "hpa_name": "$HPA_NAME",
  "model_name": "$MODEL_NAME",
  "hpa_enabled": $HPA_ENABLED,
  "job_count": $JOB_COUNT,
  "cooldown_enabled": $(yq e '.cooldown.enabled // false' "$CONFIG_FILE"),
  "cooldown_duration": $(yq e '.cooldown.duration // 0' "$CONFIG_FILE")
}
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

# Check if any jobs have start_delay (parallel mode)
HAS_START_DELAYS=false
for i in $(seq 0 $((JOB_COUNT - 1))); do
    START_DELAY=$(yq e ".jobs[$i].start_delay // 0" "$CONFIG_FILE")
    if [[ "$START_DELAY" != "0" ]]; then
        HAS_START_DELAYS=true
        break
    fi
done

if [[ "$HAS_START_DELAYS" == "true" ]]; then
    # PARALLEL MODE: Launch jobs with staggered start times
    print_info "Parallel mode enabled (jobs have start_delay configured)"
    echo ""
    
    # Calculate total experiment duration (max of start_delay + duration)
    TOTAL_DURATION=0
    for i in $(seq 0 $((JOB_COUNT - 1))); do
        START_DELAY=$(yq e ".jobs[$i].start_delay // 0" "$CONFIG_FILE")
        JOB_DURATION=$(yq e ".jobs[$i].duration" "$CONFIG_FILE")
        END_TIME=$((START_DELAY + JOB_DURATION))
        if [[ $END_TIME -gt $TOTAL_DURATION ]]; then
            TOTAL_DURATION=$END_TIME
        fi
    done
    
    print_info "Total experiment duration: ${TOTAL_DURATION}s"
    echo ""
    
    # Record experiment start time
    EXPERIMENT_START=$(date +%s)
    
    # Launch all jobs in background with delays
    declare -a JOB_PIDS=()
    for i in $(seq 0 $((JOB_COUNT - 1))); do
        JOB_NAME=$(yq e ".jobs[$i].name" "$CONFIG_FILE")
        JOB_MANIFEST=$(yq e ".jobs[$i].manifest" "$CONFIG_FILE")
        START_DELAY=$(yq e ".jobs[$i].start_delay // 0" "$CONFIG_FILE")
        JOB_DURATION=$(yq e ".jobs[$i].duration" "$CONFIG_FILE")
        
        # Resolve job manifest path
        JOB_PATH="$(dirname "$CONFIG_FILE")/$JOB_MANIFEST"
        if [[ ! -f "$JOB_PATH" ]]; then
            print_warn "Job manifest not found, skipping: $JOB_PATH"
            continue
        fi
        
        # Launch job in background
        (
            # Wait for start delay
            if [[ $START_DELAY -gt 0 ]]; then
                sleep "$START_DELAY"
            fi
            
            # Deploy job
            ELAPSED=$(($(date +%s) - EXPERIMENT_START))
            echo "[T+${ELAPSED}s] Starting job: $JOB_NAME (duration: ${JOB_DURATION}s)"
            kubectl apply -f "$JOB_PATH" &>/dev/null
            
            # Wait for job duration
            sleep "$JOB_DURATION"
            
            ELAPSED=$(($(date +%s) - EXPERIMENT_START))
            echo "[T+${ELAPSED}s] Job completed: $JOB_NAME"
        ) &
        JOB_PIDS+=($!)
        
        print_info "[$((i+1))/$JOB_COUNT] Scheduled: $JOB_NAME (starts at T+${START_DELAY}s, runs for ${JOB_DURATION}s)"
    done
    
    echo ""
    print_info "All jobs scheduled. Monitoring experiment progress..."
    echo ""
    
    # Monitor experiment progress
    for ((t=0; t<=TOTAL_DURATION; t+=5)); do
        ELAPSED=$t
        remaining=$((TOTAL_DURATION - t))
        
        # Check which jobs are currently running
        ACTIVE_JOBS=""
        for i in $(seq 0 $((JOB_COUNT - 1))); do
            JOB_NAME=$(yq e ".jobs[$i].name" "$CONFIG_FILE")
            START_DELAY=$(yq e ".jobs[$i].start_delay // 0" "$CONFIG_FILE")
            JOB_DURATION=$(yq e ".jobs[$i].duration" "$CONFIG_FILE")
            END_TIME=$((START_DELAY + JOB_DURATION))
            
            if [[ $t -ge $START_DELAY ]] && [[ $t -lt $END_TIME ]]; then
                if [[ -n "$ACTIVE_JOBS" ]]; then
                    ACTIVE_JOBS="$ACTIVE_JOBS, $JOB_NAME"
                else
                    ACTIVE_JOBS="$JOB_NAME"
                fi
            fi
        done
        
        if [[ -z "$ACTIVE_JOBS" ]]; then
            ACTIVE_JOBS="none"
        fi
        
        printf "\r  [T+%ds/%ds] Active jobs: %-50s (remaining: %ds)  " "$ELAPSED" "$TOTAL_DURATION" "$ACTIVE_JOBS" "$remaining"
        
        if [[ $t -lt $TOTAL_DURATION ]]; then
            sleep 5
        fi
    done
    printf "\n\n"
    
    # Wait for all background jobs to complete
    print_info "Waiting for all background job processes to finish..."
    for pid in "${JOB_PIDS[@]}"; do
        wait "$pid" 2>/dev/null || true
    done
    
else
    # SEQUENTIAL MODE: Execute jobs one at a time (original behavior)
    print_info "Sequential mode (no start_delay configured)"
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
        kubectl delete job -n "$NAMESPACE" -l 'experiment in (sharegpt-e2e,sharegpt-high-load)' --ignore-not-found=true 2>/dev/null || true
        print_info "  Job phase completed"
        echo ""
        
        # Wait between phases
        if [[ $i -lt $((JOB_COUNT - 1)) ]]; then
            print_info "Waiting 30s before next phase..."
            sleep 30
        fi
    done
fi

print_header "All Job Phases Complete"
print_info "Experiment finished successfully"
echo ""

# Check if cooldown/observation period is enabled
COOLDOWN_ENABLED=$(yq e '.cooldown.enabled // false' "$CONFIG_FILE")
COOLDOWN_DURATION=$(yq e '.cooldown.duration // 0' "$CONFIG_FILE")

if [[ "$COOLDOWN_ENABLED" == "true" ]] && [[ "$COOLDOWN_DURATION" -gt 0 ]]; then
    print_header "Scale-Down Observation Period"
    print_info "Monitoring HPA scale-down for ${COOLDOWN_DURATION}s..."
    print_info "This captures how HPA scales back to minReplicas after load completes"
    print_info "Metrics collection continues in background (PID: $MONITOR_PID)"
    echo ""
    
    # Delete jobs now to trigger scale-down
    print_info "Deleting load jobs to trigger HPA scale-down..."
    kubectl delete jobs -n "$NAMESPACE" -l 'experiment in (sharegpt-e2e,sharegpt-high-load)' --ignore-not-found=true 2>/dev/null || true
    JOBS_DELETED=true  # Mark jobs as deleted to prevent duplicate deletion in cleanup
    echo ""
    
    # Continue monitoring during cooldown
    INITIAL_REPLICAS=$(kubectl get deployment "$DEPLOYMENT" -n "$NAMESPACE" -o jsonpath='{.status.replicas}' 2>/dev/null || echo "0")
    print_info "Starting observation with $INITIAL_REPLICAS replicas..."
    echo ""
    
    for ((t=0; t<COOLDOWN_DURATION; t+=5)); do
        remaining=$((COOLDOWN_DURATION - t))
        
        # Get current replica count
        CURRENT_REPLICAS=$(kubectl get deployment "$DEPLOYMENT" -n "$NAMESPACE" -o jsonpath='{.status.replicas}' 2>/dev/null || echo "0")
        READY_REPLICAS=$(kubectl get deployment "$DEPLOYMENT" -n "$NAMESPACE" -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo "0")
        DESIRED_REPLICAS=$(kubectl get hpa "$HPA_NAME" -n "$NAMESPACE" -o jsonpath='{.status.desiredReplicas}' 2>/dev/null || echo "0")
        
        printf "\r  [Cooldown: %ds/%ds] Replicas: %s ready / %s current / %s desired (remaining: %ds)  " \
            "$t" "$COOLDOWN_DURATION" "$READY_REPLICAS" "$CURRENT_REPLICAS" "$DESIRED_REPLICAS" "$remaining"
        
        if [[ $t -lt $COOLDOWN_DURATION ]]; then
            sleep 5
        fi
    done
    printf "\n\n"
    
    FINAL_REPLICAS=$(kubectl get deployment "$DEPLOYMENT" -n "$NAMESPACE" -o jsonpath='{.status.replicas}' 2>/dev/null || echo "0")
    print_info "Scale-down observation complete: $INITIAL_REPLICAS → $FINAL_REPLICAS replicas"
    echo ""
fi

# Cleanup will run via trap
exit 0
