#!/bin/bash#!/bin/bash



# MultiFish API Examples# MultiFish API Quick Start Examples

# This script demonstrates all available API endpoints with example requests# This script demonstrates common API operations with both Base and Extend types

# Prerequisites: MultiFish service must be running (default: http://localhost:8080)

BASE_URL="http://localhost:8080/MultiFish/v1"

BASE_URL="${MULTIFISH_URL:-http://localhost:8080}"

PAYLOADS_DIR="./payloads"echo "==================================="

echo "MultiFish API Quick Start Examples"

# Colors for outputecho "==================================="

GREEN='\033[0;32m'echo ""

BLUE='\033[0;34m'

YELLOW='\033[1;33m'# 1. Get Service Root

RED='\033[0;31m'echo "1. Getting Service Root..."

NC='\033[0m' # No Colorcurl -s -X GET "$BASE_URL" | jq .

echo ""

# Helper function to print section headers

print_header() {# 2. Add a machine with Extend type (OpenBMC with OEM features)

    echo -e "\n${BLUE}========================================${NC}"echo "2. Adding a new machine with Extend type (server1)..."

    echo -e "${BLUE}$1${NC}"curl -s -X POST "$BASE_URL/Platform" \

    echo -e "${BLUE}========================================${NC}\n"  -H "Content-Type: application/json" \

}  -d '{

    "Id": "server1",

# Helper function to print sub-headers    "Name": "OpenBMC Production Server",

print_subheader() {    "Type": "Extend",

    echo -e "\n${GREEN}>>> $1${NC}\n"    "Endpoint": "https://192.168.10.181",

}    "Username": "root",

    "Password": "0penBmc",

# Helper function to print warnings    "Insecure": true,

print_warning() {    "HTTPClientTimeout": 30,

    echo -e "${YELLOW}⚠ $1${NC}"    "DisableEtagMatch": true

}  }' | jq .

echo ""

# Helper function to print errors  }' | jq .

print_error() {echo ""

    echo -e "${RED}✗ $1${NC}"

}# 2b. Add a machine with Base type (Standard Redfish)

echo "2b. Adding a machine with Base type (server2)..."

# Helper function to print successcurl -s -X POST "$BASE_URL/Platform" \

print_success() {  -H "Content-Type: application/json" \

    echo -e "${GREEN}✓ $1${NC}"  -d '{

}    "Id": "server2",

    "Name": "Standard Redfish Server",

# Helper function to make API calls with pretty output    "Type": "Base",

api_call() {    "Endpoint": "https://192.168.10.182",

    local method=$1    "Username": "admin",

    local endpoint=$2    "Password": "password",

    local data=$3    "Insecure": true,

    local description=$4    "HTTPClientTimeout": 30

      }' | jq .

    print_subheader "$description"echo ""

    echo "Request: $method $BASE_URL$endpoint"

    # 3. List all machines

    if [ -n "$data" ]; thenecho "3. Listing all machines..."

        echo "Payload:"curl -s -X GET "$BASE_URL/Platform" | jq .

        echo "$data" | jq '.' 2>/dev/null || echo "$data"echo ""

        echo ""

        response=$(curl -s -X "$method" "$BASE_URL$endpoint" \# 4. Get machine details (shows Type field)

            -H "Content-Type: application/json" \echo "4. Getting machine details (Extend type)..."

            -d "$data")curl -s -X GET "$BASE_URL/Platform/server1" | jq .

    elseecho ""

        echo ""

        response=$(curl -s -X "$method" "$BASE_URL$endpoint")# 5. Get manager details

    fiecho "5. Getting manager details (Extend type with OEM)..."

    curl -s -X GET "$BASE_URL/Platform/server1/Managers/bmc" | jq .

    echo "Response:"echo ""

    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    echo ""# 5b. Get manager details for Base type

}echo "5b. Getting manager details (Base type, no OEM)..."

curl -s -X GET "$BASE_URL/Platform/server2/Managers/bmc" | jq .

# Helper function to make API calls with file payloadecho ""

