# Tests Module# Tests Module



## Overview## Overview



The `tests/` directory contains comprehensive testing infrastructure for the MultiFish project, including test runners, coverage reports, and test output management.The `tests` directory provides **comprehensive testing infrastructure** for the MultiFish project, including automated test execution scripts, coverage reporting, and test result analysis tools.



## Why Testing?## Why Dedicated Test Infrastructure?



Testing ensures:As the project grows, testing needs become more complex:

- **Code Quality**: Validates that all components work as expected- **Multiple Packages**: Tests scattered across providers, scheduler, utility, etc.

- **Regression Prevention**: Catches bugs before they reach production- **Coverage Tracking**: Need to monitor code coverage over time

- **Documentation**: Tests serve as living examples of how code should behave- **CI/CD Integration**: Automated testing for continuous integration

- **Refactoring Confidence**: Enables safe code improvements- **Developer Productivity**: Quick feedback on code changes

- **Quality Assurance**: Consistent testing standards across the project

## Structure

The test infrastructure provides:

```1. **Automated Test Execution**: Run all or specific tests with one command

tests/2. **Coverage Reporting**: Track which code is tested

├── coverage_report.sh       # Generate code coverage reports3. **HTML Reports**: Visual coverage analysis

├── run_all_tests.sh        # Execute all test suites4. **Test Summaries**: Quick overview of test results

├── run_specific_test.sh    # Run specific tests or packages5. **Historical Tracking**: Timestamped reports for comparison

├── test_summary.sh         # Generate test result summaries

├── README.md               # This file## Directory Structure

└── reports/                # Test outputs and coverage reports

    ├── coverage_*.html     # HTML coverage reports```

    ├── coverage_*.out      # Go coverage datatests/

    ├── coverage_summary_*.txt  # Coverage summaries├── README.md                  # This file

    └── test_output_*.log   # Test execution logs├── run_all_tests.sh          # Run all tests in the project

```├── run_specific_test.sh      # Run specific package or test

├── coverage_report.sh        # Generate coverage reports

## Test Scripts├── test_summary.sh           # Generate test summary dashboard

└── reports/                  # Generated test reports (auto-created)

### 1. Run All Tests (`run_all_tests.sh`)    ├── coverage_*.out        # Coverage data files

    ├── coverage_*.html       # HTML coverage reports

Executes the entire test suite across all packages.    ├── test_output_*.log     # Test execution logs

    └── coverage_summary_*.txt # Coverage summaries

**Usage:**```

```bash

./tests/run_all_tests.sh## Test Organization

```

Tests are co-located with source code following Go conventions:

**What it does:**

- Discovers all test files in the project```

- Runs tests with verbose outputGofish/

- Generates timestamped test logs├── providers/

- Provides summary of test results│   ├── redfish/

│   │   ├── BaseManager.go

**Output:**│   │   └── BaseManager_test.go      # Tests for BaseManager

- Test results printed to console│   └── extend/

- Detailed log saved to `reports/test_output_TIMESTAMP.log`│       ├── ExtendManager.go

│       ├── ExtendManager_test.go    # Tests for ExtendManager

### 2. Generate Coverage Report (`coverage_report.sh`)│       ├── ExtendService.go

│       └── ExtendService_test.go    # Tests for ExtendService

Creates comprehensive code coverage reports.├── scheduler/

│   ├── job_models.go

**Usage:**│   ├── job_models_test.go           # Tests for job models

```bash│   ├── job_service.go

./tests/coverage_report.sh│   ├── job_service_test.go          # Tests for job service

```│   ├── job_executor.go

│   └── job_executor_test.go         # Tests for job executor

**What it does:**├── utility/

- Runs all tests with coverage profiling│   ├── errors.go

- Generates HTML coverage report│   ├── errors_test.go               # Tests for error handling

- Creates coverage summary statistics│   ├── helpers.go

- Identifies untested code paths│   └── helpers_test.go              # Tests for helper functions

└── tests/                           # Test infrastructure scripts

**Output:**```

