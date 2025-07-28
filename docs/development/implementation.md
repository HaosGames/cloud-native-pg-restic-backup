# Implementation Details

This document provides detailed information about the implementation of key components in the CloudNative PostgreSQL Restic Backup Plugin.

## WAL Management

### WAL Segment Structure
```go
type Segment struct {
    Timeline    Timeline    // PostgreSQL timeline number
    LogicalID   uint64      // Logical WAL file number
    SegmentID   uint64      // Segment number within logical WAL file
    Path        string      // Original file path
    BackupID    string      // Associated Restic backup ID
    ArchivedAt  time.Time   // Archive timestamp
}
```

### WAL File Naming
- Format: `XXXXXXXX XXXXXXXX XXXXXXXX`
- Example: `000000010000000000000001`
- Components:
  - Timeline ID (8 hex digits)
  - Logical WAL File ID (8 hex digits)
  - Segment ID (8 hex digits)

### WAL Timeline Management
- Timeline tracking for database forking
- Automatic timeline detection
- Consistent timeline handling during restore

## Backup Operations

### Full Backup Process
1. Initialize backup:
   ```go
   func (h *handlerImpl) CreateBackup(ctx context.Context, dataDir string) error {
       // Get current WAL timeline
       timeline, err := h.walManager.GetWALTimeline(ctx)
       // Create backup with timeline information
       tags := []string{
           "type:full",
           fmt.Sprintf("timeline:%d", timeline),
       }
       return h.client.Backup(ctx, dataDir, tags)
   }
   ```

### WAL Archiving
1. Parse WAL file name
2. Extract timeline and segment information
3. Archive with metadata tags
4. Verify successful storage

## Restore Operations

### Backup Restore Process
1. Validate backup existence
2. Restore full backup
3. Apply WAL segments if PITR requested

### PITR Implementation
1. Identify target timeline
2. Locate required WAL segments
3. Restore segments sequentially
4. Verify WAL continuity

## Storage Operations

### Restic Integration
- Command execution wrapper
- Environment configuration
- Error handling and retries
- Repository management

### S3 Configuration
```go
type Config struct {
    Repository  string   // S3 URL for repository
    Password    string   // Repository encryption password
    S3Endpoint  string   // Custom S3 endpoint (optional)
    S3AccessKey string   // S3 access credentials
    S3SecretKey string   // S3 secret credentials
}
```

## HTTP API

### Backup Endpoint
- Path: `/backup`
- Method: `POST`
- Request:
  ```json
  {
    "backupID": "string",
    "dataFolder": "string",
    "destinationPath": "string"
  }
  ```

### Restore Endpoint
- Path: `/restore`
- Method: `POST`
- Request:
  ```json
  {
    "backupID": "string",
    "destFolder": "string",
    "recoveryTarget": {
      "targetTime": "string",
      "targetXID": "string",
      "targetLSN": "string",
      "targetName": "string",
      "targetInclusive": boolean
    }
  }
  ```

### WAL Archive Endpoint
- Path: `/wal-archive`
- Method: `POST`
- Request:
  ```json
  {
    "walFileName": "string",
    "walFilePath": "string"
  }
  ```

### WAL Restore Endpoint
- Path: `/wal-restore`
- Method: `POST`
- Request:
  ```json
  {
    "walFileName": "string",
    "destFolder": "string"
  }
  ```

## Logging Implementation

### Logger Configuration
```go
type Config struct {
    Level      string    // Log level (debug, info, warn, error)
    JSONOutput bool      // Enable JSON-formatted output
}
```

### Structured Logging
- Component identification
- Operation tracking
- Error context
- Performance metrics

### Log Levels
- DEBUG: Detailed debugging information
- INFO: Normal operation information
- WARN: Warning conditions
- ERROR: Error conditions
- FATAL: Critical conditions

## Error Handling

### Error Categories
1. Configuration Errors
2. Storage Errors
3. WAL Processing Errors
4. Backup/Restore Errors

### Error Recovery
- Automatic retries for transient failures
- Cleanup on partial failures
- Error reporting and logging
- State recovery mechanisms
