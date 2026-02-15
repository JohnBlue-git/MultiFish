#!/bin/bash

# run_specific_test.sh - Run specific test(s) in the multifish project
# Usage: ./run_specific_test.sh <package_path> [test_name] [options]
# Examples:
#   ./run_specific_test.sh ./scheduler
#   ./run_specific_test.sh ./scheduler TestJobExecutor
#   ./run_specific_test.sh ./utility -v -c

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
PACKAGE_PATH=""
TEST_NAME=""

# Show help
show_help() {
    echo "Usage: $0 <package_path> [test_name] [options]"
    echo ""
    echo "Arguments:"
    echo "  package_path     Path to the package (e.g., ./scheduler, ./utility)"
    echo "  test_name        (Optional) Specific test function name (e.g., TestJobExecutor)"
    echo ""
    echo "Options:"
    echo "  -v, --verbose    Enable verbose output"
    echo "  -h, --html       Generate HTML coverage report"
    echo "  -c, --coverage   Generate coverage report"
    echo "  --help           Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 ./scheduler"
    echo "  $0 ./scheduler TestJobExecutor"
    echo "  $0 ./utility -v -c"
    echo "  $0 ./scheduler TestJobExecutor -v -c -h"
}

# Parse arguments
if [ $# -eq 0 ] || [ "$1" = "--help" ]; then
    show_help
    exit 0
fi

# First argument is always the package path
PACKAGE_PATH="$1"
shift

# Check if second argument is a test name (doesn't start with -)
if [ $# -gt 0 ] && [[ ! "$1" =~ ^- ]]; then
    TEST_NAME="$1"
    shift
fi

# Parse remaining options
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -h|--html)
            HTML_REPORT=true
            COVERAGE=true  # HTML report requires coverage
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_help
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

# Clean package path for file naming
CLEAN_PATH=$(echo "$PACKAGE_PATH" | tr '/' '_' | sed 's/^[._]*//')
if [ -n "$TEST_NAME" ]; then
    REPORT_PREFIX="${CLEAN_PATH}_${TEST_NAME}"
else
    REPORT_PREFIX="${CLEAN_PATH}"
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Running Tests${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${YELLOW}Package:${NC} $PACKAGE_PATH"
if [ -n "$TEST_NAME" ]; then
    echo -e "${YELLOW}Test:${NC} $TEST_NAME"
fi
echo ""

cd "$PROJECT_ROOT"

# Prepare test flags
TEST_FLAGS="-race -timeout 30s"
if [ "$VERBOSE" = true ]; then
    TEST_FLAGS="$TEST_FLAGS -v"
fi

# Add test name filter if specified
if [ -n "$TEST_NAME" ]; then
    TEST_FLAGS="$TEST_FLAGS -run ^${TEST_NAME}$"
fi

# Run tests
if [ "$COVERAGE" = true ]; then
    echo -e "${YELLOW}Running tests with coverage...${NC}"
    COVERAGE_FILE="$REPORTS_DIR/coverage_${REPORT_PREFIX}_${TIMESTAMP}.out"
    
    go test $TEST_FLAGS -coverprofile="$COVERAGE_FILE" -covermode=atomic "$PACKAGE_PATH" 2>&1 | tee "$REPORTS_DIR/test_output_${REPORT_PREFIX}_${TIMESTAMP}.log"
    
    TEST_EXIT_CODE=${PIPESTATUS[0]}
    
    if [ $TEST_EXIT_CODE -eq 0 ]; then
        echo -e "${GREEN}✓ Tests passed!${NC}"
        
        # Display coverage summary
        echo ""
        echo -e "${YELLOW}Coverage Summary:${NC}"
        go tool cover -func="$COVERAGE_FILE"
        
        # Generate HTML coverage report if requested
        if [ "$HTML_REPORT" = true ]; then
            HTML_FILE="$REPORTS_DIR/coverage_${REPORT_PREFIX}_${TIMESTAMP}.html"
            go tool cover -html="$COVERAGE_FILE" -o "$HTML_FILE"
            echo ""
            echo -e "${GREEN}✓ HTML coverage report generated: $HTML_FILE${NC}"
        fi
        
        echo -e "${BLUE}Coverage file saved: $COVERAGE_FILE${NC}"
    else
        echo -e "${RED}✗ Tests failed!${NC}"
        exit $TEST_EXIT_CODE
    fi
else
    # Run tests without coverage
    echo -e "${YELLOW}Running tests...${NC}"
    
    go test $TEST_FLAGS "$PACKAGE_PATH" 2>&1 | tee "$REPORTS_DIR/test_output_${REPORT_PREFIX}_${TIMESTAMP}.log"
    
    TEST_EXIT_CODE=${PIPESTATUS[0]}
    
    if [ $TEST_EXIT_CODE -eq 0 ]; then
        echo -e "${GREEN}✓ Tests passed!${NC}"
    else
        echo -e "${RED}✗ Tests failed!${NC}"
        exit $TEST_EXIT_CODE
    fi
fi

echo ""
echo -e "${BLUE}Test output saved: $REPORTS_DIR/test_output_${REPORT_PREFIX}_${TIMESTAMP}.log${NC}"
echo -e "${BLUE}========================================${NC}"
