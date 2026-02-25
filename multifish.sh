#!/bin/bash

# MultiFish Management Script
# Utility script for common MultiFish operations
# Supports running with or without configuration file

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BINARY="$SCRIPT_DIR/multifish"
PID_FILE="$SCRIPT_DIR/multifish.pid"
LOG_FILE="$SCRIPT_DIR/multifish.log"

# Configuration file (can be overridden with -c or --config)
CONFIG_FILE=""
DEFAULT_CONFIG="$SCRIPT_DIR/config.yaml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

function print_debug() {
    echo -e "${BLUE}⚙ $1${NC}"
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

    # Build command with optional config
    CMD="$BINARY"
    if [ -n "$CONFIG_FILE" ]; then
        if [ ! -f "$CONFIG_FILE" ]; then
            print_error "Configuration file not found: $CONFIG_FILE"
            return 1
        fi
        CMD="$CMD -config $CONFIG_FILE"
        print_info "Using configuration file: $CONFIG_FILE"
    elif [ -f "$DEFAULT_CONFIG" ]; then
        CMD="$CMD -config $DEFAULT_CONFIG"
        print_info "Using default configuration file: $DEFAULT_CONFIG"
    else
        print_info "Starting without configuration file (using defaults)"
    fi

    print_info "Starting MultiFish..."
    print_debug "Command: $CMD"
    
    nohup $CMD > "$LOG_FILE" 2>&1 &
    PID=$!
    echo $PID > "$PID_FILE"
    sleep 2
    
    if ps -p $PID > /dev/null 2>&1; then
        print_success "MultiFish started successfully (PID: $PID)"
        print_info "Logs: $LOG_FILE"
        return 0
    else
        print_error "Failed to start MultiFish"
        print_error "Check logs for details: $LOG_FILE"
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

Usage: $0 [options] {command}

Commands:
    build       Build the MultiFish binary
    start       Start the MultiFish service
    stop        Stop the MultiFish service
    restart     Restart the MultiFish service
    status      Check service status
    logs        Tail the log file
    test        Test API endpoints
    help        Show this help message

Options:
    -c, --config FILE    Use specified configuration file
                         (default: config.yaml if exists)
    -h, --help          Show this help message

Configuration Priority:
    1. Explicitly specified config file (-c or --config)
    2. Default config.yaml in the same directory
    3. Built-in defaults (no config file)

Examples:
    # Build the binary
    $0 build

    # Start with default config (config.yaml if exists)
    $0 start

    # Start with specific config file
    $0 -c config.production.yaml start
    $0 --config /path/to/custom.yaml start

    # Start without any config file (use defaults)
    rm config.yaml  # Remove default config
    $0 start

    # Check status
    $0 status

    # View logs
    $0 logs

    # Test API
    $0 test

    # Restart with different config
    $0 -c config.production.yaml restart

Files:
    Binary:  $BINARY
    PID:     $PID_FILE
    Logs:    $LOG_FILE
    Config:  ${CONFIG_FILE:-$DEFAULT_CONFIG (if exists)}

Environment Variables:
    MULTIFISH_CONFIG    Default configuration file path
                        (overrides config.yaml)

EOF
}

# Main script logic

# Parse options
while [[ $# -gt 0 ]]; do
    case "$1" in
        -c|--config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -h|--help|help)
            usage
            exit 0
            ;;
        build|start|stop|restart|status|logs|test)
            COMMAND="$1"
            shift
            break
            ;;
        *)
            print_error "Invalid option or command: $1"
            echo ""
            usage
            exit 1
            ;;
    esac
done

# Check for environment variable if config not set
if [ -z "$CONFIG_FILE" ] && [ -n "$MULTIFISH_CONFIG" ]; then
    CONFIG_FILE="$MULTIFISH_CONFIG"
    print_info "Using config from MULTIFISH_CONFIG: $CONFIG_FILE"
fi

# Execute command
case "$COMMAND" in
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
    *)
        if [ -z "$COMMAND" ]; then
            print_error "No command specified"
        else
            print_error "Invalid command: $COMMAND"
        fi
        echo ""
        usage
        exit 1
        ;;
esac

exit $?
