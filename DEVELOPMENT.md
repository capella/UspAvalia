# Development Guide - UspAvalia

## Overview

This project follows **Go best practices for 2025** with modern tooling, automated quality checks, and streamlined development workflows.

## Quick Start

```bash
# Install dependencies
make deps

# Run quality checks
make audit

# Run tests
make test

# Build and run
make run
```

## Development Tools

### Makefile Commands

We use a Makefile to standardize development tasks across the team. Run `make help` to see all available commands:

**Quality Control:**
- `make audit` - Run all quality checks (format, vet, staticcheck, test)
- `make fmt` - Format all Go code with gofmt
- `make vet` - Run go vet static analysis
- `make staticcheck` - Run staticcheck analyzer
- `make vulncheck` - Check for known security vulnerabilities
- `make tidy` - Tidy and verify module dependencies

**Testing:**
- `make test` - Run all tests with race detection
- `make test/cover` - Run tests with coverage report (generates `coverage.html`)

**Building:**
- `make build` - Build the binary to `./bin/uspavalia`
- `make run` - Build and run the application
- `make clean` - Remove build artifacts

**Docker:**
- `make docker/build` - Build Docker image
- `make docker/run` - Run Docker container

**Database:**
- `make db/migrate` - Run database migrations
- `make db/fetch-disciplines` - Fetch disciplines from Jupiter Web
- `make db/fetch-units` - Fetch teaching units from Jupiter Web

**Production:**
- `make production/deploy` - Deploy to production (requires confirmation)

## Code Quality Standards

### Automated Checks (CI/CD)

Every push and pull request automatically runs:

1. **Format Check** - Ensures code is formatted with `gofmt`
2. **Go Vet** - Static analysis to find common bugs
3. **Staticcheck** - Official Go static analyzer for best practices
4. **Security Scan** - `govulncheck` for known vulnerabilities
5. **Tests** - Unit tests with race detection and coverage
6. **Build** - Verifies the project builds successfully
7. **Dependency Verification** - Ensures `go.mod` and `go.sum` are in sync

### Tools Used

**Built-in Go Tools:**
- `gofmt` - Code formatting (official Go formatter)
- `go vet` - Static analysis (detects common bugs)
- `go test` - Testing with `-race` flag for race condition detection

**Official Go Tools:**
- **[staticcheck](https://staticcheck.dev/)** - Official Go static analyzer
- **[govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)** - Official vulnerability scanner

### Why These Tools?

Based on **[Go best practices 2025](https://www.bacancytechnology.com/blog/go-best-practices)**, we use:

1. **Built-in tooling first** - Go's official tools are battle-tested and maintained
2. **Staticcheck** - Recommended by the Go team, catches bugs `go vet` misses
3. **No golangci-lint** - Removed due to version compatibility issues with Go 1.25
4. **Security scanning** - `govulncheck` detects known vulnerabilities in dependencies

## Testing Best Practices

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test/cover

# Run specific package
go test -v ./internal/handlers/

# Run specific test
go test -v -run TestMagicLinkAuth ./internal/handlers/
```

### Writing Tests

Follow **[table-driven testing patterns](https://medium.com/@nandoseptian/testing-go-code-like-a-pro-what-i-wish-i-knew-starting-out-2025-263574b0168f)**:

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case 1", "input1", "output1"},
        {"case 2", "input2", "output2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := SomeFunction(tt.input)
            if got != tt.expected {
                t.Errorf("got %v, want %v", got, tt.expected)
            }
        })
    }
}
```

## GitHub Actions CI/CD

### Workflows

**`.github/workflows/lint.yml`** - Quality & Tests
- Runs on every push and pull request
- Parallel execution of all quality checks
- Uploads coverage to Codecov

**`.github/workflows/deploy_docker.yml`** - Production Deployment
- Runs on pushes to `master` branch
- Builds Docker image
- Pushes to Google Container Registry
- Triggers deployment

### Best Practices Implemented

Based on **[GitHub Actions best practices 2025](https://dev.to/ticatwolves/automate-your-go-project-best-practices-cicd-with-github-actions-4bo4)**:

✅ **Caching** - Go module cache enabled for faster builds
✅ **Parallel jobs** - All checks run concurrently
✅ **Latest actions** - Uses `actions/checkout@v4` and `actions/setup-go@v5`
✅ **Security** - `govulncheck` scans for vulnerabilities
✅ **Coverage tracking** - Codecov integration
✅ **Build verification** - Ensures `go build` succeeds

## Project Structure

```
.
├── cmd/                    # CLI commands (Cobra-based)
├── internal/               # Private application code
│   ├── config/            # Configuration management
│   ├── database/          # Database setup
│   ├── handlers/          # HTTP handlers
│   ├── middleware/        # HTTP middleware
│   ├── models/            # Database models (GORM)
│   └── services/          # Business logic services
├── pkg/                   # Public packages
├── templates/             # HTML templates
├── static/                # Static assets
├── .github/workflows/     # CI/CD workflows
├── Makefile              # Development tasks
└── go.mod                # Go module definition
```

## Code Review Checklist

Before submitting a PR, ensure:

- [ ] `make audit` passes (format, vet, staticcheck, tests)
- [ ] `make vulncheck` passes (no vulnerabilities)
- [ ] New tests added for new functionality
- [ ] Code follows Go idioms (see [Effective Go](https://go.dev/doc/effective_go))
- [ ] No commented-out code
- [ ] Meaningful commit messages

## Local Development

### Prerequisites

- Go 1.25 or later
- Make
- Docker (optional, for containerized development)

### First-Time Setup

```bash
# Clone the repository
git clone <repo-url>
cd UspAvalia

# Install dependencies
make deps

# Run database migrations
make db/migrate

# Fetch initial data (optional)
make db/fetch-units
make db/fetch-disciplines

# Build and run
make run
```

### Configuration

Configuration can be provided via:
1. `.uspavalia.yaml` file in the current directory
2. Environment variables with `USPAVALIA_` prefix

See `.uspavalia.yaml` for all available options.

### Development Mode

Enable dev mode for debugging:

```yaml
# .uspavalia.yaml
dev_mode: true
```

When enabled:
- CSRF protection is disabled
- Email content is logged to console
- Additional debug logging

**⚠️ Never use dev_mode in production!**

## Resources

### Go Best Practices (2025)
- [Go Best Practices by Bacancy](https://www.bacancytechnology.com/blog/go-best-practices)
- [Testing Go Code Like a Pro](https://medium.com/@nandoseptian/testing-go-code-like-a-pro-what-i-wish-i-knew-starting-out-2025-263574b0168f)
- [Go Linting Best Practices for CI/CD](https://medium.com/@tedious/go-linting-best-practices-for-ci-cd-with-github-actions-aa6d96e0c509)

### GitHub Actions & CI/CD
- [Automate Your Go Project with GitHub Actions](https://dev.to/ticatwolves/automate-your-go-project-best-practices-cicd-with-github-actions-4bo4)
- [Continuous Integration with Go and GitHub Actions](https://www.alexedwards.net/blog/ci-with-go-and-github-actions)

### Makefile Best Practices
- [Time-saving Makefile for Go Projects](https://www.alexedwards.net/blog/a-time-saving-makefile-for-your-go-projects)
- [Ultimate Makefile for Golang](https://www.mohitkhare.com/blog/go-makefile/)

### Official Go Resources
- [Effective Go](https://go.dev/doc/effective_go)
- [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Testing Package Documentation](https://pkg.go.dev/testing)

## Support

For questions or issues, please open an issue on GitHub.