api_call_file() {

    local method=$1# 6. Update ServiceIdentification

    local endpoint=$2echo "6. Updating ServiceIdentification..."

    local file=$3curl -s -X PATCH "$BASE_URL/Platform/server1/Managers/bmc" \

    local description=$4  -H "Content-Type: application/json" \

      -d '{

    if [ ! -f "$file" ]; then    "ServiceIdentification": "Production BMC Server"

        print_error "Payload file not found: $file"  }' | jq .

        return 1echo ""

    fi

    # 7. Get Profile

    local data=$(cat "$file")echo "7. Getting current profile..."

    api_call "$method" "$endpoint" "$data" "$description"curl -s -X GET "$BASE_URL/Platform/server1/Managers/bmc/Oem/OpenBmc/Profile" | jq .

}echo ""



############################################## 8. Update Profile

# Platform Discovery & Informationecho "8. Updating profile to Performance..."

#############################################curl -s -X PATCH "$BASE_URL/Platform/server1/Managers/bmc/Oem/OpenBmc/Profile" \

  -H "Content-Type: application/json" \

platform_examples() {  -d '{

    print_header "PLATFORM DISCOVERY & INFORMATION"    "Profile": "Performance"

      }' | jq .

    # Get all platformsecho ""

    api_call "GET" "/MultiFish/v1/Platforms" "" \

        "Get all registered platforms (machines)"# 9. Get Fan Controller (replace Fan_9 with actual fan controller ID)

    echo "9. Getting fan controller details..."

    # Get specific platformcurl -s -X GET "$BASE_URL/Platform/server1/Managers/bmc/Oem/OpenBmc/FanController/Fan_9" | jq .

    api_call "GET" "/MultiFish/v1/Platforms/machine-1" "" \echo ""

        "Get details for a specific platform"

    # 10. Update Fan Controller

    # Get platform systemsecho "10. Updating fan controller settings..."

    api_call "GET" "/MultiFish/v1/Platforms/machine-1/Systems" "" \curl -s -X PATCH "$BASE_URL/Platform/server1/Managers/bmc/Oem/OpenBmc/FanController/Fan_9" \

        "Get systems under a platform"  -H "Content-Type: application/json" \

      -d '{

    # Get platform managers    "FFGainCoefficient": 1.0,

    api_call "GET" "/MultiFish/v1/Platforms/machine-1/Managers" "" \    "FFOffCoefficient": 0.0,

        "Get managers under a platform"    "ICoefficient": 0.0,

}    "ILimitMax": 0.0,

    "ILimitMin": 0.0,

#############################################    "Inputs": ["Fan 9"],

# Manager Operations    "NegativeHysteresis": 0.0,

#############################################    "OutLimitMax": 100.0,

    "OutLimitMin": 30.0,

manager_examples() {    "Outputs": ["Pwm 15"],

    print_header "MANAGER OPERATIONS"    "PCoefficient": 0.0,

        "PositiveHysteresis": 0.0,

    # Get manager details    "SlewNeg": 0.0,

    api_call "GET" "/MultiFish/v1/Platforms/machine-1/Managers/bmc" "" \    "SlewPos": 0.0,

        "Get manager details"    "Zones": [

          {

    # Patch manager        "@odata.id": "/redfish/v1/Managers/bmc#/Oem/OpenBmc/Fan/FanZones/Right"

    local patch_manager_payload='{      }

  "ServiceIdentification": "Updated BMC Service"    ]

}'  }' | jq .

    api_call "PATCH" "/MultiFish/v1/Platforms/machine-1/Managers/bmc" \echo ""

        "$patch_manager_payload" \

        "Patch manager properties"# 11. Update machine configuration

    echo "11. Updating machine configuration..."

    # Get manager OEM datacurl -s -X PATCH "$BASE_URL/Platform/server1" \

    api_call "GET" "/MultiFish/v1/Platforms/machine-1/Managers/bmc/Oem" "" \  -H "Content-Type: application/json" \

        "Get manager OEM extensions"  -d '{

}    "Name": "Updated Server Name",

    "HTTPClientTimeout": 60

#############################################  }' | jq .

# Profile Management (Extended API)echo ""

#############################################

# 12. Delete machine (commented out to prevent accidental deletion)

