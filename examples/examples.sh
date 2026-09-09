#!/bin/bash

set -u

BASE_URL="${MULTIFISH_URL:-http://localhost:8080/MultiFish/v1}"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PAYLOADS_DIR="$PROJECT_DIR/payloads"

print_header() {
    printf '\n========================================\n%s\n========================================\n' "$1"
}

request() {
    local method="$1"
    local endpoint="$2"
    local data="${3:-}"

    if [[ -n "$data" ]]; then
        curl --fail-with-body -sS -X "$method" "$BASE_URL$endpoint" \
            -H 'Content-Type: application/json' \
            --data "$data" | jq .
    else
        curl --fail-with-body -sS -X "$method" "$BASE_URL$endpoint" | jq .
    fi
}

platform_examples() {
    print_header 'Platform examples'
    request GET /Platform
    request GET /Platform/server1
}

manager_examples() {
    print_header 'Manager examples'
    request GET /Platform/server1/Managers
    request GET /Platform/server1/Managers/bmc
}

profile_examples() {
    print_header 'Profile examples'
    request GET /Platform/server1/Managers/bmc/Profiles
}

fan_controller_examples() {
    print_header 'Fan controller examples'
    request GET /Platform/server1/Managers/bmc/FanController
}

job_creation_examples() {
    print_header 'Job creation examples'
    if [[ -f "$PAYLOADS_DIR/continuous_daily.json" ]]; then
        request POST /JobService/Jobs "$(cat "$PAYLOADS_DIR/continuous_daily.json")"
    else
        printf 'Payload not found: %s\n' "$PAYLOADS_DIR/continuous_daily.json" >&2
        return 1
    fi
}

run_all_examples() {
    platform_examples
    manager_examples
    profile_examples
    fan_controller_examples
    job_creation_examples
}

usage() {
    printf 'Usage: %s [platform|manager|profile|fan-controller|job-create|all]\n' "$0"
}

main() {
    local example="${1:-all}"

    case "$example" in
        platform) platform_examples ;;
        manager) manager_examples ;;
        profile) profile_examples ;;
        fan-controller) fan_controller_examples ;;
        job-create) job_creation_examples ;;
        all) run_all_examples ;;
        -h|--help) usage ;;
        *) usage; return 1 ;;
    esac
}

main "$@"