- `reports/coverage_TIMESTAMP.html` - Visual coverage report

- `reports/coverage_TIMESTAMP.out` - Raw coverage data## Testing Scripts

- `reports/coverage_summary_TIMESTAMP.txt` - Coverage statistics

### 1. Run All Tests (`run_all_tests.sh`)

**Coverage Metrics:**

- Overall code coverage percentageRuns all tests in the entire project across all packages.

- Per-package coverage breakdown

- Per-function coverage details**Usage:**

- Untested line identification```bash

./run_all_tests.sh [options]

### 3. Run Specific Tests (`run_specific_test.sh`)```



Executes targeted tests for specific packages or test functions.**Options:**

- `-v, --verbose`: Enable verbose test output

**Usage:**- `-h, --html`: Generate HTML coverage report

```bash- `-c, --coverage`: Generate coverage report

# Run all tests in a package

./tests/run_specific_test.sh multifish/scheduler**Examples:**

```bash

# Run specific test function# Run all tests

./tests/run_specific_test.sh multifish/scheduler TestJobCreation./run_all_tests.sh



# Run tests matching pattern# Run all tests with verbose output

./tests/run_specific_test.sh multifish/scheduler TestJob.*./run_all_tests.sh -v

```

# Run all tests with coverage

**What it does:**./run_all_tests.sh -c

- Runs tests for specified package

- Optionally filters by test name pattern# Run all tests with coverage and HTML report

- Provides focused test feedback./run_all_tests.sh -c -h

```

### 4. Test Summary (`test_summary.sh`)

**What It Does:**

Generates human-readable test summaries from test logs.1. Discovers all test files in the project

2. Runs `go test` on all packages

**Usage:**3. Optionally generates coverage data

```bash4. Creates timestamped log files

./tests/test_summary.sh reports/test_output_20260209_154346.log5. Displays summary of results

```

**Output:**

**What it does:**```

- Parses test output logsRunning all tests...

- Summarizes pass/fail countsPackage: multifish/providers/redfish

- Lists failed tests with details  ✓ TestRedfishProviderSupports

- Provides quick test result overview  ✓ TestPatchManagerData

  ✓ TestManagerAllowedPatchFields

## Test Categories

Package: multifish/scheduler

### Unit Tests  ✓ TestJobValidation

  ✓ TestScheduleCalculation

Test individual functions and methods in isolation.  ✓ TestJobExecution



**Locations:**Total: 45 tests, 45 passed, 0 failed

- `scheduler/job_models_test.go` - Job model validationCoverage: 78.5%

- `scheduler/job_executor_test.go` - Job execution logic```

- `scheduler/job_service_test.go` - Job service operations

- `utility/errors_test.go` - Error handling### 2. Run Specific Tests (`run_specific_test.sh`)

- `utility/helpers_test.go` - Helper functions

- `providers/redfish/BaseManager_test.go` - Base manager operationsRun tests for a specific package or test function.

- `providers/extend/ExtendManager_test.go` - Extended manager operations

**Usage:**

**Example:**```bash

```go./run_specific_test.sh <package_path> [test_name] [options]

func TestJobValidation(t *testing.T) {```

    job := &Job{

        Name: "Test Job",**Options:**

        Machines: []string{"machine-1"},- `-v, --verbose`: Enable verbose test output

        Action: ActionPatchProfile,- `-h, --html`: Generate HTML coverage report

        // ...- `-c, --coverage`: Generate coverage report

    }

    **Examples:**

    err := job.Validate()```bash

    assert.NoError(t, err)# Run all tests in the scheduler package

}./run_specific_test.sh ./scheduler