profile_examples() {# echo "12. Deleting machine..."

    print_header "PROFILE MANAGEMENT (OEM Extended)"# curl -s -X DELETE "$BASE_URL/Platform/server1" | jq .

    # echo ""

    # Get current profile

    api_call "GET" "/MultiFish/v1/Platforms/machine-1/Managers/bmc/Oem/Profile" "" \echo "==================================="

        "Get current profile setting"echo "Examples completed!"

    echo "==================================="

    # Patch profile to Performance
    local patch_profile_payload='{
  "Profile": "Performance"
}'
    api_call "PATCH" "/MultiFish/v1/Platforms/machine-1/Managers/bmc/Oem/Profile" \
        "$patch_profile_payload" \
        "Set profile to Performance mode"
    
    # Patch profile to Balanced
    patch_profile_payload='{
  "Profile": "Balanced"
}'
    api_call "PATCH" "/MultiFish/v1/Platforms/machine-1/Managers/bmc/Oem/Profile" \
        "$patch_profile_payload" \
        "Set profile to Balanced mode"
    
    print_warning "Valid profile values: Performance, Balanced, PowerSaver, Custom"
}

#############################################
# Fan Controller Management
#############################################

fan_controller_examples() {
    print_header "FAN CONTROLLER MANAGEMENT (OEM Extended)"
    
    # Get fan controllers collection
    api_call "GET" "/MultiFish/v1/Platforms/machine-1/Managers/bmc/Oem/FanControllers" "" \
        "Get all fan controllers"
    
    # Get specific fan controller
    api_call "GET" "/MultiFish/v1/Platforms/machine-1/Managers/bmc/Oem/FanControllers/cpu_fan_controller" "" \
        "Get specific fan controller details"
    
    # Patch fan controller
    local patch_fan_controller_payload='{
  "Multiplier": 1.5,
  "StepDown": 2,
  "StepUp": 5
}'
    api_call "PATCH" "/MultiFish/v1/Platforms/machine-1/Managers/bmc/Oem/FanControllers/cpu_fan_controller" \
        "$patch_fan_controller_payload" \
        "Update fan controller parameters"
}

#############################################
# Fan Zone Management
#############################################

fan_zone_examples() {
    print_header "FAN ZONE MANAGEMENT (OEM Extended)"
    
    # Get fan zones collection
    api_call "GET" "/MultiFish/v1/Platforms/machine-1/Managers/bmc/Oem/FanZones" "" \
        "Get all fan zones"
    
    # Get specific fan zone
    api_call "GET" "/MultiFish/v1/Platforms/machine-1/Managers/bmc/Oem/FanZones/cpu_zone" "" \
        "Get specific fan zone details"
    
    # Patch fan zone
    local patch_fan_zone_payload='{
  "FailSafePercent": 100.0,
  "MinThermalOutput": 35.0
}'
    api_call "PATCH" "/MultiFish/v1/Platforms/machine-1/Managers/bmc/Oem/FanZones/cpu_zone" \
        "$patch_fan_zone_payload" \
        "Update fan zone configuration"
}

#############################################
# PID Controller Management
#############################################

pid_controller_examples() {
    print_header "PID CONTROLLER MANAGEMENT (OEM Extended)"
    
    # Get PID controllers collection
    api_call "GET" "/MultiFish/v1/Platforms/machine-1/Managers/bmc/Oem/PidControllers" "" \
        "Get all PID controllers"
    
    # Get specific PID controller
    api_call "GET" "/MultiFish/v1/Platforms/machine-1/Managers/bmc/Oem/PidControllers/cpu_temp_controller" "" \
        "Get specific PID controller details"
    
    # Patch PID controller
    local patch_pid_payload='{
  "FFGainCoefficient": 0.5,
  "FFOffCoefficient": 10.0,
  "ICoefficient": 0.1,
  "ILimitMax": 50.0,
  "ILimitMin": -50.0,
  "OutLimitMax": 100.0,
  "OutLimitMin": 0.0,
  "PCoefficient": 0.8,
  "SetPoint": 75.0,
  "SlewNeg": -5.0,
  "SlewPos": 5.0
}'
    api_call "PATCH" "/MultiFish/v1/Platforms/machine-1/Managers/bmc/Oem/PidControllers/cpu_temp_controller" \
        "$patch_pid_payload" \
        "Update PID controller parameters"
}

