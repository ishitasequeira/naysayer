# NAYSAYER Makefile

.PHONY: build run test test-coverage clean install help docker fmt vet lint lint-fix e2e e2e-examples e2e-performance e2e-report

# Default target
help:
	@echo "NAYSAYER Build Commands:"
	@echo ""
	@echo "Build & Run:"
	@echo "  build          Build the naysayer binary"
	@echo "  run            Build and run the server"
	@echo "  docker         Build Docker image"
	@echo ""
	@echo "Testing:"
	@echo "  test           Run all unit tests"
	@echo "  test-coverage  Generate test coverage report"
	@echo "  e2e            Run all end-to-end tests"
	@echo "  e2e-examples   Run E2E example tests only"
	@echo "  e2e-performance Run E2E performance tests only"
	@echo "  e2e-report     Run E2E tests with detailed report generation"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint           Run golangci-lint"
	@echo "  lint-fix       Run golangci-lint with automatic fixes"
	@echo "  fmt            Format code with gofmt"
	@echo "  vet            Run go vet"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean          Remove built binaries and coverage files"
	@echo "  install        Install dependencies"
	@echo ""

# Build the binary
build: lint fmt vet test
	@echo "Building naysayer..."
	go mod download && go mod tidy && go mod vendor
	go build -o naysayer cmd/main.go
	@echo "✅ Built naysayer binary"

# Build and run
run: build
	@echo "Starting naysayer server..."
	./naysayer

# Run unit tests
test:
	@echo "Running unit tests..."
	go test ./... -v -race -cover

# Generate test coverage report
test-coverage:
	@echo "Generating test coverage report..."
	@mkdir -p coverage
	go test ./... -coverprofile=coverage/coverage.out -covermode=atomic
	@echo "📊 Coverage Summary:"
	go tool cover -func=coverage/coverage.out | tail -1
	@echo "✅ Coverage report completed"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "✅ go vet completed"

# Run linter
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run --skip-dirs=vendor ./...; \
		echo "✅ Linting completed"; \
	else \
		echo "⚠️  golangci-lint not installed. Install with:"; \
		echo "   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.54.2"; \
	fi

# Run linter with automatic fixes
lint-fix:
	@echo "Running golangci-lint with automatic fixes..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run --fix --skip-dirs=vendor ./...; \
		echo "✅ Linting with fixes completed"; \
	else \
		echo "⚠️  golangci-lint not installed. Install with:"; \
		echo "   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.54.2"; \
	fi

# Clean built files and coverage files
clean:
	@echo "Cleaning..."
	rm -f naysayer
	rm -rf coverage/
	@echo "✅ Cleaned"

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download
	@echo "✅ Dependencies installed"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t quay.io/redhat-data-and-ai/naysayer:latest .
	@echo "✅ Docker image built: quay.io/redhat-data-and-ai/naysayer:latest"


docker-push:
	docker push quay.io/redhat-data-and-ai/naysayer:latest
	@echo "✅ Docker image pushed: quay.io/redhat-data-and-ai/naysayer:latest"

# End-to-End Testing Targets

# Run all E2E tests
e2e:
	@echo "Running all end-to-end tests..."
	go test ./e2e -v -timeout 10m
	@echo "✅ E2E tests completed"

# Run E2E example tests only
e2e-examples:
	@echo "Running E2E example tests..."
	go test ./e2e -v -run "TestE2E_Examples" -timeout 5m
	@echo "✅ E2E example tests completed"

# Run E2E performance tests only
e2e-performance:
	@echo "Running E2E performance tests..."
	go test ./e2e -v -run "Performance" -timeout 15m
	@echo "✅ E2E performance tests completed"

# Run E2E tests with detailed report generation
e2e-report:
	@echo "Running E2E tests with detailed reporting..."
	@mkdir -p reports
	go test ./e2e -v -timeout 10m -args -generate-report -report-path=reports/e2e_report.json
	@if [ -f reports/e2e_report.json ]; then \
		echo "📊 E2E test report generated: reports/e2e_report.json"; \
	fi
	@echo "✅ E2E tests with reporting completed"

# Run E2E tests filtered by tags
e2e-warehouse:
	@echo "Running warehouse-related E2E tests..."
	go test ./e2e -v -run "Warehouse" -timeout 5m
	@echo "✅ Warehouse E2E tests completed"

e2e-service-account:
	@echo "Running service account E2E tests..."
	go test ./e2e -v -run "Service_Account" -timeout 5m
	@echo "✅ Service account E2E tests completed"

e2e-documentation:
	@echo "Running documentation E2E tests..."
	go test ./e2e -v -run "Documentation" -timeout 5m
	@echo "✅ Documentation E2E tests completed"

# Run E2E tests with coverage
e2e-coverage:
	@echo "Running E2E tests with coverage..."
	@mkdir -p coverage
	go test ./e2e -v -timeout 10m -coverprofile=coverage/e2e_coverage.out -covermode=atomic
	@echo "📊 E2E Coverage Summary:"
	go tool cover -func=coverage/e2e_coverage.out | tail -1
	@echo "✅ E2E coverage report completed"

# Benchmark E2E tests
e2e-bench:
	@echo "Running E2E benchmarks..."
	go test ./e2e -v -bench=. -benchmem -timeout 10m
	@echo "✅ E2E benchmarks completed"

# Validate E2E test scenarios
e2e-validate:
	@echo "Validating E2E test scenarios..."
	go test ./e2e -v -run "TestE2E_CustomTestRunner" -timeout 2m
	@echo "✅ E2E scenario validation completed"
