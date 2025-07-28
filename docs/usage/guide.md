# Usage Guide

This guide provides detailed information about using the CloudNative PostgreSQL Restic Backup Plugin.

## Prerequisites

### Required Software
- Kubernetes cluster
- CloudNative PostgreSQL operator (v1.26.1+)
- Restic (latest version)
- Access to S3-compatible storage

### Required Permissions
- S3 bucket access
- Kubernetes RBAC permissions
- PostgreSQL superuser access

## Installation

### 1. Deploy the Plugin

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

### 2. Configure CloudNative PostgreSQL

```yaml
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: postgresql-test
  namespace: cnpg-system
spec:
  instances: 1
  
  # PostgreSQL configuration
  postgresql:
    parameters:
      max_connections: "100"
      shared_buffers: "256MB"
      
  # Storage configuration
  storage:
    size: 1Gi
    
  # Backup configuration
  backup:
    target: primary
    custom:
      method: http
      endpointURL: http://localhost:8080
      pluginImage: your-registry/cnpg-restic-plugin:latest
      secretName: restic-backup-config
      env:
        - name: RESTIC_REPOSITORY
          value: "s3:https://your-endpoint/your-bucket/backups"
        - name: AWS_ENDPOINT
          value: "https://your-s3-endpoint"
  
  # WAL archiving configuration
  walArchive:
    custom:
      method: http
      endpointURL: http://localhost:8080
      pluginImage: your-registry/cnpg-restic-plugin:latest
      secretName: restic-backup-config
      env:
        - name: RESTIC_REPOSITORY
          value: "s3:https://your-endpoint/your-bucket/wal-archive"
        - name: AWS_ENDPOINT
          value: "https://your-s3-endpoint"
```

## Configuration

### Plugin Configuration

#### Environment Variables
- `RESTIC_REPOSITORY`: S3 repository URL
- `RESTIC_PASSWORD`: Repository encryption password
- `S3_ENDPOINT`: S3-compatible storage endpoint
- `S3_ACCESS_KEY`: S3 access key
- `S3_SECRET_KEY`: S3 secret key

#### Command-line Flags
- `--listen`: HTTP server listen address (default: `:8080`)
- `--log-level`: Logging level (default: `info`)
- `--log-json`: Enable JSON log format (default: `false`)

### Backup Configuration

#### Full Backups
- Automatically includes current WAL timeline
- Creates consistent backup including all required WAL segments
- Tags backups for easy identification

#### WAL Archiving
- Continuous WAL archiving
- Timeline tracking
- Segment verification

## Operations

### Monitoring

#### Logging
The plugin provides structured logging with the following components:

```json
{
  "level": "info",
  "timestamp": "2025-07-27T15:04:05Z",
  "component": "backup",
  "operation": "create_backup",
  "backup_id": "backup-123",
  "data_dir": "/var/lib/postgresql/data",
  "message": "Starting backup"
}
```

#### Log Levels
- `debug`: Detailed debugging information
- `info`: Normal operation information
- `warn`: Warning conditions
- `error`: Error conditions

### Backup Management

#### Creating Backups
Backups are automatically created according to the schedule defined in the Cluster specification.

#### Listing Backups
View available backups:
```bash
kubectl get backups -n cnpg-system
```

### Restore Operations

#### Full Restore
To restore a full backup:
1. Create a new cluster with restore configuration
2. Specify the backup ID to restore from
3. Monitor restore progress

#### Point-in-Time Recovery
To perform PITR:
1. Identify target recovery time or LSN
2. Configure recovery target in cluster spec
3. Create new cluster with restore configuration

Example PITR configuration:
```yaml
spec:
  bootstrap:
    recovery:
      source: sourceCluster
      recoveryTarget:
        targetTime: "2025-07-27 15:04:05.000000+00"
```

## Troubleshooting

### Common Issues

#### Backup Failures
1. Check S3 credentials
2. Verify repository access
3. Check available storage space
4. Review plugin logs

#### Restore Failures
1. Verify backup existence
2. Check WAL segment availability
3. Verify storage access
4. Review restore logs

### Log Analysis

Example error log:
```json
{
  "level": "error",
  "timestamp": "2025-07-27T15:04:05Z",
  "component": "backup",
  "operation": "create_backup",
  "error": "failed to access S3 bucket",
  "backup_id": "backup-123",
  "message": "Backup failed"
}
```

### Health Checks

The plugin provides health check endpoints:
- `/healthz`: Basic health check
- `/readyz`: Readiness check
- `/metrics`: Prometheus metrics