#############################################
# Job Service Management
#############################################

job_service_examples() {
    print_header "JOB SERVICE MANAGEMENT"
    
    # Get job service root
    api_call "GET" "/MultiFish/v1/JobService" "" \
        "Get job service information and capabilities"
    
    # Update worker pool size
    local update_pool_payload='{
  "WorkerPoolSize": 10
}'
    api_call "PATCH" "/MultiFish/v1/JobService" \
        "$update_pool_payload" \
        "Update worker pool size"
    
    # Get all jobs
    api_call "GET" "/MultiFish/v1/JobService/Jobs" "" \
        "Get all jobs in the system"
}

#############################################
# Job Creation Examples
#############################################

job_creation_examples() {
    print_header "JOB CREATION EXAMPLES"
    
    # Create job - Patch Profile (Once)
    print_subheader "Create a one-time profile update job"
    api_call_file "POST" "/MultiFish/v1/JobService/Jobs" \
        "$PAYLOADS_DIR/patch_profile.json" \
        "Schedule one-time profile update"
    
    # Create job - Patch Profile Multiple Managers
    print_subheader "Create job for multiple managers"
    api_call_file "POST" "/MultiFish/v1/JobService/Jobs" \
        "$PAYLOADS_DIR/patch_profile_multiple_managers.json" \
        "Schedule profile update to multiple managers"
    
    # Create job - Patch Manager
    print_subheader "Create manager update job"
    api_call_file "POST" "/MultiFish/v1/JobService/Jobs" \
        "$PAYLOADS_DIR/patch_manager.json" \
        "Schedule manager property update"
    
    # Create job - Patch Fan Controller
    print_subheader "Create fan controller update job"
    api_call_file "POST" "/MultiFish/v1/JobService/Jobs" \
        "$PAYLOADS_DIR/patch_fan_controller.json" \
        "Schedule fan controller update"
    
    # Create job - Patch Fan Zone
    print_subheader "Create fan zone update job"
    api_call_file "POST" "/MultiFish/v1/JobService/Jobs" \
        "$PAYLOADS_DIR/patch_fan_zone.json" \
        "Schedule fan zone update"
    
    # Create job - Patch PID Controller
    print_subheader "Create PID controller update job"
    api_call_file "POST" "/MultiFish/v1/JobService/Jobs" \
        "$PAYLOADS_DIR/patch_pid_controller.json" \
        "Schedule PID controller update"
}

#############################################
# Continuous Schedule Examples
#############################################

continuous_schedule_examples() {
    print_header "CONTINUOUS SCHEDULE EXAMPLES"
    
    # Daily schedule
    print_subheader "Create daily recurring job"
    api_call_file "POST" "/MultiFish/v1/JobService/Jobs" \
        "$PAYLOADS_DIR/continuous_daily.json" \
        "Schedule daily profile update at 22:00"
    
    # Weekday schedule
    print_subheader "Create weekday recurring job"
    api_call_file "POST" "/MultiFish/v1/JobService/Jobs" \
        "$PAYLOADS_DIR/continuous_weekdays.json" \
        "Schedule weekday profile update at 08:00"
    
    # Monthly schedule
    print_subheader "Create monthly recurring job"
    api_call_file "POST" "/MultiFish/v1/JobService/Jobs" \
        "$PAYLOADS_DIR/continuous_monthly.json" \
        "Schedule monthly profile update on 1st and 15th"
}

#############################################
# Job Management Operations
#############################################

job_management_examples() {
    print_header "JOB MANAGEMENT OPERATIONS"
    
    print_warning "These examples assume job IDs exist. Replace with actual job IDs."
    
    # Get specific job
    api_call "GET" "/MultiFish/v1/JobService/Jobs/job_12345" "" \
        "Get details of a specific job"
    
    # Update job
    local update_job_payload='{
  "Name": "Updated Job Name",
  "Schedule": {
    "Type": "Once",
    "Time": "14:00:00",
    "Period": null
  }
}'
    api_call "PATCH" "/MultiFish/v1/JobService/Jobs/job_12345" \
        "$update_job_payload" \
        "Update job name and schedule"
    
    # Delete job
    api_call "DELETE" "/MultiFish/v1/JobService/Jobs/job_12345" "" \
        "Delete a job"
    
    # Trigger job immediately
    api_call "POST" "/MultiFish/v1/JobService/Jobs/job_12345/Actions/Trigger" "" \
        "Trigger job execution immediately"
}

