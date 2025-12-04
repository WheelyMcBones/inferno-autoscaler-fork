#!/bin/bash
#
# WVA Experiment Runner
# Runs WVA experiments based on YAML configuration files
#

set -euo pipefail

# Ensure common binary paths are available
export PATH="/usr/local/bin:/opt/homebrew/bin:$PATH"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_DIR="$SCRIPT_DIR/experiment-configs"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_section() {
    echo -e "${CYAN}>>> $1${NC}"
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
for cmd in kubectl yq jq; do
    if ! command -v $cmd >/dev/null 2>&1; then
        print_error "$cmd not found. Please install it first."
        exit 1
    fi
done

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

print_header "Loading WVA Experiment Configuration"
print_info "Config file: $CONFIG_FILE"
echo ""

# Parse configuration
EXP_NAME=$(yq e '.name' "$CONFIG_FILE")
EXP_DESC=$(yq e '.description' "$CONFIG_FILE")
EXP_MODE=$(yq e '.mode' "$CONFIG_FILE")
NAMESPACE=$(yq e '.namespace' "$CONFIG_FILE")
CONTROLLER_NS=$(yq e '.controller_namespace' "$CONFIG_FILE")
CONTROLLER_PREFIX=$(yq e '.controller_pod_prefix' "$CONFIG_FILE")
DEPLOYMENT=$(yq e '.deployment' "$CONFIG_FILE")
MODEL_NAME=$(yq e '.model_name' "$CONFIG_FILE")
METRICS_INTERVAL=$(yq e '.metrics.interval' "$CONFIG_FILE")
OUTPUT_BASE_DIR=$(yq e '.output.base_dir' "$CONFIG_FILE")

# Create experiment output directory
EXPERIMENT_DIR="$OUTPUT_BASE_DIR/${EXP_NAME}-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$EXPERIMENT_DIR"

# Copy config to output directory
cp "$CONFIG_FILE" "$EXPERIMENT_DIR/experiment-config.yaml"

print_info "Experiment: $EXP_NAME"
print_info "Description: $EXP_DESC"
print_info "Mode: $EXP_MODE"
print_info "Namespace: $NAMESPACE"
print_info "Controller Namespace: $CONTROLLER_NS"
print_info "Deployment: $DEPLOYMENT"
print_info "Model: $MODEL_NAME"
print_info "Output Directory: $EXPERIMENT_DIR"
echo ""

# Verify WVA mode matches configuration
print_section "Verifying WVA Mode"
POD=$(kubectl get pods -n "$CONTROLLER_NS" -o name | grep "$CONTROLLER_PREFIX" | head -n1 | sed 's/pod\///')
if [[ -z "$POD" ]]; then
    print_error "No WVA controller pod found with prefix: $CONTROLLER_PREFIX"
    exit 1
fi

print_info "Controller pod: $POD"

# Get recent logs to check mode
RECENT_LOGS=$(kubectl logs -n "$CONTROLLER_NS" "$POD" --since=2m --tail=50 2>/dev/null || true)
if echo "$RECENT_LOGS" | grep -q "Operating in MODEL-ONLY mode"; then
    ACTUAL_MODE="model-based"
elif echo "$RECENT_LOGS" | grep -q "Operating in CAPACITY-ONLY mode"; then
    ACTUAL_MODE="capacity-based"
elif echo "$RECENT_LOGS" | grep -q "Operating in HYBRID mode"; then
    ACTUAL_MODE="hybrid"
else
    print_warn "Could not detect WVA mode from logs. Proceeding anyway..."
    ACTUAL_MODE="unknown"
fi

if [[ "$ACTUAL_MODE" != "unknown" ]]; then
    print_info "Detected WVA mode: $ACTUAL_MODE"
    
    if [[ "$EXP_MODE" != "$ACTUAL_MODE" ]]; then
        print_error "Mode mismatch!"
        print_error "  Expected: $EXP_MODE"
        print_error "  Actual: $ACTUAL_MODE"
        echo ""
        print_warn "To change WVA mode:"
        if [[ "$EXP_MODE" == "model-based" ]]; then
            echo "  kubectl set env deployment/$CONTROLLER_PREFIX EXPERIMENTAL_HYBRID_OPTIMIZATION=model-only -n $CONTROLLER_NS"
        elif [[ "$EXP_MODE" == "capacity-based" ]]; then
            echo "  kubectl set env deployment/$CONTROLLER_PREFIX EXPERIMENTAL_HYBRID_OPTIMIZATION=off -n $CONTROLLER_NS"
        fi
        echo ""
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        print_info "✓ WVA mode matches configuration"
    fi
fi
echo ""

# Verify deployment exists
print_section "Verifying Deployment"
if ! kubectl get deployment "$DEPLOYMENT" -n "$NAMESPACE" >/dev/null 2>&1; then
    print_error "Deployment not found: $DEPLOYMENT in namespace $NAMESPACE"
    exit 1
fi

REPLICAS=$(kubectl get deployment "$DEPLOYMENT" -n "$NAMESPACE" -o jsonpath='{.status.replicas}')
print_info "Deployment: $DEPLOYMENT"
print_info "Current replicas: $REPLICAS"
echo ""

# Initialize job tracking arrays (used by cleanup trap)
declare -a JOB_NAMES_ARRAY=()
declare -a JOB_PIDS=()

# Start WVA log collection in background
print_section "Starting Log Collection"
LOG_FILE="$EXPERIMENT_DIR/wva-controller-logs.jsonl"
MONITOR_SCRIPT="$SCRIPT_DIR/monitor-wva.sh"

chmod +x "$MONITOR_SCRIPT"

print_info "Launching log collector..."
"$MONITOR_SCRIPT" "$CONTROLLER_NS" "$CONTROLLER_PREFIX" "$LOG_FILE" "$METRICS_INTERVAL" > "$EXPERIMENT_DIR/monitor.log" 2>&1 &
MONITOR_PID=$!

print_info "Monitor PID: $MONITOR_PID"
print_info "Log file: $LOG_FILE"
print_info "Monitor log: $EXPERIMENT_DIR/monitor.log"

# Wait a bit for monitor to start
sleep 3

if ! kill -0 $MONITOR_PID 2>/dev/null; then
    print_error "Monitor process died. Check $EXPERIMENT_DIR/monitor.log"
    exit 1
fi

print_info "✓ Log collection started"
echo ""

# Trap to cleanup on exit
cleanup() {
    print_section "Cleaning Up"
    
    # Kill all background job processes (subshells) and their children
    if [[ ${#JOB_PIDS[@]} -gt 0 ]]; then
        print_info "Terminating background job processes..."
        for pid in "${JOB_PIDS[@]}"; do
            if kill -0 "$pid" 2>/dev/null; then
                # Kill the process group (this gets the subshell and all its children)
                kill -TERM -"$pid" 2>/dev/null || true
                # Also try to kill the PID directly
                kill -TERM "$pid" 2>/dev/null || true
            fi
        done
        # Wait briefly for graceful termination
        sleep 1
        # Force kill any remaining processes
        for pid in "${JOB_PIDS[@]}"; do
            if kill -0 "$pid" 2>/dev/null; then
                kill -KILL -"$pid" 2>/dev/null || true
                kill -KILL "$pid" 2>/dev/null || true
            fi
        done
    fi
    
    # Stop any running Kubernetes jobs
    if [[ ${#JOB_NAMES_ARRAY[@]} -gt 0 ]]; then
        print_info "Deleting Kubernetes jobs..."
        for job_name in "${JOB_NAMES_ARRAY[@]}"; do
            kubectl delete job "$job_name" -n "$NAMESPACE" --ignore-not-found=true &>/dev/null || true
        done
    fi
    
    # Stop log collector
    if [[ -n "${MONITOR_PID:-}" ]] && kill -0 $MONITOR_PID 2>/dev/null; then
        print_info "Stopping log collector (PID: $MONITOR_PID)..."
        kill $MONITOR_PID 2>/dev/null || true
        wait $MONITOR_PID 2>/dev/null || true
    fi
    
    # Parse logs to CSV
    print_info "Parsing logs to CSV..."
    if [[ -f "$LOG_FILE" ]] && [[ -s "$LOG_FILE" ]]; then
        parse_logs_to_csv "$LOG_FILE" "$EXPERIMENT_DIR/metrics.csv" "$EXP_MODE"
    fi
    
    print_info "Experiment complete!"
    print_info "Data saved to: $EXPERIMENT_DIR"
}

trap cleanup EXIT INT TERM

# Function to parse logs to CSV based on mode
parse_logs_to_csv() {
    local log_file="$1"
    local csv_file="$2"
    local mode="$3"
    
    if [[ "$mode" == "model-based" ]]; then
        # Parse model-based logs
        cat > "$EXPERIMENT_DIR/parse_model_based.py" << 'EOF'
import json
import sys
import csv
import re

def parse_optimization_solution(solution_str):
    """Extract predicted ITL and TTFT from optimization solution."""
    itl_match = re.search(r'itl=([0-9.]+)', solution_str)
    ttft_match = re.search(r'ttft=([0-9.]+)', solution_str)
    replicas_match = re.search(r'numRep=([0-9]+)', solution_str)
    
    return {
        'predicted_itl': float(itl_match.group(1)) if itl_match else None,
        'predicted_ttft': float(ttft_match.group(1)) if ttft_match else None,
        'predicted_replicas': int(replicas_match.group(1)) if replicas_match else None
    }

log_file = sys.argv[1]
csv_file = sys.argv[2]

events = []
with open(log_file, 'r') as f:
    for line in f:
        try:
            log = json.loads(line.strip())
            ts = log.get('ts', '')
            msg = log.get('msg', '')
            level = log.get('level', '')
            
            # Extract different event types
            if 'Optimization solution' in msg:
                prediction = parse_optimization_solution(msg)
                events.append({
                    'timestamp': ts,
                    'event_type': 'prediction',
                    **prediction
                })
            elif 'Processing decision' in msg:
                # Extract scaling decision
                current_match = re.search(r'current=([0-9]+)', msg)
                target_match = re.search(r'target=([0-9]+)', msg)
                action_match = re.search(r'action=([a-z-]+)', msg)
                
                events.append({
                    'timestamp': ts,
                    'event_type': 'scaling_decision',
                    'current_replicas': int(current_match.group(1)) if current_match else None,
                    'target_replicas': int(target_match.group(1)) if target_match else None,
                    'action': action_match.group(1) if action_match else None
                })
            elif 'Found SLO for model' in msg:
                # Extract SLO values
                slo_itl_match = re.search(r'slo-tpot=([0-9]+)', msg)
                slo_ttft_match = re.search(r'slo-ttft=([0-9]+)', msg)
                
                events.append({
                    'timestamp': ts,
                    'event_type': 'slo',
                    'slo_itl': int(slo_itl_match.group(1)) if slo_itl_match else None,
                    'slo_ttft': int(slo_ttft_match.group(1)) if slo_ttft_match else None
                })
            elif 'Metrics collected for VA' in msg or 'observed' in msg.lower():
                # Try to extract observed metrics (these come from Prometheus)
                replicas_match = re.search(r'replicas=([0-9]+)', msg)
                ttft_match = re.search(r'ttft=([0-9.]+)', msg)
                itl_match = re.search(r'itl=([0-9.]+)', msg)
                
                if any([replicas_match, ttft_match, itl_match]):
                    events.append({
                        'timestamp': ts,
                        'event_type': 'observed_metrics',
                        'observed_replicas': int(replicas_match.group(1)) if replicas_match else None,
                        'observed_ttft': float(ttft_match.group(1)) if ttft_match else None,
                        'observed_itl': float(itl_match.group(1)) if itl_match else None
                    })
                    
        except json.JSONDecodeError:
            continue
        except Exception as e:
            print(f"Error parsing line: {e}", file=sys.stderr)
            continue

# Write to CSV
if events:
    fieldnames = sorted(set().union(*(d.keys() for d in events)))
    with open(csv_file, 'w', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(events)
    print(f"Parsed {len(events)} events to {csv_file}")
else:
    print("No events found in logs", file=sys.stderr)
EOF
        python3 "$EXPERIMENT_DIR/parse_model_based.py" "$log_file" "$csv_file"
        
    elif [[ "$mode" == "capacity-based" ]]; then
        # Parse capacity-based logs
        cat > "$EXPERIMENT_DIR/parse_capacity_based.py" << 'EOF'
import json
import sys
import csv
import re

log_file = sys.argv[1]
csv_file = sys.argv[2]

events = []
with open(log_file, 'r') as f:
    for line in f:
        try:
            log = json.loads(line.strip())
            ts = log.get('ts', '')
            msg = log.get('msg', '')
            
            # KV cache metrics
            if 'KV cache metric' in msg:
                pod_match = re.search(r'pod=([a-z0-9-]+)', msg)
                usage_match = re.search(r'usage=([0-9.]+)', msg)
                percent_match = re.search(r'\(([0-9.]+)%\)', msg)
                
                events.append({
                    'timestamp': ts,
                    'event_type': 'kv_cache',
                    'pod': pod_match.group(1) if pod_match else None,
                    'kv_cache_usage': float(usage_match.group(1)) if usage_match else None,
                    'kv_cache_percent': float(percent_match.group(1)) if percent_match else None
                })
            
            # Queue metrics
            elif 'Queue metric' in msg:
                pod_match = re.search(r'pod=([a-z0-9-]+)', msg)
                queue_match = re.search(r'queueLength=([0-9]+)', msg)
                
                events.append({
                    'timestamp': ts,
                    'event_type': 'queue',
                    'pod': pod_match.group(1) if pod_match else None,
                    'queue_length': int(queue_match.group(1)) if queue_match else None
                })
            
            # Metrics collected
            elif 'Metrics collected for VA' in msg:
                replicas_match = re.search(r'replicas=([0-9]+)', msg)
                ttft_match = re.search(r'ttft=([0-9.]+)', msg)
                itl_match = re.search(r'itl=([0-9.]+)', msg)
                cost_match = re.search(r'cost=([0-9.]+)', msg)
                
                events.append({
                    'timestamp': ts,
                    'event_type': 'metrics',
                    'replicas': int(replicas_match.group(1)) if replicas_match else None,
                    'ttft': float(ttft_match.group(1)) if ttft_match else None,
                    'itl': float(itl_match.group(1)) if itl_match else None,
                    'cost': float(cost_match.group(1)) if cost_match else None
                })
            
            # Capacity analysis
            elif 'Capacity analysis completed' in msg:
                total_match = re.search(r'totalReplicas=([0-9]+)', msg)
                nonsaturated_match = re.search(r'nonSaturated=([0-9]+)', msg)
                scaleup_match = re.search(r'shouldScaleUp=([a-z]+)', msg)
                scaledown_match = re.search(r'scaleDownSafe=([a-z]+)', msg)
                
                events.append({
                    'timestamp': ts,
                    'event_type': 'capacity_analysis',
                    'total_replicas': int(total_match.group(1)) if total_match else None,
                    'non_saturated_replicas': int(nonsaturated_match.group(1)) if nonsaturated_match else None,
                    'should_scale_up': scaleup_match.group(1) == 'true' if scaleup_match else None,
                    'scale_down_safe': scaledown_match.group(1) == 'true' if scaledown_match else None
                })
            
            # Scaling decisions
            elif 'Processing decision' in msg:
                current_match = re.search(r'current=([0-9]+)', msg)
                target_match = re.search(r'target=([0-9]+)', msg)
                action_match = re.search(r'action=([a-z-]+)', msg)
                
                events.append({
                    'timestamp': ts,
                    'event_type': 'scaling_decision',
                    'current_replicas': int(current_match.group(1)) if current_match else None,
                    'target_replicas': int(target_match.group(1)) if target_match else None,
                    'action': action_match.group(1) if action_match else None
                })
                    
        except json.JSONDecodeError:
            continue
        except Exception as e:
            print(f"Error parsing line: {e}", file=sys.stderr)
            continue

# Write to CSV
if events:
    fieldnames = sorted(set().union(*(d.keys() for d in events)))
    with open(csv_file, 'w', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(events)
    print(f"Parsed {len(events)} events to {csv_file}")
else:
    print("No events found in logs", file=sys.stderr)
EOF
        python3 "$EXPERIMENT_DIR/parse_capacity_based.py" "$log_file" "$csv_file"
    fi
}

# Run workloads
print_section "Running Workload Sequence"
WORKLOAD_COUNT=$(yq e '.workloads | length' "$CONFIG_FILE")
print_info "Total workloads: $WORKLOAD_COUNT"
echo ""

# Check if any workloads have start_delay (parallel mode)
HAS_START_DELAYS=false
for ((i=0; i<WORKLOAD_COUNT; i++)); do
    START_DELAY=$(yq e ".workloads[$i].start_delay // 0" "$CONFIG_FILE")
    if [[ "$START_DELAY" != "0" ]]; then
        HAS_START_DELAYS=true
        break
    fi
done

if [[ "$HAS_START_DELAYS" == "true" ]]; then
    # PARALLEL MODE: Launch jobs with staggered start times
    print_info "🔀 Parallel mode enabled (workloads have start_delay configured)"
    echo ""
    
    # Calculate total experiment duration (max of start_delay + duration)
    TOTAL_DURATION=0
    for ((i=0; i<WORKLOAD_COUNT; i++)); do
        START_DELAY=$(yq e ".workloads[$i].start_delay // 0" "$CONFIG_FILE")
        WORKLOAD_DURATION=$(yq e ".workloads[$i].duration" "$CONFIG_FILE")
        END_TIME=$((START_DELAY + WORKLOAD_DURATION))
        if [[ $END_TIME -gt $TOTAL_DURATION ]]; then
            TOTAL_DURATION=$END_TIME
        fi
    done
    
    print_info "Total experiment duration: ${TOTAL_DURATION}s ($(($TOTAL_DURATION / 60))m $(($TOTAL_DURATION % 60))s)"
    echo ""
    
    # Record experiment start time
    EXPERIMENT_START=$(date +%s)
    
    # Launch all jobs in background with delays
    for ((i=0; i<WORKLOAD_COUNT; i++)); do
        WORKLOAD_NAME=$(yq e ".workloads[$i].name" "$CONFIG_FILE")
        JOB_MANIFEST=$(yq e ".workloads[$i].job_manifest" "$CONFIG_FILE")
        START_DELAY=$(yq e ".workloads[$i].start_delay // 0" "$CONFIG_FILE")
        WORKLOAD_DURATION=$(yq e ".workloads[$i].duration" "$CONFIG_FILE")
        
        # Resolve job manifest path
        JOB_PATH="$(dirname "$CONFIG_FILE")/$JOB_MANIFEST"
        if [[ ! -f "$JOB_PATH" ]]; then
            print_warn "Job manifest not found, skipping: $JOB_PATH"
            continue
        fi
        
        # Extract actual job name from manifest
        ACTUAL_JOB_NAME=$(yq e '.metadata.name' "$JOB_PATH")
        JOB_NAMES_ARRAY+=("$ACTUAL_JOB_NAME")
        
        # Launch job in background
        (
            # Wait for start delay
            if [[ $START_DELAY -gt 0 ]]; then
                sleep "$START_DELAY"
            fi
            
            # Deploy job
            ELAPSED=$(($(date +%s) - EXPERIMENT_START))
            echo "[T+${ELAPSED}s] ▶ Starting job: $WORKLOAD_NAME (duration: ${WORKLOAD_DURATION}s)"
            kubectl apply -f "$JOB_PATH" -n "$NAMESPACE" &>/dev/null
            
            # Wait for job duration
            sleep "$WORKLOAD_DURATION"
            
            ELAPSED=$(($(date +%s) - EXPERIMENT_START))
            echo "[T+${ELAPSED}s] ✓ Job completed: $WORKLOAD_NAME"
            
            # Clean up job
            kubectl delete -f "$JOB_PATH" -n "$NAMESPACE" --ignore-not-found=true &>/dev/null
        ) &
        JOB_PIDS+=($!)
        
        print_info "[$((i+1))/$WORKLOAD_COUNT] Scheduled: $WORKLOAD_NAME (starts at T+${START_DELAY}s, runs for ${WORKLOAD_DURATION}s)"
    done
    
    echo ""
    print_info "All jobs scheduled. Monitoring experiment progress..."
    echo ""
    
    # Monitor experiment progress
    for ((t=0; t<=TOTAL_DURATION; t+=10)); do
        ELAPSED=$t
        REMAINING=$((TOTAL_DURATION - t))
        
        # Show active jobs at this time
        ACTIVE_JOBS=""
        for ((i=0; i<WORKLOAD_COUNT; i++)); do
            START_DELAY=$(yq e ".workloads[$i].start_delay // 0" "$CONFIG_FILE")
            WORKLOAD_DURATION=$(yq e ".workloads[$i].duration" "$CONFIG_FILE")
            END_TIME=$((START_DELAY + WORKLOAD_DURATION))
            
            if [[ $t -ge $START_DELAY ]] && [[ $t -lt $END_TIME ]]; then
                WORKLOAD_NAME=$(yq e ".workloads[$i].name" "$CONFIG_FILE")
                if [[ -n "$ACTIVE_JOBS" ]]; then
                    ACTIVE_JOBS="$ACTIVE_JOBS, $WORKLOAD_NAME"
                else
                    ACTIVE_JOBS="$WORKLOAD_NAME"
                fi
            fi
        done
        
        if [[ -n "$ACTIVE_JOBS" ]]; then
            echo "[T+${ELAPSED}s] Running: $ACTIVE_JOBS (${REMAINING}s remaining)"
        else
            echo "[T+${ELAPSED}s] No active jobs (${REMAINING}s remaining)"
        fi
        
        # Don't sleep on the last iteration
        if [[ $t -lt $TOTAL_DURATION ]]; then
            sleep 10
        fi
    done
    
    echo ""
    print_info "Waiting for all job processes to finish..."
    for pid in "${JOB_PIDS[@]}"; do
        wait "$pid" 2>/dev/null || true
    done
    
    # Final cleanup - make sure all jobs are deleted
    print_info "Cleaning up jobs..."
    for job_name in "${JOB_NAMES_ARRAY[@]}"; do
        kubectl delete job "$job_name" -n "$NAMESPACE" --ignore-not-found=true &>/dev/null || true
    done
    
else
    # SEQUENTIAL MODE: Original behavior
    print_info "📋 Sequential mode (workloads run one after another)"
    echo ""
    
    for ((i=0; i<WORKLOAD_COUNT; i++)); do
        WORKLOAD_NAME=$(yq e ".workloads[$i].name" "$CONFIG_FILE")
        WORKLOAD_DESC=$(yq e ".workloads[$i].description" "$CONFIG_FILE")
        JOB_MANIFEST=$(yq e ".workloads[$i].job_manifest" "$CONFIG_FILE")
        WAIT_COMPLETION=$(yq e ".workloads[$i].wait_completion" "$CONFIG_FILE")
        
        # Resolve job manifest path relative to config file
        JOB_PATH="$(dirname "$CONFIG_FILE")/$JOB_MANIFEST"
        
        if [[ ! -f "$JOB_PATH" ]]; then
            print_error "Job manifest not found: $JOB_PATH"
            continue
        fi
        
        print_header "Workload $((i+1))/$WORKLOAD_COUNT: $WORKLOAD_NAME"
        print_info "Description: $WORKLOAD_DESC"
        print_info "Manifest: $JOB_MANIFEST"
        echo ""
        
        # Apply job
        print_info "Launching job..."
        # Extract job name from manifest
        JOB_NAME=$(yq e '.metadata.name' "$JOB_PATH")
        JOB_NAMES_ARRAY+=("$JOB_NAME")
        
        kubectl apply -f "$JOB_PATH" -n "$NAMESPACE"
        
        if [[ "$WAIT_COMPLETION" == "true" ]]; then
            print_info "Waiting for job completion: $JOB_NAME"
            kubectl wait --for=condition=complete --timeout=30m "job/$JOB_NAME" -n "$NAMESPACE" || {
                print_warn "Job did not complete within timeout"
            }
            
            print_info "✓ Job completed: $JOB_NAME"
            
            # Delete job after completion
            print_info "Cleaning up job..."
            kubectl delete job "$JOB_NAME" -n "$NAMESPACE" --ignore-not-found=true
            # Remove from tracking array since it's cleaned up
            JOB_NAMES_ARRAY=("${JOB_NAMES_ARRAY[@]/$JOB_NAME}")
        else
            print_info "Job launched (not waiting for completion)"
        fi
        
        echo ""
    done
fi

print_header "Experiment Complete"
print_info "Collecting final logs for 120 seconds..."
sleep 120

# Cleanup will be called via trap
