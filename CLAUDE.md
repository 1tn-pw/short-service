# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a Go-based URL shortening microservice that provides gRPC APIs for creating and managing short URLs. The service extracts webpage metadata (title, favicon, description) when creating short URLs and stores them in MongoDB.

## Build and Development Commands

### Using Make (preferred):
- `make setup` - Install required development tools (golangci-lint, goimports, protoc plugins)
- `make fmt` - Format code with gofmt and goimports
- `make lint` - Run golangci-lint
- `make test` - Run tests with race detection and coverage
- `make build` - Build container images for multiple platforms
- `make deploy` - Deploy to Kubernetes cluster
- `make clean` - Clean up build artifacts

### Using Task (alternative):
- `task setup` - Install development tools
- `task fmt` - Format and tidy code
- `task test` - Run tests
- `task build` - Build container images
- `task deploy` - Deploy to Kubernetes
- `task clean` - Clean artifacts

## Architecture

### Service Structure
```
cmd/short-service/    # Application entry point
internal/
├── service/         # Service initialization, configuration, and server setup
└── short/           # Core business logic and gRPC service implementation
k8s/                 # Kubernetes manifests
```

### Key Components
- **gRPC Service**: Implements URL shortening APIs defined in external protobufs (`github.com/1tn-pw/protobufs`)
- **HTTP Server**: Health check endpoints on port 80
- **MongoDB**: Primary data store for URLs and metadata
- **Vault Integration**: Secrets management for configuration
- **Dual Server**: Runs both HTTP (health) and gRPC (API) servers concurrently

### External Dependencies
- Protocol buffer definitions from `github.com/1tn-pw/protobufs/go/short`
- MongoDB for persistence
- Vault for secrets management
- Custom container registry at `containers.chewed-k8s.net`

## Development Guidelines

### Code Organization
- All business logic resides in `internal/short/`
- Service initialization and configuration in `internal/service/`
- Main application entry in `cmd/short-service/`

### Configuration
- Environment variables are primary configuration method
- Vault integration for sensitive values
- Configuration struct defined in `internal/service/config.go`

### Testing
- Run tests with race detection: `make test`
- Currently no active unit tests for gRPC service (main test file commented out)
- Tests should cover gRPC service methods in `internal/short/`

### Deployment
- Kubernetes deployment with 2 replicas
- Container images built for amd64 and arm64
- Uses custom container registry: `containers.chewed-k8s.net`
- Secrets managed via Kubernetes secrets

### CI/CD
- GitHub Actions automate testing and deployment
- Pull requests trigger test runs
- Dependabot manages dependency updates
- Release drafts created automatically