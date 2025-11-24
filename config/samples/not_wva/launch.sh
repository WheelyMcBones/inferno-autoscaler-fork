#!/bin/bash
#
# HPA Experiment Launcher - Main Entry Point
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

clear

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                                                                ║${NC}"
echo -e "${BLUE}║          ${BOLD}HPA Scaling Experiments - Launcher${NC}${BLUE}                   ║${NC}"
echo -e "${BLUE}║                                                                ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${CYAN}This directory contains two experiment systems:${NC}"
echo ""
echo -e "${GREEN}1) NEW SYSTEM (Recommended)${NC} - Configuration-driven experiments"
echo -e "   ${BOLD}Location:${NC} new-experiments/"
echo -e "   ${BOLD}Features:${NC}"
echo -e "     • YAML-based experiment configurations"
echo -e "     • Collects TTFT and ITL metrics (WVA-compatible)"
echo -e "     • Multi-phase job sequencing"
echo -e "     • Enhanced monitoring"
echo -e "     • Interactive menu interface"
echo ""
echo -e "${YELLOW}2) LEGACY SYSTEM${NC} - Original manual experiment runner"
echo -e "   ${BOLD}Location:${NC} legacy-scripts/"
echo -e "   ${BOLD}Features:${NC}"
echo -e "     • Manual job deployment"
echo -e "     • Basic HPA metrics"
echo -e "     • Simple monitoring"
echo ""
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${BOLD}Select experiment system:${NC}"
echo ""
echo -e "  ${GREEN}1${NC}) New configuration-driven system (Recommended)"
echo -e "  ${YELLOW}2${NC}) Legacy manual system"
echo -e "  ${CYAN}3${NC}) View documentation"
echo -e "  q) Quit"
echo ""
echo -ne "${BOLD}Your choice:${NC} "
read -r choice

case $choice in
    1)
        echo ""
        echo -e "${GREEN}Starting new experiment system...${NC}"
        echo ""
        exec bash "$SCRIPT_DIR/new-experiments/start-experiments.sh"
        ;;
    2)
        echo ""
        echo -e "${YELLOW}Legacy System${NC}"
        echo ""
        echo "Available scripts in legacy-scripts/:"
        echo "  • run-hpa-experiment.sh - Run experiments manually"
        echo "  • monitor-hpa-experiment.sh - Basic monitoring"
        echo "  • analyze-hpa-experiment.py - Analysis script"
        echo "  • view-experiment.sh - View results"
        echo ""
        echo "To use, run scripts from legacy-scripts/ directory"
        ;;
    3)
        echo ""
        echo -e "${CYAN}Documentation${NC}"
        echo ""
        echo "Quick Start:"
        echo "  less QUICKSTART.md"
        echo ""
        echo "Complete Setup Guide:"
        echo "  less EXPERIMENT_SETUP.md"
        echo ""
        echo "Summary of Changes:"
        echo "  less SETUP_SUMMARY.md"
        echo ""
        echo "Main README:"
        echo "  less README.md"
        ;;
    q|Q)
        echo ""
        echo "Goodbye!"
        exit 0
        ;;
    *)
        echo ""
        echo -e "${YELLOW}Invalid choice${NC}"
        exit 1
        ;;
esac