```

# Run a specific test function

### Integration Tests./run_specific_test.sh ./scheduler TestJobExecutor



Test interactions between multiple components.# Run specific test with coverage and verbose output

./run_specific_test.sh ./scheduler TestJobExecutor -v -c

**Examples:**

- `handler/handleJobService_test.go` - API endpoint tests# Run specific test with HTML coverage report

- `handler/handlePlatform_test.go` - Platform management tests./run_specific_test.sh ./utility TestHelpers -c -h

- `scheduler/job_service_worker_pool_test.go` - Worker pool integration```



**Example:**### 3. Coverage Report (`coverage_report.sh`)

```go

func TestJobServiceIntegration(t *testing.T) {Generate comprehensive coverage reports with various output formats.

    // Setup test server

    router := setupTestRouter()**Usage:**

    ```bash

    // Create job via API./coverage_report.sh [options]

    w := httptest.NewRecorder()```

    req, _ := http.NewRequest("POST", "/MultiFish/v1/JobService/Jobs", body)

    router.ServeHTTP(w, req)**Options:**

    - `-h, --html`: Generate HTML coverage report (default: true)

    assert.Equal(t, http.StatusCreated, w.Code)- `-t, --text`: Display text coverage report in terminal

}- `-p, --packages`: Show per-package coverage breakdown

```- `--threshold N`: Fail if coverage is below N%



## Writing Tests**Examples:**

```bash

### Test File Naming# Generate HTML coverage report

./coverage_report.sh

- Test files must end with `_test.go`

- Name should match the file being tested: `job_models.go` → `job_models_test.go`# Show coverage breakdown by package

./coverage_report.sh -p

### Test Function Naming

# Show detailed text report

```go./coverage_report.sh -t

func TestFunctionName(t *testing.T) {

    // Test implementation# Generate all reports with package breakdown

}./coverage_report.sh -h -t -p



func TestFunctionName_ErrorCase(t *testing.T) {# Enforce minimum 70% coverage

    // Test error handling./coverage_report.sh --threshold 70

}```



func BenchmarkFunctionName(b *testing.B) {### 4. Test Summary (`test_summary.sh`)

    // Performance benchmark

}Generate a comprehensive dashboard showing test statistics and coverage trends.

```

**Usage:**

### Test Structure (AAA Pattern)```bash

./test_summary.sh

```go```

func TestExampleFunction(t *testing.T) {

    // Arrange - Setup test data and dependenciesThis script provides:

    input := "test data"- Total number of test files and test functions

    expected := "expected result"- Test execution status

    - Coverage statistics

    // Act - Execute the function being tested- Package-level breakdown

    result, err := ExampleFunction(input)- Recent test history

    

    // Assert - Verify the results## Quick Start

    assert.NoError(t, err)

    assert.Equal(t, expected, result)1. **Make scripts executable:**

}   ```bash

```   chmod +x *.sh

   ```

### Table-Driven Tests

2. **Run all tests:**

```go   ```bash

func TestValidation(t *testing.T) {   ./run_all_tests.sh

    tests := []struct {   ```

        name        string

        input       Job3. **Generate coverage report:**

        expectError bool   ```bash

        errorMsg    string   ./coverage_report.sh -p

    }{   ```

        {

            name: "valid job",4. **View HTML coverage report:**

            input: Job{Name: "Valid", Machines: []string{"m1"}},   ```bash

            expectError: false,   ./coverage_report.sh -h

        },   # Then open the generated HTML file in the reports/ directory

        {   ```

            name: "missing name",

            input: Job{Machines: []string{"m1"}},## Test Organization

            expectError: true,

            errorMsg: "name is required",The project contains tests in the following packages:

        },

    }- `./scheduler/` - Job scheduling and execution tests

      - `job_models_test.go`

    for _, tt := range tests {  - `job_service_test.go`

        t.Run(tt.name, func(t *testing.T) {  - `job_executor_test.go`

            err := tt.input.Validate()

            if tt.expectError {- `./utility/` - Utility function tests

                assert.Error(t, err)  - `errors_test.go`

                assert.Contains(t, err.Error(), tt.errorMsg)  - `helpers_test.go`

            } else {

                assert.NoError(t, err)- `./providers/redfish/` - Redfish provider tests

            }  - `BaseManager_test.go`

        })

    }- `./providers/extend/` - Extended provider tests

}  - `ExtendManager_test.go`

