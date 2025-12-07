#!/bin/bash
#
# Interactive Experiment Launcher (New Configuration-Driven System)
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_DIR="$SCRIPT_DIR/experiment-configs"
DATA_DIR="$SCRIPT_DIR/../experiment-data"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

clear

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                                                                ║${NC}"
echo -e "${BLUE}║     ${BOLD}HPA Scaling Experiments with TTFT/ITL Metrics${NC}${BLUE}          ║${NC}"
echo -e "${BLUE}║     ${CYAN}Configuration-Driven Experiment System${NC}${BLUE}                 ║${NC}"
echo -e "${BLUE}║                                                                ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Function to print menu header
print_header() {
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
}

# Function to check prerequisites
check_prerequisites() {
    local missing=0
    
    echo -e "${YELLOW}Checking prerequisites...${NC}"
    
    if ! command -v kubectl &> /dev/null; then
        echo -e "${RED}✗ kubectl not found${NC}"
        missing=1
    else
        echo -e "${GREEN}✓ kubectl found${NC}"
    fi
    
    if ! command -v yq &> /dev/null; then
        echo -e "${RED}✗ yq not found (install: brew install yq)${NC}"
        missing=1
    else
        echo -e "${GREEN}✓ yq found${NC}"
    fi
    
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}✗ jq not found${NC}"
        missing=1
    else
        echo -e "${GREEN}✓ jq found${NC}"
    fi
    
    if ! oc whoami -t &> /dev/null; then
        echo -e "${YELLOW}⚠ OpenShift token not available (Prometheus queries may fail)${NC}"
    else
        echo -e "${GREEN}✓ OpenShift token available${NC}"
    fi
    
    echo ""
    
    if [[ $missing -eq 1 ]]; then
        return 1
    fi
    return 0
}

# Function to list available experiments
list_experiments() {
    print_header "Available Experiment Configurations"
    echo ""
    
    if [[ ! -d "$CONFIG_DIR" ]] || [[ -z "$(ls -A "$CONFIG_DIR"/*.yaml 2>/dev/null)" ]]; then
        echo -e "${YELLOW}No experiment configurations found in $CONFIG_DIR${NC}"
        return
    fi
    
    local i=1
    for config in "$CONFIG_DIR"/*.yaml; do
        local name=$(yq e '.name' "$config" 2>/dev/null || echo "unknown")
        local desc=$(yq e '.description' "$config" 2>/dev/null || echo "No description")
        # Support both 'jobs' and 'workloads' array
        local jobs=$(yq e '.jobs | length' "$config" 2>/dev/null || echo "0")
        if [[ "$jobs" == "0" ]] || [[ "$jobs" == "null" ]]; then
            jobs=$(yq e '.workloads | length' "$config" 2>/dev/null || echo "0")
        fi
        
        echo -e "${GREEN}[$i]${NC} ${BOLD}$(basename "$config")${NC}"
        echo -e "    Name: $name"
        echo -e "    Description: $desc"
        echo -e "    Job Phases: $jobs"
        echo ""
        i=$((i+1))
    done
}

# Function to list past experiments
list_past_experiments() {
    print_header "Recent Experiment Results"
    echo ""
    
    if [[ ! -d "$DATA_DIR" ]] || [[ -z "$(ls -A "$DATA_DIR" 2>/dev/null)" ]]; then
        echo -e "${YELLOW}No experiment data found${NC}"
        return
    fi
    
    echo -e "${BOLD}Last 10 experiments:${NC}"
    echo ""
    
    ls -dt "$DATA_DIR"/*/ 2>/dev/null | head -10 | while read -r dir; do
        local dirname=$(basename "$dir")
        local config_file="$dir/experiment-config.yaml"
        
        if [[ -f "$config_file" ]]; then
            local exp_name=$(yq e '.name' "$config_file" 2>/dev/null || echo "unknown")
            echo -e "  ${GREEN}•${NC} $dirname - $exp_name"
        else
            echo -e "  ${GREEN}•${NC} $dirname"
        fi
    done
    echo ""
}

