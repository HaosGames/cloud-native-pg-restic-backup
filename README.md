# Cloud Native PostgreSQL Restic Backup Plugin

A backup plugin for CloudNativePG that implements backup and restore functionality using Restic, supporting S3-compatible storage with Point-in-Time Recovery (PITR) capability.

## Features

- S3-compatible storage support
- Point-in-Time Recovery (PITR) capability
- WAL archiving with timeline management
- Incremental backups using Restic
- Structured logging
- Kubernetes-native deployment

## Documentation

### Architecture
- [Overview](docs/architecture/overview.md) - High-level architecture and component interaction
- [Components](docs/architecture/overview.md#system-components) - Detailed component descriptions
- [Data Flow](docs/architecture/overview.md#data-flow) - System data flow diagrams

### Development
- [Implementation Details](docs/development/implementation.md) - Detailed implementation documentation
- [Development Guide](docs/development/guide.md) - Guide for developers
- [Code Organization](docs/development/guide.md#code-organization) - Project structure and organization

### Usage
- [Usage Guide](docs/usage/guide.md) - Complete usage documentation
- [Configuration](docs/usage/guide.md#configuration) - Configuration options
- [Operations](docs/usage/guide.md#operations) - Operational procedures
- [Troubleshooting](docs/usage/guide.md#troubleshooting) - Common issues and solutions

## Quick Start

### Prerequisites

- Kubernetes cluster
- CloudNativePG operator installed (v1.26.1+)
- Restic
- Access to S3-compatible storage

### Installation

1. Deploy the configuration secret:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: restic-backup-config
  namespace: cnpg-system
type: Opaque
data:
  RESTIC_PASSWORD: <base64-encoded-password>
  AWS_ACCESS_KEY_ID: <base64-encoded-access-key>
  AWS_SECRET_ACCESS_KEY: <base64-encoded-secret-key>
```

2. Configure the plugin in your PostgreSQL cluster:
```yaml
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: postgresql-test
spec:
  instances: 1
  backup:
    target: primary
    custom:
      method: http
      pluginImage: your-registry/cnpg-restic-plugin:latest
      secretName: restic-backup-config
      env:
        - name: RESTIC_REPOSITORY
          value: "s3:https://your-endpoint/your-bucket/backups"
```

See the [Usage Guide](docs/usage/guide.md) for complete configuration options.

## Development

### Building

```bash
# Build the plugin
make build

# Run tests
make test

# Build Docker image
make docker-build PLUGIN_IMAGE=your-registry/cnpg-restic-plugin
```

See the [Development Guide](docs/development/guide.md) for detailed development instructions.

## Contributing

1. Fork the repository
2. Create your feature branch
3. Make your changes
4. Submit a pull request

Please read our [Development Guide](docs/development/guide.md) before contributing.
