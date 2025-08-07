# Development Scripts

This directory contains scripts to help with development and validation of the Terraform provider.

## validate.sh

A comprehensive validation script that runs all the checks that would run in CI, helping catch issues before pushing code.

### Usage

```bash
# Run standard validation (no acceptance tests)
./scripts/validate.sh

# Run validation including acceptance tests
./scripts/validate.sh --acceptance
# or
./scripts/validate.sh -a
```

### What it checks

1. **Go formatting** - Ensures all Go code is properly formatted
2. **go.mod tidiness** - Verifies dependencies are properly managed
3. **go vet** - Runs static analysis
4. **golangci-lint** - Comprehensive linting (if installed)
5. **Unit tests** - Runs all unit tests with race detection and coverage
6. **Build** - Ensures the provider builds successfully
7. **Security scan** - Checks for security issues with gosec (if installed)
8. **Documentation** - Verifies provider docs are up to date (if tfplugindocs installed)
9. **Acceptance tests** - Runs full acceptance tests (optional, requires API credentials)
10. **Uncommitted changes** - Warns about any uncommitted changes

### Installing Optional Tools

The script will work without these tools but will show warnings:

```bash
# Install all development tools at once
make tools

# Or install individually:
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
```

### Environment Variables

For acceptance tests, set these environment variables:
- `UPTIME_API_KEY` - Your Uptime Monitor API key
- `UPTIME_BASE_URL` - The API base URL (defaults to https://api.uptime-monitor.io)

## Git Hooks

The `.githooks/` directory contains Git hooks that can be installed to run validation automatically.

### Installing Hooks

```bash
# Install git hooks
make install-hooks

# Uninstall git hooks
make uninstall-hooks
```

### Available Hooks

- **pre-push** - Runs `validate.sh` before allowing push to remote

To bypass hooks temporarily (not recommended):
```bash
git push --no-verify
```

## Quick Commands

Use the Makefile for common tasks:

```bash
# Show all available commands
make help

# Run validation (same as ./scripts/validate.sh)
make validate

# Run all checks including acceptance tests
make validate-all

# Install git hooks
make install-hooks

# Run specific checks
make fmt        # Format code
make test       # Run tests
make lint       # Run linter
make security   # Run security scan
```