# Cloud Native PostgreSQL Restic Backup Plugin

A backup plugin for CloudNativePG that implements backup and restore functionality using Restic, supporting S3-compatible storage with Point-in-Time Recovery (PITR) capability.

## Features

- S3-compatible storage support
- Point-in-Time Recovery (PITR) capability
- WAL archiving
- Incremental backups using Restic
- Kubernetes-native deployment

## Prerequisites

- Go 1.21+
- Kubernetes cluster
- CloudNativePG operator installed (v1.26.1)
- Restic
- Docker
- Access to S3-compatible storage

## Development Setup

### Manual Installation Steps

1. Install Kind (if using local development):
   ```bash
   # Download Kind
   curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
   chmod +x ./kind
   sudo mv ./kind /usr/local/bin/kind

   # Create a cluster
   kind create cluster --name cnpg-dev
   ```

2. Run the development setup script:
   ```bash
   chmod +x setup-dev.sh
   ./setup-dev.sh
   ```

3. Source the updated environment:
   ```bash
   source ~/.bashrc
   ```

### Project Structure

```
.
├── cmd
│   └── plugin          # Main plugin executable
├── internal
│   ├── backup         # Backup implementation
│   ├── restore        # Restore implementation
│   └── restic         # Restic client wrapper
├── examples           # Example configurations
└── tests             # Integration tests
```

## Development

### Building

```bash
# Build the plugin binary
make build

# Run tests
make test

# Build Docker image
make docker-build PLUGIN_IMAGE=your-registry/cnpg-restic-plugin
```

### Configuration

The plugin requires the following environment variables:

- `RESTIC_REPOSITORY`: S3 repository URL
- `RESTIC_PASSWORD`: Repository encryption password
- `S3_ACCESS_KEY`: S3 access key
- `S3_SECRET_KEY`: S3 secret key
- `S3_ENDPOINT`: S3 endpoint (optional, defaults to AWS S3)

### Testing

To run integration tests with S3:

```bash
export TEST_S3_ENDPOINT=your-endpoint
export TEST_RESTIC_REPOSITORY=s3:your-bucket/path
export TEST_RESTIC_PASSWORD=your-password
export TEST_AWS_ACCESS_KEY_ID=your-access-key
export TEST_AWS_SECRET_ACCESS_KEY=your-secret-key

make test
```

### Deployment

1. Update the image registry in examples/plugin-config.yaml
2. Deploy the configuration:
   ```bash
   kubectl apply -f examples/plugin-config.yaml
   ```

3. Monitor the plugin:
   ```bash
   make logs
   ```

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.
