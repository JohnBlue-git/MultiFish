#!/bin/bash

# coverage_report.sh - Generate comprehensive coverage reports for the multifish project
# Usage: ./coverage_report.sh [options]
# Options:
#   -h, --html       Generate HTML coverage report (default: true)
#   -t, --text       Display text coverage report in terminal
#   -p, --packages   Show per-package coverage breakdown
#   --threshold N    Fail if coverage is below N% (e.g., --threshold 70)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default options
HTML_REPORT=true
TEXT_REPORT=false
PACKAGE_BREAKDOWN=false
COVERAGE_THRESHOLD=0

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--html)
            HTML_REPORT=true
            shift
            ;;
        -t|--text)
            TEXT_REPORT=true
            shift
            ;;
        -p|--packages)
            PACKAGE_BREAKDOWN=true
            shift
            ;;
        --threshold)
            COVERAGE_THRESHOLD="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  -h, --html       Generate HTML coverage report (default: true)"
            echo "  -t, --text       Display text coverage report in terminal"
            echo "  -p, --packages   Show per-package coverage breakdown"
            echo "  --threshold N    Fail if coverage is below N% (e.g., --threshold 70)"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Get the project root directory (parent of tests folder)
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REPORTS_DIR="$SCRIPT_DIR/reports"

# Create reports directory if it doesn't exist
mkdir -p "$REPORTS_DIR"

# Timestamp for report files
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Generating Coverage Report${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

cd "$PROJECT_ROOT"

# Run tests with coverage
COVERAGE_FILE="$REPORTS_DIR/coverage_${TIMESTAMP}.out"
echo -e "${YELLOW}Running tests to collect coverage data...${NC}"

go test -race -timeout 30s -coverprofile="$COVERAGE_FILE" -covermode=atomic ./... > /dev/null 2>&1

if [ $? -ne 0 ]; then
    echo -e "${RED}✗ Tests failed! Cannot generate coverage report.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Tests completed successfully${NC}"
echo ""

# Calculate total coverage percentage
TOTAL_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}' | sed 's/%//')

echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}COVERAGE SUMMARY${NC}"
echo -e "${CYAN}========================================${NC}"
echo -e "${YELLOW}Total Coverage:${NC} ${GREEN}${TOTAL_COVERAGE}%${NC}"
echo ""

# Check coverage threshold
if (( $(echo "$TOTAL_COVERAGE < $COVERAGE_THRESHOLD" | bc -l) )); then
    echo -e "${RED}✗ Coverage ${TOTAL_COVERAGE}% is below threshold ${COVERAGE_THRESHOLD}%${NC}"
    THRESHOLD_FAILED=true
else
    if [ "$COVERAGE_THRESHOLD" -gt 0 ]; then
        echo -e "${GREEN}✓ Coverage meets threshold of ${COVERAGE_THRESHOLD}%${NC}"
        echo ""
    fi
    THRESHOLD_FAILED=false
fi

# Display per-package breakdown
if [ "$PACKAGE_BREAKDOWN" = true ]; then
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}PER-PACKAGE COVERAGE BREAKDOWN${NC}"
    echo -e "${CYAN}========================================${NC}"
    
    # Get unique packages
    PACKAGES=$(go tool cover -func="$COVERAGE_FILE" | grep -v "total:" | awk '{print $1}' | sed 's/:[^:]*$//' | sort -u)
    
    while IFS= read -r package; do
        if [ -n "$package" ]; then
            # Calculate coverage for this package
            PACKAGE_COV=$(go tool cover -func="$COVERAGE_FILE" | grep "^${package}:" | awk '{sum+=$3; count++} END {if(count>0) printf "%.1f", sum/count; else print "0.0"}')
            
            # Color code based on coverage level
            if (( $(echo "$PACKAGE_COV >= 80" | bc -l) )); then
                COLOR=$GREEN
            elif (( $(echo "$PACKAGE_COV >= 60" | bc -l) )); then
                COLOR=$YELLOW
            else
                COLOR=$RED
            fi
            
            printf "${COLOR}%-60s %6s%%${NC}\n" "$package" "$PACKAGE_COV"
        fi
    done <<< "$PACKAGES"
    echo ""
fi

# Display detailed text report
if [ "$TEXT_REPORT" = true ]; then
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}DETAILED COVERAGE REPORT${NC}"
    echo -e "${CYAN}========================================${NC}"
    go tool cover -func="$COVERAGE_FILE"
    echo ""
fi

# Generate HTML report
if [ "$HTML_REPORT" = true ]; then
    HTML_FILE="$REPORTS_DIR/coverage_${TIMESTAMP}.html"
    go tool cover -html="$COVERAGE_FILE" -o "$HTML_FILE"
    echo -e "${GREEN}✓ HTML coverage report generated:${NC}"
    echo -e "  ${BLUE}$HTML_FILE${NC}"
    echo ""
    echo -e "${YELLOW}To view the HTML report:${NC}"
    echo -e "  xdg-open $HTML_FILE"
    echo -e "  or open it in your browser"
    echo ""
fi

# Save coverage data summary to a text file
SUMMARY_FILE="$REPORTS_DIR/coverage_summary_${TIMESTAMP}.txt"
{
    echo "Coverage Report - $(date)"
    echo "========================================"
    echo "Total Coverage: ${TOTAL_COVERAGE}%"
    echo ""
    echo "Coverage by Package:"
    echo "----------------------------------------"
    
    PACKAGES=$(go tool cover -func="$COVERAGE_FILE" | grep -v "total:" | awk '{print $1}' | sed 's/:[^:]*$//' | sort -u)
    
    while IFS= read -r package; do
        if [ -n "$package" ]; then
            PACKAGE_COV=$(go tool cover -func="$COVERAGE_FILE" | grep "^${package}:" | awk '{sum+=$3; count++} END {if(count>0) printf "%.1f", sum/count; else print "0.0"}')
            printf "%-60s %6s%%\n" "$package" "$PACKAGE_COV"
        fi
    done <<< "$PACKAGES"
    
    echo ""
    echo "========================================"
    echo "Detailed Function Coverage:"
    echo "========================================"
    go tool cover -func="$COVERAGE_FILE"
} > "$SUMMARY_FILE"

echo -e "${GREEN}✓ Coverage summary saved:${NC}"
echo -e "  ${BLUE}$SUMMARY_FILE${NC}"
echo ""

echo -e "${BLUE}Coverage data file:${NC}"
echo -e "  ${BLUE}$COVERAGE_FILE${NC}"
echo ""

echo -e "${BLUE}========================================${NC}"

# Exit with error if threshold check failed
if [ "$THRESHOLD_FAILED" = true ]; then
    exit 1
fi
