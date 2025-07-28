# Development Guide

This guide provides information for developers working on the CloudNative PostgreSQL Restic Backup Plugin.

## Project Structure

```
.
├── cmd/
│   └── plugin/              # Main application
│       └── main.go         # Entry point
├── docs/
│   ├── architecture/       # Architecture documentation
│   ├── development/       # Development guides
│   └── usage/            # User guides
├── internal/
│   ├── backup/           # Backup implementation
│   ├── logging/          # Logging framework
│   ├── plugin/           # Plugin HTTP interface
│   ├── restic/           # Restic client
│   ├── restore/          # Restore implementation
│   └── wal/             # WAL management
├── examples/            # Example configurations
└── tests/              # Integration tests
```

## Code Organization

### Main Components

1. Plugin Interface (`internal/plugin/`)
   - HTTP request handling
   - Request/response types
   - Operation coordination

2. Backup Handler (`internal/backup/`)
   - Full backup implementation
   - WAL archiving integration
   - Backup metadata management

3. Restore Handler (`internal/restore/`)
   - Backup restoration
   - PITR implementation
   - WAL restoration

4. WAL Manager (`internal/wal/`)
   - WAL segment handling
   - Timeline management
   - Segment cleanup

5. Restic Client (`internal/restic/`)
   - Restic command execution
   - Repository management
   - S3 integration

6. Logging Framework (`internal/logging/`)
   - Structured logging
   - Operation tracking
   - Error reporting

## Development Workflow

### Setting Up Development Environment

1. Install required tools:
   ```bash
   ./setup-dev.sh
   ```

2. Configure test environment:
   ```bash
   export TEST_S3_ENDPOINT=your-endpoint
   export TEST_RESTIC_REPOSITORY=s3:your-bucket/path
   export TEST_RESTIC_PASSWORD=your-password
   export TEST_AWS_ACCESS_KEY_ID=your-access-key
   export TEST_AWS_SECRET_ACCESS_KEY=your-secret-key
   ```

### Building

```bash
# Build binary
make build

# Run tests
make test

# Build Docker image
make docker-build
```

### Testing

#### Unit Tests
Run unit tests with:
```bash
go test ./...
```

#### Integration Tests
Run integration tests with:
```bash
make test
```

#### Manual Testing
1. Deploy to Kubernetes:
   ```bash
   kubectl apply -f examples/plugin-config.yaml
   ```

2. Monitor logs:
   ```bash
   make logs
   ```

### Code Style

#### Go Guidelines
- Follow standard Go formatting (use `gofmt`)
- Use meaningful variable names
- Add comments for exported types and functions
- Implement proper error handling
- Use context for cancellation

#### Error Handling
```go
// Good
if err := doSomething(); err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Bad
if err := doSomething(); err != nil {
    return err
}
```

#### Logging
```go
// Good
logger.WithFields(map[string]interface{}{
    "operation": "backup",
    "backup_id": backupID,
}).Info("Starting backup")

// Bad
logger.Printf("Starting backup %s", backupID)
```

### Adding New Features

1. Create feature branch:
   ```bash
   git checkout -b feature/new-feature
   ```

2. Implement changes:
   - Add tests first
   - Implement feature
   - Update documentation
   - Run tests

3. Submit pull request:
   - Clear description
   - List of changes
   - Test results
   - Documentation updates

## Testing Guidelines

### Unit Test Structure
```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "result",
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Integration Test Requirements
- S3-compatible storage access
- PostgreSQL instance
- Kubernetes cluster
- CloudNative PostgreSQL operator

### Mocking
```go
// Example mock
type mockHandler struct {
    createBackupErr error
}

func (m *mockHandler) CreateBackup(ctx context.Context, path string) error {
    return m.createBackupErr
}
```

## Documentation

### Code Comments
- Add package documentation
- Document exported types and functions
- Explain complex algorithms
- Include examples for usage

### Documentation Updates
- Update architecture docs for design changes
- Update implementation docs for new features
- Update usage guide for new functionality
- Keep examples current

## Troubleshooting

### Common Issues

1. Build Failures
   - Check Go version
   - Verify dependencies
   - Check import paths

2. Test Failures
   - Check S3 configuration
   - Verify test environment
   - Review test logs

3. Runtime Issues
   - Check logs
   - Verify permissions
   - Monitor resource usage

### Debugging

1. Enable Debug Logging
   ```bash
   ./plugin --log-level=debug
   ```

2. Use Test Environment
   ```bash
   make test-restic
   ```

3. Check Connectivity
   ```bash
   restic -r s3:your-bucket check
   ```
