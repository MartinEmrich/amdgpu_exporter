# AMD GPU Exporter - Agent Guidelines

## Project Overview
Go-based Prometheus exporter for AMD GPU metrics using `amdgpu_top`. Exposes metrics at `:9042/metrics` and health check at `:9042/health`.

## Build & Run Commands

```bash
# Build the binary
go build -o amdgpu_exporter main.go

# Run directly
go run main.go

# Check code formatting
gofmt -l .

# Format code
gofmt -w .

# Run tests
go test ./...

# Run single test file
go test -v <test_file>.go

# Run specific test function
go test -v -run Test<FunctionName>
```

## Code Style Guidelines

### Imports
- Group imports: standard library first, then third-party, then local
- No unused imports (will fail `go build`)
- Use short import paths where appropriate

### Formatting
- Use `gofmt` for all formatting (default Go style)
- 4-space indentation (no tabs in this codebase)
- Max line length: ~120 characters
- Empty lines between logical blocks

### Types
- Define types at package level with clear names
- Use struct tags for JSON serialization: `` `json:"field_name"` ``
- Prefer named types over type aliases for clarity
- Return `(T, error)` pattern for functions that may fail

### Naming Conventions
- Packages: lowercase, no underscores (e.g., `main`)
- Types/Structs: PascalCase (e.g., `GPUActivity`, `MetricValue`)
- Functions: camelCase (e.g., `fetchGPUMetrics`, `formatPrometheusMetric`)
- Constants: camelCase, prefixed with context (e.g., `port`)
- Private functions/types: lowercase first letter
- Variables: descriptive names, avoid single letters except loop counters

### Error Handling
- Always check and handle errors immediately
- Wrap errors with context using `fmt.Errorf("failed to %s: %w", operation, err)`
- Return early on error with clear messages
- Log fatal errors with `log.Fatalf` in main()
- Use `http.Error` for HTTP error responses

### Logging
- Use `log.Printf` for informational messages
- Use `log.Fatalf` for unrecoverable errors
- Include context in log messages (what operation failed, why)

### HTTP Handlers
- Set Content-Type header explicitly
- Use appropriate status codes (200 OK, 500 Internal Server Error)
- Return early on error conditions
- JSON responses: `application/json`, Prometheus metrics: `text/plain; charset=utf-8`

### Prometheus Metrics
- Prefix all metrics with `amdgpu_`
- Use gauge type for measurements
- Include HELP and TYPE comments
- Add device/ASIC labels to all metrics
- Use snake_case for metric names
- Units in name (e.g., `_percent`, `_mb`, `_watts`, `_celsius`, `_mhz`)

### Constants
- Define configuration constants at package level
- Use typed constants where appropriate

### Code Organization
- Keep functions focused and small (<50 lines preferred)
- Group related types together
- Put type definitions before function implementations
- Main() should be last in the file

## Architecture Notes
- Single-file main.go structure
- Uses `amdgpu_top -d -gm -J` for JSON output parsing
- No external dependencies beyond stdlib
- Port 9042 is hardcoded (change via const if needed)
- Graceful shutdown not implemented (SIGINT/SIGTERM will exit immediately)

## Testing Guidelines
- Mock `amdgpu_top` output for unit tests
- Test JSON unmarshaling with valid/invalid data
- Test metric formatting edge cases
- Verify label escaping in Prometheus output
- Check error paths for command failures