#############################################
# Main Menu
#############################################

show_menu() {
    echo -e "\n${BLUE}╔════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║   MultiFish API Examples Menu         ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════╝${NC}\n"
    echo "1.  Platform Discovery & Information"
    echo "2.  Manager Operations"
    echo "3.  Profile Management (OEM)"
    echo "4.  Fan Controller Management (OEM)"
    echo "5.  Fan Zone Management (OEM)"
    echo "6.  PID Controller Management (OEM)"
    echo "7.  Job Service Management"
    echo "8.  Job Creation Examples"
    echo "9.  Continuous Schedule Examples"
    echo "10. Job Management Operations"
    echo "11. Run ALL Examples"
    echo "0.  Exit"
    echo ""
}

run_all_examples() {
    platform_examples
    manager_examples
    profile_examples
    fan_controller_examples
    fan_zone_examples
    pid_controller_examples
    job_service_examples
    job_creation_examples
    continuous_schedule_examples
    job_management_examples
}

#############################################
# Main Script
#############################################

main() {
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        print_warning "jq is not installed. Output will not be formatted."
        print_warning "Install jq for better output: sudo apt-get install jq"
    fi
    
    # Check if curl is installed
    if ! command -v curl &> /dev/null; then
        print_error "curl is not installed. Please install curl to use this script."
        exit 1
    fi
    
    # Check if service is running
    if ! curl -s "$BASE_URL/MultiFish/v1/Platforms" > /dev/null 2>&1; then
        print_error "MultiFish service is not running at $BASE_URL"
        print_warning "Start the service with: ./multifish or go run main.go"
        exit 1
    fi
    
    print_success "MultiFish service is running at $BASE_URL"
    
    # Interactive or direct mode
    if [ $# -eq 0 ]; then
        # Interactive mode
        while true; do
            show_menu
            read -p "Select an option (0-11): " choice
            
            case $choice in
                1) platform_examples ;;
                2) manager_examples ;;
                3) profile_examples ;;
                4) fan_controller_examples ;;
                5) fan_zone_examples ;;
                6) pid_controller_examples ;;
                7) job_service_examples ;;
                8) job_creation_examples ;;
                9) continuous_schedule_examples ;;
                10) job_management_examples ;;
                11) run_all_examples ;;
                0) 
                    print_success "Goodbye!"
                    exit 0 
                    ;;
                *) 
                    print_error "Invalid option. Please try again."
                    ;;
            esac
            
            read -p "Press Enter to continue..."
        done
    else
        # Direct mode with arguments
        case $1 in
            platform) platform_examples ;;
            manager) manager_examples ;;
            profile) profile_examples ;;
            fan-controller) fan_controller_examples ;;
            fan-zone) fan_zone_examples ;;
            pid-controller) pid_controller_examples ;;
            job-service) job_service_examples ;;
            job-create) job_creation_examples ;;
            job-continuous) continuous_schedule_examples ;;
            job-manage) job_management_examples ;;
            all) run_all_examples ;;
            *)
                echo "Usage: $0 [option]"
                echo ""
                echo "Options:"
                echo "  platform        - Platform discovery examples"
                echo "  manager         - Manager operations examples"
                echo "  profile         - Profile management examples"
                echo "  fan-controller  - Fan controller examples"
                echo "  fan-zone        - Fan zone examples"
                echo "  pid-controller  - PID controller examples"
                echo "  job-service     - Job service examples"
                echo "  job-create      - Job creation examples"
                echo "  job-continuous  - Continuous schedule examples"
                echo "  job-manage      - Job management examples"
                echo "  all             - Run all examples"
                echo ""
                echo "No arguments: Interactive mode"
                exit 1
                ;;
        esac
    fi
}

# Run main function
main "$@"
