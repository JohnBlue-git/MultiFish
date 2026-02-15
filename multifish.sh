#!/bin/bash

# MultiFish Management Script
# Utility script for common MultiFish operations

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BINARY="$SCRIPT_DIR/multifish"
PID_FILE="$SCRIPT_DIR/multifish.pid"
LOG_FILE="$SCRIPT_DIR/multifish.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

function print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

function print_error() {
    echo -e "${RED}✗ $1${NC}"
}

function print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

function build() {
    print_info "Building MultiFish..."
    cd "$SCRIPT_DIR"
    if go build -o multifish *.go; then
        print_success "Build completed successfully"
        return 0
    else
        print_error "Build failed"
        return 1
    fi
}

function start() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if ps -p $PID > /dev/null 2>&1; then
            print_error "MultiFish is already running (PID: $PID)"
            return 1
        else
            rm "$PID_FILE"
        fi
    fi

    if [ ! -f "$BINARY" ]; then
        print_error "Binary not found. Building first..."
        build || return 1
    fi

    print_info "Starting MultiFish..."
    nohup "$BINARY" > "$LOG_FILE" 2>&1 &
    PID=$!
    echo $PID > "$PID_FILE"
    sleep 2
    
    if ps -p $PID > /dev/null 2>&1; then
        print_success "MultiFish started successfully (PID: $PID)"
        print_info "Logs: $LOG_FILE"
        return 0
    else
        print_error "Failed to start MultiFish"
        rm "$PID_FILE"
        return 1
    fi
}

function stop() {
    if [ ! -f "$PID_FILE" ]; then
        print_error "MultiFish is not running"
        return 1
    fi

    PID=$(cat "$PID_FILE")
    if ps -p $PID > /dev/null 2>&1; then
        print_info "Stopping MultiFish (PID: $PID)..."
        kill $PID
        sleep 2
        
        if ps -p $PID > /dev/null 2>&1; then
            print_info "Force stopping..."
            kill -9 $PID
        fi
        
        rm "$PID_FILE"
        print_success "MultiFish stopped"
        return 0
    else
        print_error "Process not found (PID: $PID)"
        rm "$PID_FILE"
        return 1
    fi
}

function status() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if ps -p $PID > /dev/null 2>&1; then
            print_success "MultiFish is running (PID: $PID)"
            echo ""
            print_info "Testing API endpoint..."
            if curl -s http://localhost:8080/MultiFish/v1 > /dev/null; then
                print_success "API is responding"
            else
                print_error "API is not responding"
            fi
            return 0
        else
            print_error "PID file exists but process is not running"
            return 1
        fi
    else
        print_error "MultiFish is not running"
        return 1
    fi
}

function restart() {
    print_info "Restarting MultiFish..."
    stop
    sleep 1
    start
}

function logs() {
    if [ -f "$LOG_FILE" ]; then
        tail -f "$LOG_FILE"
    else
        print_error "Log file not found"
        return 1
    fi
}

function test_api() {
    print_info "Testing MultiFish API..."
    
    if ! curl -s http://localhost:8080/MultiFish/v1 > /dev/null; then
        print_error "API is not responding. Is the server running?"
        return 1
    fi
    
    print_success "API is responding"
    echo ""
    
    print_info "Service Root:"
    curl -s http://localhost:8080/MultiFish/v1 | jq .
    echo ""
    
    print_info "Platform (Machines):"
    curl -s http://localhost:8080/MultiFish/v1/Platform | jq .
}

function usage() {
    cat << EOF
MultiFish Management Script

Usage: $0 {command}

Commands:
    build       Build the MultiFish binary
    start       Start the MultiFish service
    stop        Stop the MultiFish service
    restart     Restart the MultiFish service
    status      Check service status
    logs        Tail the log file
    test        Test API endpoints
    help        Show this help message

Examples:
    $0 build
    $0 start
    $0 status
    $0 logs
    $0 test

Files:
    Binary:  $BINARY
    PID:     $PID_FILE
    Logs:    $LOG_FILE

EOF
}

# Main script logic
case "${1:-}" in
    build)
        build
        ;;
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    status)
        status
        ;;
    logs)
        logs
        ;;
    test)
        test_api
        ;;
    help|--help|-h)
        usage
        ;;
    *)
        print_error "Invalid command: ${1:-}"
        echo ""
        usage
        exit 1
        ;;
esac

exit $?
