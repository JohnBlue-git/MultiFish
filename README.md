# MultiFish - Multi-BMC Redfish Management API

MultiFish is a Go service for managing multiple BMC platforms through Redfish and OEM provider implementations. This document is the user-facing guide for installing, configuring, running, and using the service.

For internal architecture and project structure, see [DESIGN.md](docs/DESIGN.md).

## Features

- Manage multiple BMC platforms through a single API.
- Support common Redfish operations and extended OEM operations.
- Schedule profile, manager, fan, and PID actions.
- Authenticate requests with token or basic authentication.
- Apply configurable rate limiting.
- Run locally, under systemd, in Docker, or on Kubernetes.

## Table of Contents

- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Running the Service](#running-the-service)
- [Deployment](#deployment)
- [Security](#security)
- [API Usage](#api-usage)
- [API Examples](#api-examples)
- [Scheduled Jobs](#scheduled-jobs)
- [Payload Examples](#payload-examples)
- [Logging and Troubleshooting](#logging-and-troubleshooting)
- [Development and Testing](#development-and-testing)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [Acknowledgements](#acknowledgements)
- [Support](#support)

## Quick Start

### Prerequisites

- Go 1.24 or later for local builds
- A reachable Redfish-compatible BMC for platform operations
- `curl` and `jq` for command-line examples

### Build and start

```bash
go build -o multifish .
./multifish
```

The default server listens on `http://localhost:8080`.

To use a configuration file, copy the development template and edit it:

```bash
cp config/config.example.yaml config/config.yaml
./multifish -config config/config.yaml
```

The local runtime file `config/config.yaml` is ignored by Git.

### Verify the service

```bash
curl http://localhost:8080/MultiFish/v1
```

## Configuration

MultiFish loads settings from built-in defaults, optionally overlaid by a YAML file, then overlaid by environment variables — in that order, so environment variables win over the YAML file, which wins over defaults. The only command-line flag is `-config <path>`, which selects which YAML file to load; it does not set individual fields.

### Configuration files

- [config/config.example.yaml](config/config.example.yaml): development template
- [config/config.production.yaml](config/config.production.yaml): production template
- [config/.env.production.example](config/.env.production.example): environment variable template
- [config/README.md](config/README.md): complete configuration reference

Use a YAML file explicitly:

```bash
./multifish -config config/config.production.yaml
```

Common environment variables include `PORT`, `LOG_LEVEL`, `WORKER_POOL_SIZE`, `TOKEN_AUTH_TOKENS`, and `AUTH_MODE`. See [SECURITY.md](docs/SECURITY.md) for authentication settings.

## Running the Service

The management script is located at [service/multifish.sh](service/multifish.sh):

```bash
./service/multifish.sh build
./service/multifish.sh start
./service/multifish.sh status
./service/multifish.sh logs
./service/multifish.sh stop
./service/multifish.sh restart
./service/multifish.sh test
```

Use a specific configuration file:

```bash
./service/multifish.sh -c config/config.production.yaml start
```

## Deployment

### Systemd

Edit [service/multifish.service](service/multifish.service) for the target user and installation path, then install it:

```bash
sudo cp service/multifish.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now multifish
sudo systemctl status multifish
```

### Docker

The maintained Dockerfile is [docker/Dockerfile](docker/Dockerfile). Build from the repository root so the Docker build context includes the Go module and source files:

```bash
docker build -f docker/Dockerfile -t multifish:latest .
docker run -d --name multifish -p 8080:8080 multifish:latest
```

### Kubernetes

Review the manifests in [k8s/](k8s/) and configure the Secret and ConfigMap for the target environment before applying them:

```bash
kubectl apply -f k8s/
```

See [DEPLOYMENT.md](docs/DEPLOYMENT.md) for production deployment procedures and operational details.

## Security

Authentication and rate limiting should be enabled for any network-facing deployment. At minimum:

- use strong, unique authentication tokens;
- avoid committing production secrets or local configuration files;
- restrict BMC network access to trusted service hosts;
- use HTTPS or a protected network path for API traffic;
- review logs for authentication failures and unexpected platform access.

See [SECURITY.md](docs/SECURITY.md) for the security model, authentication modes, and production checklist.

## API Usage

The API base path is `/MultiFish/v1`.

### Platform management

```bash
# List registered platforms
curl http://localhost:8080/MultiFish/v1/Platform

# Register a platform
curl -X POST http://localhost:8080/MultiFish/v1/Platform \
  -H 'Content-Type: application/json' \
  -d @payloads/patch_profile.json

# Read a platform
curl http://localhost:8080/MultiFish/v1/Platform/server1
```

### Manager and OEM operations

```bash
curl http://localhost:8080/MultiFish/v1/Platform/server1/Managers
curl http://localhost:8080/MultiFish/v1/Platform/server1/Managers/bmc
```

Available platform, manager, fan, profile, and PID operations are documented in [docs/PLATFORM.md](docs/PLATFORM.md).

### Job service

```bash
# List jobs
curl http://localhost:8080/MultiFish/v1/JobService/Jobs

# Create a job from a payload file
curl -X POST http://localhost:8080/MultiFish/v1/JobService/Jobs \
  -H 'Content-Type: application/json' \
  -d @payloads/continuous_daily.json
```

Job lifecycle, schedules, actions, and response formats are documented in [docs/JOBSERVICE.md](docs/JOBSERVICE.md).

## API Examples

The following examples show a typical platform workflow. Replace the endpoint,
credentials, and platform identifiers with values for your environment.

### Register a platform

```bash
curl -X POST http://localhost:8080/MultiFish/v1/Platform \
  -H 'Content-Type: application/json' \
  -d '{
    "Id": "server-1",
    "Name": "Production Server 1",
    "Type": "Extend",
    "Endpoint": "https://192.168.1.100",
    "Username": "root",
    "Password": "password",
    "Insecure": true
  }'
```

### Update a thermal profile

```bash
curl -X PATCH http://localhost:8080/MultiFish/v1/Platform/server-1/Managers/bmc/Oem/OpenBmc/Fan/Profile \
  -H 'Content-Type: application/json' \
  -d '{"Profile": "Performance"}'
```

Common profiles include `Performance`, `Balanced`, `PowerSaver`, and `Custom`.

### Configure a fan controller

```bash
curl -X PATCH http://localhost:8080/MultiFish/v1/Platform/server-1/Managers/bmc/Oem/OpenBmc/Fan/FanControllers/cpu_fan \
  -H 'Content-Type: application/json' \
  -d '{"Multiplier": 1.2, "StepDown": 2, "StepUp": 5}'
```

### Update manager properties

```bash
curl -X PATCH http://localhost:8080/MultiFish/v1/Platform/server-1/Managers/bmc \
  -H 'Content-Type: application/json' \
  -d '{"ServiceIdentification": "Production BMC v2.0"}'
```

## Scheduled Jobs

Jobs can run once or continuously according to a schedule. Common uses include:

- applying a thermal profile at a fixed time;
- changing fan controller settings during a maintenance window;
- applying different profiles on weekdays and weekends;
- managing multiple managers on one platform.

The scheduler uses the configured worker pool to bound concurrent execution. Validate payloads before submitting production jobs and monitor job status after creation.

## Payload Examples

Reusable JSON payloads are available in [payloads/](payloads/):

- `patch_profile.json`
- `patch_profile_multiple_managers.json`
- `patch_manager.json`
- `patch_fan_controller.json`
- `patch_fan_zone.json`
- `patch_pid_controller.json`
- `continuous_daily.json`
- `continuous_weekdays.json`
- `continuous_monthly.json`

## Usage Examples

### Shell examples

The executable [examples/examples.sh](examples/examples.sh) supports focused examples or the complete sequence:

```bash
./examples/examples.sh platform
./examples/examples.sh manager
./examples/examples.sh profile
./examples/examples.sh fan-controller
./examples/examples.sh job-create
./examples/examples.sh all
```

Set `MULTIFISH_URL` when the API is not at the default address:

```bash
MULTIFISH_URL=http://localhost:9090/MultiFish/v1 ./examples/examples.sh platform
```

### Postman

Import [examples/MultiFish.postman_collection.json](examples/MultiFish.postman_collection.json) into Postman for interactive API requests.

## Logging and Troubleshooting

The default log level and output directory are controlled by configuration. For a service managed by systemd:

```bash
sudo journalctl -u multifish -f
sudo systemctl status multifish
```

For a local process, inspect the configured log file or run the management script:

```bash
./service/multifish.sh logs
```

### Service will not start

1. Check whether port 8080 is already in use.
2. Confirm the selected configuration file exists and is readable.
3. Check authentication and rate-limit settings for validation errors.
4. Run `./service/multifish.sh build` and inspect the output.

### Platform registration fails

1. Verify the BMC endpoint is reachable from the MultiFish host.
2. Confirm credentials and the provider type.
3. Check TLS and `Insecure` settings for the target environment.
4. Inspect the service logs for provider or connection errors.

### Jobs do not execute

1. Confirm the job payload passed validation.
2. Check the job status through `/JobService/Jobs`.
3. Verify the schedule time and timezone.
4. Confirm the worker pool is not exhausted.

## Development and Testing

Install dependencies and run the service locally:

```bash
go mod download
go run .
```

Run all tests:

```bash
go test ./...
```

Focused test and coverage helpers are documented in [docs/TESTS.md](docs/TESTS.md).

## Documentation

- [Design guide and project structure](docs/DESIGN.md)
- [Deployment guide](docs/DEPLOYMENT.md)
- [Security guide](docs/SECURITY.md)
- [Configuration reference](config/README.md)
- [Platform API guide](docs/PLATFORM.md)
- [Job Service guide](docs/JOBSERVICE.md)
- [Provider guide](docs/PROVIDER.md)
- [Scheduler guide](docs/SCEDULER.md)
- [Test guide](docs/TESTS.md)

### Feature Documentation

- [Platform Management](docs/PLATFORM.md): platform registration, managers, and provider operations
- [Job Service](docs/JOBSERVICE.md): schedules, actions, worker pools, and execution logs

### Internal Module Documentation

- [Config module](config/README.md): configuration sources and validation
- [Providers module](docs/PROVIDER.md): provider architecture and BMC support
- [Scheduler module](docs/SCEDULER.md): job scheduling internals
- [Utility module](utility/README.md): shared helpers, logging, and errors
- [Testing guide](docs/TESTS.md): test commands and coverage helpers

## Contributing

1. Follow Go conventions and the existing package patterns.
2. Add focused tests for new behavior and bug fixes.
3. Update the relevant user or design documentation.
4. Run `go test ./...` before submitting changes.
5. Keep commits focused and describe the behavior they change.

Provider additions should follow the extension model described in [DESIGN.md](docs/DESIGN.md#provider-extension-model) and the guidance in [docs/PROVIDER.md](docs/PROVIDER.md).

## Acknowledgements

- Built with [Gofish](https://github.com/stmcginnis/gofish), a Redfish and Swordfish client library.
- Uses [Gin](https://github.com/gin-gonic/gin) for the HTTP web framework.
- Inspired by the Redfish specification from [DMTF](https://www.dmtf.org/standards/redfish).

## Support

For issues or questions:

1. Check [Logging and Troubleshooting](#logging-and-troubleshooting).
2. Review the relevant [module documentation](#documentation).
3. Try the [shell examples](#shell-examples) or the [Postman collection](#postman).
4. Open an issue in the repository with the version, configuration shape, request, and relevant logs. Do not include credentials or tokens.

### Quick Links

- [Quick Start](#quick-start)
- [Design guide](docs/DESIGN.md)
- [Deployment guide](docs/DEPLOYMENT.md)
- [Security guide](docs/SECURITY.md)
- [Shell examples](examples/examples.sh)
- [Postman collection](examples/MultiFish.postman_collection.json)
- [Payload examples](payloads/)