# Function to run experiment
run_experiment() {
    local config_file="$1"
    
    if [[ ! -f "$config_file" ]]; then
        echo -e "${RED}Configuration file not found: $config_file${NC}"
        return 1
    fi
    
    echo -e "${CYAN}Running experiment: $(basename "$config_file")${NC}"
    echo ""
    
    bash "$SCRIPT_DIR/run-experiment.sh" "$config_file"
}

# Function to view experiment results
view_results() {
    local exp_dir="$1"
    
    if [[ ! -d "$exp_dir" ]]; then
        echo -e "${RED}Experiment directory not found: $exp_dir${NC}"
        return 1
    fi
    
    print_header "Experiment Results: $(basename "$exp_dir")"
    echo ""
    
    # Show README if exists
    if [[ -f "$exp_dir/README.md" ]]; then
        cat "$exp_dir/README.md"
        echo ""
    fi
    
    # Show quick stats from metrics.csv
    if [[ -f "$exp_dir/metrics.csv" ]]; then
        echo -e "${BOLD}Quick Statistics:${NC}"
        echo ""
        
        # Count rows (excluding header)
        local samples=$(wc -l < "$exp_dir/metrics.csv")
        samples=$((samples - 1))
        echo -e "  Total samples: $samples"
        
        # Show scaling events if log exists
        if [[ -f "$exp_dir/scaling-events.log" ]]; then
            local events=$(grep -c "SCALE" "$exp_dir/scaling-events.log" 2>/dev/null || echo "0")
            echo -e "  Scaling events: $events"
        fi
        
        echo ""
        echo -e "${BOLD}First 5 samples:${NC}"
        head -6 "$exp_dir/metrics.csv" | column -t -s,
        echo ""
        
        echo -e "${BOLD}Last 5 samples:${NC}"
        tail -5 "$exp_dir/metrics.csv" | column -t -s,
        echo ""
    fi
}

# Function to analyze results
analyze_results() {
    local exp_dir="$1"
    
    if [[ ! -f "$exp_dir/metrics.csv" ]]; then
        echo -e "${RED}No metrics.csv found in $exp_dir${NC}"
        return 1
    fi
    
    echo -e "${CYAN}Opening Jupyter notebook for analysis...${NC}"
    echo ""
    
    # Check if jupyter is available
    if command -v jupyter &> /dev/null; then
        cd "$SCRIPT_DIR"
        jupyter notebook analyze-hpa-experiment.ipynb
    else
        echo -e "${YELLOW}Jupyter not found. Install with: pip install jupyter${NC}"
        echo ""
        echo -e "Alternatively, use Python script:"
        echo -e "  python legacy-scripts/analyze-hpa-experiment.py $exp_dir/metrics.csv"
    fi
}

