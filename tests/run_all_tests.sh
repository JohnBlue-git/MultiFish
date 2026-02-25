#!/bin/bash

# run_all_tests.sh - Run all tests in the multifish project
# Usage: ./run_all_tests.sh [options]
# Options:
#   -v, --verbose    Enable verbose output
#   -h, --html       Generate HTML report
#   -c, --coverage   Generate coverage report

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default options
VERBOSE=false
HTML_REPORT=false
COVERAGE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -h|--html)
            HTML_REPORT=true
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  -v, --verbose    Enable verbose output"
            echo "  -h, --html       Generate HTML report"
            echo "  -c, --coverage   Generate coverage report"
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
echo -e "${BLUE}Running All Tests for Multifish Project${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

cd "$PROJECT_ROOT"

# Prepare test flags
TEST_FLAGS="-race -timeout 30s"
if [ "$VERBOSE" = true ]; then
    TEST_FLAGS="$TEST_FLAGS -v"
fi

# Run tests with coverage if requested
if [ "$COVERAGE" = true ]; then
    echo -e "${YELLOW}Running tests with coverage...${NC}"
    COVERAGE_FILE="$REPORTS_DIR/coverage_${TIMESTAMP}.out"
    
    if [ "$VERBOSE" = true ]; then
        go test $TEST_FLAGS -coverprofile="$COVERAGE_FILE" -covermode=atomic ./... | tee "$REPORTS_DIR/test_output_${TIMESTAMP}.log"
    else
        go test $TEST_FLAGS -coverprofile="$COVERAGE_FILE" -covermode=atomic ./... 2>&1 | tee "$REPORTS_DIR/test_output_${TIMESTAMP}.log"
    fi
    
    TEST_EXIT_CODE=${PIPESTATUS[0]}
    
    if [ $TEST_EXIT_CODE -eq 0 ]; then
        echo -e "${GREEN}✓ All tests passed!${NC}"
        
        # Display coverage summary
        echo ""
        echo -e "${YELLOW}Coverage Summary:${NC}"
        go tool cover -func="$COVERAGE_FILE" | tail -n 1
        
        # Generate HTML coverage report if requested
        if [ "$HTML_REPORT" = true ]; then
            HTML_FILE="$REPORTS_DIR/coverage_${TIMESTAMP}.html"
            go tool cover -html="$COVERAGE_FILE" -o "$HTML_FILE"
            echo -e "${GREEN}✓ HTML coverage report generated: $HTML_FILE${NC}"
        fi
        
        echo -e "${BLUE}Coverage file saved: $COVERAGE_FILE${NC}"
    else
        echo -e "${RED}✗ Some tests failed!${NC}"
        exit $TEST_EXIT_CODE
    fi
else
    # Run tests without coverage
    echo -e "${YELLOW}Running tests...${NC}"
    
    if [ "$VERBOSE" = true ]; then
        go test $TEST_FLAGS ./... 2>&1 | tee "$REPORTS_DIR/test_output_${TIMESTAMP}.log"
    else
        go test $TEST_FLAGS ./... 2>&1 | tee "$REPORTS_DIR/test_output_${TIMESTAMP}.log"
    fi
    
    TEST_EXIT_CODE=${PIPESTATUS[0]}
    
    if [ $TEST_EXIT_CODE -eq 0 ]; then
        echo -e "${GREEN}✓ All tests passed!${NC}"
    else
        echo -e "${RED}✗ Some tests failed!${NC}"
        exit $TEST_EXIT_CODE
    fi
fi

echo ""
echo -e "${BLUE}Test output saved: $REPORTS_DIR/test_output_${TIMESTAMP}.log${NC}"
echo -e "${BLUE}========================================${NC}"