```  - `ExtendService_test.go`



## Test Coverage Goals- Root level handler tests:

  - `handler/handleJobService_test.go`

| Package | Target Coverage | Current Status |  - `handler/handlePlatform_test.go`

|---------|----------------|----------------|

| scheduler | 80%+ | ✓ Meeting target |## CI/CD Integration

| utility | 85%+ | ✓ Meeting target |

| providers | 75%+ | ✓ Meeting target |These scripts can be easily integrated into CI/CD pipelines:

| handlers | 70%+ | ✓ Meeting target |

```bash

## Continuous Integration# Example: Run tests with coverage threshold

./run_all_tests.sh -c

Tests should be run:./coverage_report.sh --threshold 70

- Before committing code changes```

- In CI/CD pipeline before deployment

- After dependencies are updated## Viewing Reports

- Before releases

### HTML Coverage Reports

## Troubleshooting

HTML coverage reports provide an interactive view of your code coverage:

### Tests Failing Locally

```bash

```bash# Generate HTML report

# Clean test cache./coverage_report.sh -h

go clean -testcache

# Open in default browser

# Run with verbose outputxdg-open tests/reports/coverage_<timestamp>.html

go test -v ./...```



# Run specific failing test### Text Reports

./tests/run_specific_test.sh package/path TestFailingFunction

```For quick terminal-based viewing:



### Coverage Report Not Generated```bash

# Show coverage summary

```bash./coverage_report.sh -t

# Ensure all dependencies are installed

go mod download# Show per-package breakdown

./coverage_report.sh -p

# Check write permissions```

ls -la tests/reports/

## Test Coverage Goals

# Run with debug output

./tests/coverage_report.sh -vCurrent coverage targets by package:

```

| Package | Target | Current | Status |

## Best Practices|---------|--------|---------|--------|

| scheduler | 80% | - | In Progress |

1. **Test Independence**: Each test should be self-contained| utility | 90% | - | In Progress |

2. **Cleanup**: Use `t.Cleanup()` for resource cleanup| providers | 75% | - | In Progress |

3. **Mocking**: Mock external dependencies (APIs, databases)| handlers | 70% | - | In Progress |

4. **Test Data**: Use realistic but minimal test data

5. **Error Testing**: Always test error paths## Writing Tests

6. **Performance**: Keep tests fast (<100ms per test)

### Test File Naming Convention

## Related Documentation

- Test files must end with `_test.go`

- [Scheduler README](../scheduler/README.md) - Scheduler implementation details- Should be in the same package as the code being tested

- [Utility README](../utility/README.md) - Testing utilities and helpers- Example: `job_service.go` → `job_service_test.go`

- [Providers README](../providers/README.md) - Provider testing strategies

- [Root README](../README.md) - Project overview### Test Function Naming



## Contributing```go

// Test a function

When adding new features:func TestFunctionName(t *testing.T) { }

1. Write tests first (TDD approach recommended)

2. Ensure tests pass locally// Test a method

3. Verify coverage meets targetsfunc TestStructName_MethodName(t *testing.T) { }

4. Update this documentation if needed

// Test with table-driven tests
func TestFunctionName_EdgeCase(t *testing.T) { }
```

### Example Test Structure

```go
package scheduler

import (
    "testing"
)