# Main menu
show_menu() {
    while true; do
        echo ""
        print_header "Main Menu"
        echo ""
        echo -e "  ${GREEN}1${NC}) Run new experiment"
        echo -e "  ${GREEN}2${NC}) View past experiments"
        echo -e "  ${GREEN}3${NC}) Analyze experiment results"
        echo -e "  ${GREEN}4${NC}) Check system status"
        echo -e "  ${GREEN}5${NC}) Open documentation"
        echo -e "  ${RED}q${NC}) Quit"
        echo ""
        echo -ne "${BOLD}Select option:${NC} "
        read -r choice
        
        case $choice in
            1)
                # Run experiment
                echo ""
                list_experiments
                echo -e "${BOLD}Enter configuration file name or number:${NC} "
                read -r exp_choice
                
                # Check if it's a number
                if [[ "$exp_choice" =~ ^[0-9]+$ ]]; then
                    # Select by number
                    config_file=$(ls "$CONFIG_DIR"/*.yaml 2>/dev/null | sed -n "${exp_choice}p")
                else
                    # Select by filename
                    if [[ ! "$exp_choice" =~ \.yaml$ ]]; then
                        exp_choice="${exp_choice}.yaml"
                    fi
                    config_file="$CONFIG_DIR/$exp_choice"
                fi
                
                if [[ -f "$config_file" ]]; then
                    run_experiment "$config_file"
                else
                    echo -e "${RED}Invalid selection${NC}"
                fi
                ;;
            2)
                # View past experiments
                echo ""
                list_past_experiments
                echo ""
                echo -e "${BOLD}Enter experiment directory name to view details (or press Enter to skip):${NC} "
                read -r exp_dir_name
                
                if [[ -n "$exp_dir_name" ]]; then
                    view_results "$DATA_DIR/$exp_dir_name"
                fi
                ;;
            3)
                # Analyze results
                echo ""
                list_past_experiments
                echo ""
                echo -e "${BOLD}Enter experiment directory name to analyze:${NC} "
                read -r exp_dir_name
                
                if [[ -n "$exp_dir_name" ]]; then
                    analyze_results "$DATA_DIR/$exp_dir_name"
                fi
                ;;
            4)
                # Check status
                echo ""
                print_header "System Status"
                echo ""
                
                echo -e "${BOLD}Namespace:${NC} llm-d-inference-scheduler"
                echo ""
                
                echo -e "${BOLD}vLLM Deployment:${NC}"
                kubectl get deploy ms-inference-scheduling-llm-d-modelservice-decode -n llm-d-inference-scheduler 2>/dev/null || echo "Not found"
                echo ""
                
                echo -e "${BOLD}vLLM Service:${NC}"
                kubectl get svc vllm-service -n llm-d-inference-scheduler 2>/dev/null || echo "Not found"
                echo ""
                
                echo -e "${BOLD}HPA Status:${NC}"
                kubectl get hpa vllm-hpa-combined -n llm-d-inference-scheduler 2>/dev/null || echo "Not deployed"
                echo ""
                
                echo -e "${BOLD}Active Jobs:${NC}"
                kubectl get jobs -n llm-d-inference-scheduler -l experiment=sharegpt-e2e 2>/dev/null || echo "No active jobs"
                echo ""
                ;;
            5)
                # Open documentation
                echo ""
                print_header "Documentation"
                echo ""
                echo -e "Available documentation files:"
                echo -e "  ${GREEN}1${NC}) QUICKSTART.md - Quick start guide"
                echo -e "  ${GREEN}2${NC}) EXPERIMENT_SETUP.md - Complete setup guide"
                echo -e "  ${GREEN}3${NC}) SETUP_SUMMARY.md - What's new summary"
                echo -e "  ${GREEN}4${NC}) README.md - Main README"
                echo ""
                echo -e "${BOLD}Enter number to view (or press Enter to skip):${NC} "
                read -r doc_choice
                
                case $doc_choice in
                    1) less "$SCRIPT_DIR/../QUICKSTART.md" 2>/dev/null || echo "File not found" ;;
                    2) less "$SCRIPT_DIR/../EXPERIMENT_SETUP.md" 2>/dev/null || echo "File not found" ;;
                    3) less "$SCRIPT_DIR/../SETUP_SUMMARY.md" 2>/dev/null || echo "File not found" ;;
                    4) less "$SCRIPT_DIR/../README.md" 2>/dev/null || echo "File not found" ;;
                    *) ;;
                esac
                ;;
            q|Q)
                echo ""
                echo -e "${CYAN}Goodbye!${NC}"
                exit 0
                ;;
            *)
                echo -e "${RED}Invalid option${NC}"
                ;;
        esac
    done
}

# Main execution
if ! check_prerequisites; then
    echo -e "${RED}Please install missing dependencies and try again${NC}"
    exit 1
fi

show_menu
