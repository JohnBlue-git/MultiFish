#!/bin/bash

# test_summary.sh - Generate a comprehensive test summary dashboard
# Usage: ./test_summary.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Get the project root directory (parent of tests folder)
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REPORTS_DIR="$SCRIPT_DIR/reports"

# Create reports directory if it doesn't exist
mkdir -p "$REPORTS_DIR"

cd "$PROJECT_ROOT"

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║${NC}          ${CYAN}MULTIFISH PROJECT TEST SUMMARY DASHBOARD${NC}            ${BLUE}║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Count test files and functions
echo -e "${YELLOW}📊 Test Statistics${NC}"
echo -e "${BLUE}────────────────────────────────────────────────────────────────${NC}"

TEST_FILES=$(find . -name "*_test.go" -type f | wc -l)
echo -e "${CYAN}Total Test Files:${NC} $TEST_FILES"

# Count test functions
TEST_FUNCTIONS=$(grep -r "^func Test" --include="*_test.go" . 2>/dev/null | wc -l)
echo -e "${CYAN}Total Test Functions:${NC} $TEST_FUNCTIONS"

# Count benchmark functions
BENCHMARK_FUNCTIONS=$(grep -r "^func Benchmark" --include="*_test.go" . 2>/dev/null | wc -l)
echo -e "${CYAN}Total Benchmark Functions:${NC} $BENCHMARK_FUNCTIONS"

echo ""

# List test files by package
echo -e "${YELLOW}📁 Test Files by Package${NC}"
echo -e "${BLUE}────────────────────────────────────────────────────────────────${NC}"

find . -name "*_test.go" -type f | sed 's|^\./||' | while read -r file; do
    dir=$(dirname "$file")
    if [ "$dir" = "." ]; then
        dir="root"
    fi
    echo -e "  ${GREEN}•${NC} $file"
done | sort

echo ""

# Run tests and capture results
echo -e "${YELLOW}🧪 Running Tests${NC}"
echo -e "${BLUE}────────────────────────────────────────────────────────────────${NC}"

TEMP_OUTPUT=$(mktemp)
TEMP_COVERAGE=$(mktemp)

if go test -race -timeout 30s -coverprofile="$TEMP_COVERAGE" -covermode=atomic ./... > "$TEMP_OUTPUT" 2>&1; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    TEST_STATUS="PASSED"
else
    echo -e "${RED}✗ Some tests failed!${NC}"
    TEST_STATUS="FAILED"
    
    # Show failed tests
    echo ""
    echo -e "${RED}Failed Tests:${NC}"
    grep -E "FAIL|--- FAIL" "$TEMP_OUTPUT" || echo "No detailed failure information available"
fi

echo ""

# Package-level test results
echo -e "${YELLOW}📦 Package Test Results${NC}"
echo -e "${BLUE}────────────────────────────────────────────────────────────────${NC}"

# Extract package results from test output
grep -E "^(ok|FAIL)" "$TEMP_OUTPUT" | while read -r line; do
    if echo "$line" | grep -q "^ok"; then
        package=$(echo "$line" | awk '{print $2}')
        time=$(echo "$line" | awk '{print $3}')
        echo -e "  ${GREEN}✓${NC} $package ${CYAN}($time)${NC}"
    else
        package=$(echo "$line" | awk '{print $2}')
        echo -e "  ${RED}✗${NC} $package"
    fi
done

echo ""

# Coverage summary
if [ -f "$TEMP_COVERAGE" ]; then
    echo -e "${YELLOW}📈 Coverage Summary${NC}"
    echo -e "${BLUE}────────────────────────────────────────────────────────────────${NC}"
    
    TOTAL_COVERAGE=$(go tool cover -func="$TEMP_COVERAGE" | grep total | awk '{print $3}' | sed 's/%//')
    
    # Color code coverage
    if (( $(echo "$TOTAL_COVERAGE >= 80" | bc -l) )); then
        COV_COLOR=$GREEN
    elif (( $(echo "$TOTAL_COVERAGE >= 60" | bc -l) )); then
        COV_COLOR=$YELLOW
    else
        COV_COLOR=$RED
    fi
    
    echo -e "${CYAN}Total Coverage:${NC} ${COV_COLOR}${TOTAL_COVERAGE}%${NC}"
    echo ""
    
    # Per-package coverage
    echo -e "${CYAN}Coverage by Package:${NC}"
    
    PACKAGES=$(go tool cover -func="$TEMP_COVERAGE" | grep -v "total:" | awk '{print $1}' | sed 's/:[^:]*$//' | sort -u)
    
    while IFS= read -r package; do
        if [ -n "$package" ]; then
            PACKAGE_COV=$(go tool cover -func="$TEMP_COVERAGE" | grep "^${package}:" | awk '{sum+=$3; count++} END {if(count>0) printf "%.1f", sum/count; else print "0.0"}')
            
            # Color code based on coverage level
            if (( $(echo "$PACKAGE_COV >= 80" | bc -l) )); then
                COLOR=$GREEN
            elif (( $(echo "$PACKAGE_COV >= 60" | bc -l) )); then
                COLOR=$YELLOW
            else
                COLOR=$RED
            fi
            
            printf "  ${COLOR}%-50s %6s%%${NC}\n" "$package" "$PACKAGE_COV"
        fi
    done <<< "$PACKAGES"
    
    echo ""
fi

# Recent test history
echo -e "${YELLOW}📜 Recent Test Reports${NC}"
echo -e "${BLUE}────────────────────────────────────────────────────────────────${NC}"

if [ -d "$REPORTS_DIR" ] && [ "$(ls -A "$REPORTS_DIR" 2>/dev/null)" ]; then
    echo -e "${CYAN}Latest 5 test reports:${NC}"
    ls -lt "$REPORTS_DIR"/test_output_*.log 2>/dev/null | head -5 | awk '{print $6, $7, $8, $9}' | while read -r line; do
        echo -e "  ${GREEN}•${NC} $line"
    done
    
    echo ""
    echo -e "${CYAN}Latest 5 coverage reports:${NC}"
    ls -lt "$REPORTS_DIR"/coverage_*.html 2>/dev/null | head -5 | awk '{print $6, $7, $8, $9}' | while read -r line; do
        echo -e "  ${GREEN}•${NC} $line"
    done
else
    echo -e "  ${YELLOW}No previous reports found${NC}"
fi

echo ""

# Summary
echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║${NC}                        ${CYAN}SUMMARY${NC}                              ${BLUE}║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"

if [ "$TEST_STATUS" = "PASSED" ]; then
    echo -e "${GREEN}Status: ✓ ALL TESTS PASSED${NC}"
else
    echo -e "${RED}Status: ✗ TESTS FAILED${NC}"
fi

echo -e "${CYAN}Test Files:${NC} $TEST_FILES"
echo -e "${CYAN}Test Functions:${NC} $TEST_FUNCTIONS"
if [ -f "$TEMP_COVERAGE" ]; then
    echo -e "${CYAN}Coverage:${NC} ${COV_COLOR}${TOTAL_COVERAGE}%${NC}"
fi
echo -e "${CYAN}Generated:${NC} $(date '+%Y-%m-%d %H:%M:%S')"

echo ""
echo -e "${BLUE}────────────────────────────────────────────────────────────────${NC}"

# Clean up
rm -f "$TEMP_OUTPUT" "$TEMP_COVERAGE"

# Exit with appropriate code
if [ "$TEST_STATUS" = "FAILED" ]; then
    exit 1
fi