func TestJobValidation(t *testing.T) {
    tests := []struct {
        name    string
        job     *JobCreateRequest
        wantErr bool
    }{
        {
            name: "valid job",
            job: &JobCreateRequest{
                Machines: []string{"machine-1"},
                Action:   ActionPatchProfile,
                Payload:  validPayload,
                Schedule: validSchedule,
            },
            wantErr: false,
        },
        {
            name: "missing machines",
            job: &JobCreateRequest{
                Machines: []string{},
                Action:   ActionPatchProfile,
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            validation := tt.job.Validate()
            if (validation.Valid == false) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", validation, tt.wantErr)
            }
        })
    }
}
```

## Best Practices

1. **Use Table-Driven Tests**: Test multiple scenarios efficiently
2. **Test Edge Cases**: Empty inputs, nil values, boundary conditions
3. **Test Error Paths**: Ensure errors are handled correctly
4. **Mock External Dependencies**: Use interfaces for testability
5. **Keep Tests Fast**: Avoid network calls and file I/O when possible
6. **Name Tests Clearly**: Test names should describe what they test
7. **One Assertion Per Test**: When possible, test one thing at a time

## Troubleshooting

### Tests Failing

1. **Check test output logs**:
   ```bash
   cat tests/reports/test_output_<timestamp>.log
   ```

2. **Run specific failing test with verbose output**:
   ```bash
   ./run_specific_test.sh ./scheduler TestJobService -v
   ```

3. **Check for race conditions**:
   ```bash
   go test -race ./...
   ```

### Coverage Not Generating

1. **Ensure tests are passing**:
   ```bash
   ./run_all_tests.sh
   ```

2. **Check coverage output file exists**:
   ```bash
   ls -la tests/reports/coverage_*.out
   ```

3. **Manually generate coverage**:
   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out
   ```

## Reports Directory

The `reports/` directory stores:
- **Coverage data files** (`.out`): Binary coverage data
- **HTML reports** (`.html`): Interactive coverage visualization
- **Test logs** (`.log`): Detailed test execution output
- **Coverage summaries** (`.txt`): Text-based coverage reports

**Note**: This directory is auto-created and should not be committed to version control. Add to `.gitignore`:

```
tests/reports/
```

## Related Documentation

- [Scheduler Tests](../scheduler/) - Job scheduling test details
- [Provider Tests](../providers/) - Provider implementation tests
- [Utility Tests](../utility/) - Utility function tests
- [Root README](../README.md) - Overall project architecture

## Future Enhancements

Planned testing infrastructure improvements:
- Integration tests with real BMC endpoints
- Performance benchmarks
- Automated regression testing
- Test result trending and analytics
- Coverage diff reports (compare coverage between branches)


```bash
./coverage_report.sh -t -p
```

## Best Practices

1. **Run tests before committing:**
   ```bash
   ./run_all_tests.sh -v
   ```

2. **Check coverage regularly:**
   ```bash
   ./coverage_report.sh -p
   ```

3. **Test specific changes:**
   ```bash
   ./run_specific_test.sh ./scheduler -v -c
   ```

4. **Maintain coverage standards:**
   ```bash
   ./coverage_report.sh --threshold 70
   ```

## Troubleshooting

### Permission Denied

If you get "Permission denied" errors:
```bash
chmod +x *.sh
```

### Tests Not Found

Make sure you're running the scripts from the `tests/` directory or use absolute paths.

### Coverage Report Empty

Ensure tests are passing before generating coverage reports. Run:
```bash
./run_all_tests.sh -v
```

## Reports Directory

All generated reports are stored in the `reports/` subdirectory with timestamps:

- `coverage_YYYYMMDD_HHMMSS.out` - Coverage data
- `coverage_YYYYMMDD_HHMMSS.html` - HTML coverage report
- `test_output_YYYYMMDD_HHMMSS.log` - Test execution logs
- `coverage_summary_YYYYMMDD_HHMMSS.txt` - Text coverage summary

Old reports are not automatically deleted, allowing you to track coverage trends over time.

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Go Coverage Tool](https://golang.org/cmd/cover/)
- [Table-Driven Tests in Go](https://github.com/golang/go/wiki/TableDrivenTests)

## Contributing

When adding new tests:
1. Follow Go testing conventions (`*_test.go` files)
2. Use table-driven tests where appropriate
3. Run the full test suite before submitting
4. Ensure coverage doesn't decrease
