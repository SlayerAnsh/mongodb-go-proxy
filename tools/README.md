# Stress Test Tools

This directory contains stress testing tools for the MongoDB Go Proxy API.

## Go-based Stress Test (Recommended)

The Go-based stress test tool provides detailed metrics and is more efficient for high-concurrency testing.

### Usage

```bash
# Run with default settings (10 concurrent requests for 30 seconds)
make stress-test

# Or run directly
go run tools/stress.go

# Custom options
go run tools/stress.go \
  -url "http://188.245.180.65:8080/api/v1/databases/andromeda-ibc-devnet/collections/ados/documents?limit=5&skip=0" \
  -secret "readonly-super-secret" \
  -c 50 \
  -d 60s

# Run for a specific number of requests
go run tools/stress.go -n 1000 -c 20
```

### Options

- `-url`: URL to stress test (default: the provided endpoint)
- `-secret`: api-secret header value (default: "readonly-super-secret")
- `-c`: Number of concurrent requests (default: 10)
- `-d`: Duration of the test (default: 30s)
- `-n`: Total number of requests (0 = run for duration, default: 0)
- `-timeout`: Request timeout (default: 10s)

### Example Output

```
Starting stress test...
URL: http://188.245.180.65:8080/api/v1/databases/andromeda-ibc-devnet/collections/ados/documents?limit=5&skip=0
Concurrency: 50
Duration: 60s
Request Timeout: 10s

=== Stress Test Results ===
Total Requests:     1250
Successful:         1248 (99.84%)
Failed:             2 (0.16%)

Response Times:
  Average:          2.4s
  Min:              1.2s
  Max:              5.8s

Status Codes:
  200: 1248
  500: 2

Requests per second: 20.83
```

## Bash-based Stress Test

A simpler bash script alternative using curl.

### Usage

```bash
# Run with default settings (10 concurrent requests for 30 seconds)
./tools/stress_test.sh

# Custom concurrency and duration
./tools/stress_test.sh 50 60
```

### Requirements

- `curl`
- `bc` (for calculations)

## Building the Stress Test Tool

```bash
# Build standalone binary
make build-stress-test

# Run the binary
./bin/stress_test -c 100 -d 120s
```

