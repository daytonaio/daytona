# Storage Recovery Implementation Guide

## Overview

This document describes the storage recovery feature implementation for Daytona sandboxes that have reached their storage limits.

## Architecture

### Runner-Side Implementation

The recovery logic is implemented in the runner at:

- **Core Logic**: `/apps/runner/pkg/docker/recovery.go`
- **Error Detection**: `/apps/runner/pkg/docker/start.go` and `/apps/runner/pkg/docker/create.go`
- **API Endpoint**: `/apps/runner/pkg/api/controllers/sandbox.go` (`RecoverStorage`)
- **Route**: `POST /sandboxes/{sandboxId}/recover-storage`

### Key Components

#### 1. Storage Tracking (`apps/runner/pkg/api/dto/sandbox.go`)

```go
type CreateSandboxDTO struct {
    StorageQuota          int64  // Original storage quota in GB
    StorageExpansionQuota int64  // Additional storage added through recovery in GB
    // ... other fields
}
```

Container labels store these values:

- `daytona.storage_quota`: Original quota
- `daytona.storage_expansion_quota`: Cumulative expansion

#### 2. Recovery Process

When a sandbox hits storage limits:

1. **Detection**: Errors containing "no space left on device" are caught in `ContainerStart` or `ContainerCreate`
2. **Error Message**: Returns specific error: `"no space left on device - storage limit reached. Run recover-storage to expand storage"`
3. **State**: Sandbox enters `ERROR` state with the above error reason

#### 3. Recovery Execution

When `POST /sandboxes/{sandboxId}/recover-storage` is called:

1. **Validation**:
   - Retrieves current storage quotas from container labels
   - Checks if expansion limit (10% of original quota) has been reached

2. **Expansion Calculation**:
   - Adds 100MB (0.1GB) per recovery attempt
   - Maximum: Original quota + 10% of original quota
   - Example: 10GB sandbox can expand to max 11GB (10 recovery attempts of 100MB each)

3. **Container Recreation**:
   - Stops original container
   - Renames it to `{sandboxId}-old`
   - Creates new container with:
     - Same configuration (env, labels, mounts, resources)
     - Expanded storage: `storage-opt size={originalQuota + expansionQuota}G`
   - Starts new container
   - Attempts data migration from old container's overlay2 layer
   - Removes old container

4. **Result**:
   - Success: Container running with expanded storage
   - Failure: Original container renamed back, error returned

## User Notification Strategy

### 1. Dashboard UI Integration

#### Error Display

When a sandbox is in `ERROR` state with `errorReason` containing "no space left on device":

**Visual Indicators**:

```
Status Badge: ⚠️ Storage Limit Reached
Color: Orange/Yellow (recoverable error)
```

**Error Message Display**:

```
⚠️ Storage Limit Reached

Your sandbox has run out of disk space and cannot start. You can recover
it by expanding the storage quota (temporary emergency storage).

Current: {originalQuota}GB
Available Emergency Storage: {maxExpansion - currentExpansion}GB

[Recover Storage] [View Details]
```

#### Dashboard Implementation Points

**Location**: `/apps/dashboard/src/pages/Sandboxes.tsx` and sandbox detail views

**Detection Logic**:

```typescript
const isStorageLimitError = (sandbox: SandboxDto) => {
  return sandbox.state === 'error' && 
         sandbox.errorReason?.toLowerCase().includes('no space left on device');
};
```

**UI Components Needed**:

1. **Storage Error Banner**: Persistent banner in sandbox detail view
2. **Recovery Button**: Primary action button to trigger recovery
3. **Storage Usage Widget**: Show current usage, quota, and available expansion
4. **Recovery History**: Log of past recovery attempts

**Recovery Button Handler**:

```typescript
const handleRecoverStorage = async (sandboxId: string) => {
  try {
    // Call runner API through API service
    await runnerAdapter.recoverSandboxStorage(sandboxId);
    
    // Show success notification
    toast.success('Storage expanded successfully. Sandbox is recovering...');
    
    // Refresh sandbox state
    await refreshSandbox(sandboxId);
  } catch (error) {
    if (error.message.includes('expansion limit reached')) {
      // Show permanent storage limit message
      toast.error('Maximum emergency storage reached. Please clean up files or create a new sandbox.');
    } else {
      toast.error(`Recovery failed: ${error.message}`);
    }
  }
};
```

### 2. SDK Integration

#### TypeScript SDK (`libs/sdk-typescript/src/Sandbox.ts`)

**Error Enhancement in `start()` method**:

```typescript
public async start(): Promise<void> {
  try {
    await this.sandboxApi.startSandbox(this.id);
    await this.waitUntilStarted();
  } catch (error) {
    // Check for storage limit error
    if (error?.response?.data?.message?.includes('no space left on device')) {
      throw new DaytonaError(
        'Sandbox storage limit reached. Use sandbox.recoverStorage() to expand storage quota.',
        'STORAGE_LIMIT_REACHED'
      );
    }
    throw error;
  }
}
```

**New Recovery Method**:

```typescript
/**
 * Recover sandbox from storage limit by expanding storage quota
 * Adds 100MB of emergency storage (max 10% of original quota)
 */
public async recoverStorage(): Promise<void> {
  try {
    // Call runner API recover-storage endpoint through API adapter
    const runner = await this.getRunner();
    await runner.recoverSandboxStorage(this.id);
    
    // Wait for sandbox to recover
    await this.refreshData();
    
    // Optionally wait for it to start
    if (this.state === 'stopped' || this.state === 'error') {
      await this.start();
    }
  } catch (error) {
    if (error?.message?.includes('expansion limit reached')) {
      throw new DaytonaError(
        'Maximum storage expansion limit reached (10% of original quota). ' +
        'Please clean up files or migrate to a larger sandbox.',
        'STORAGE_EXPANSION_LIMIT_REACHED'
      );
    }
    throw new DaytonaError(`Storage recovery failed: ${error.message}`);
  }
}
```

**Usage Example**:

```typescript
try {
  await sandbox.start();
} catch (error) {
  if (error.code === 'STORAGE_LIMIT_REACHED') {
    console.log('Storage limit reached, attempting recovery...');
    await sandbox.recoverStorage();
    console.log('Recovery successful!');
  }
}
```

#### Python SDK (`libs/sdk-python/src/daytona/_sync/sandbox.py`)

**Error Enhancement in `start()` method**:

```python
def start(self) -> None:
    try:
        self._sandbox_api.start_sandbox(self.id)
        self.wait_until_started()
    except Exception as e:
        error_msg = str(e).lower()
        if 'no space left on device' in error_msg:
            raise DaytonaError(
                'Sandbox storage limit reached. Use sandbox.recover_storage() '
                'to expand storage quota.',
                error_code='STORAGE_LIMIT_REACHED'
            ) from e
        raise
```

**New Recovery Method**:

```python
def recover_storage(self) -> None:
    """
    Recover sandbox from storage limit by expanding storage quota.
    Adds 100MB of emergency storage (max 10% of original quota).
    
    Raises:
        DaytonaError: If recovery fails or expansion limit reached
    """
    try:
        # Call runner API recover-storage endpoint
        runner = self._get_runner()
        runner.recover_sandbox_storage(self.id)
        
        # Refresh sandbox data
        self.__refresh_data_safe()
        
        # Optionally restart if stopped
        if self.state in ['stopped', 'error']:
            self.start()
            
    except Exception as e:
        error_msg = str(e).lower()
        if 'expansion limit reached' in error_msg:
            raise DaytonaError(
                'Maximum storage expansion limit reached (10% of original quota). '
                'Please clean up files or migrate to a larger sandbox.',
                error_code='STORAGE_EXPANSION_LIMIT_REACHED'
            ) from e
        raise DaytonaError(f'Storage recovery failed: {str(e)}') from e
```

**Usage Example**:

```python
try:
    sandbox.start()
except DaytonaError as e:
    if e.error_code == 'STORAGE_LIMIT_REACHED':
        print('Storage limit reached, attempting recovery...')
        sandbox.recover_storage()
        print('Recovery successful!')
```

### 3. API Service Layer (Future - Not Implemented Yet)

When exposing through the main API:

**Endpoint**: `POST /api/sandbox/{sandboxId}/recover-storage`

**Flow**:

```
User/SDK -> API Service -> Runner Adapter -> Runner API (recover-storage)
```

**Error Handling**:

- Runner unavailable: Return 503 with retry suggestion
- Expansion limit reached: Return 409 with cleanup instructions
- Unknown error: Return 500 with support contact

### 4. Proactive Warnings (Future - OTEL Integration)

Not implemented in current PR, but outlined for future:

**Warning Triggers**:

- Storage usage > 85%: Warning notification
- Storage usage > 95%: Critical notification
- Storage expanded: Info notification with remaining emergency storage

**OTEL Metrics**:

```
sandbox_storage_usage_percent{sandbox_id, org_id}
sandbox_storage_expansion_count{sandbox_id, org_id}
sandbox_storage_recovery_attempts{sandbox_id, org_id, status="success|failure"}
```

## Error Messages Reference

### Storage Limit Reached (Recoverable)

```
no space left on device - storage limit reached. Run recover-storage to expand storage
```

- **User Action**: Click "Recover Storage" button or call `sandbox.recoverStorage()`
- **System Action**: Sandbox in ERROR state, awaiting recovery

### Expansion Limit Reached (Not Recoverable)

```
storage expansion limit reached: already expanded by XGB, maximum is YGB (10% of ZGB)
```

- **User Action**: Clean up files, delete unused data, or create new sandbox
- **System Action**: Recovery API returns 400 error

### Recovery Success

```
Sandbox storage recovered
```

- **System Action**: Container recreated with expanded storage, sandbox restarting

## Testing the Feature

### Manual Testing

1. **Create test sandbox** with small storage quota:

   ```bash
   # Set storage to 1GB for easy testing
   daytona sandbox create --storage 1
   ```

2. **Fill the storage**:

   ```bash
   daytona sandbox ssh
   dd if=/dev/zero of=/large_file bs=1M count=900
   # This will eventually fail with "no space left"
   ```

3. **Observe error state**:
   - Dashboard shows storage error banner
   - SDK `start()` throws `STORAGE_LIMIT_REACHED`

4. **Trigger recovery**:

   ```bash
   # Via API
   curl -X POST http://runner:8080/sandboxes/{id}/recover-storage
   
   # Via SDK
   sandbox.recoverStorage()
   ```

5. **Verify expansion**:

   ```bash
   # Check new storage size
   df -h /
   # Should show 1.1GB total
   ```

6. **Test limit**:
   - Repeat recovery 10 times to hit 10% expansion limit
   - 11th attempt should fail with "expansion limit reached"

## Migration Notes

### Existing Sandboxes

Sandboxes created before this feature:

- Will have `StorageExpansionQuota = 0`
- Recovery will work, using current `StorageQuota` as original
- Labels will be added on first recovery

### Future Enhancements

1. **Configurable Limits**: Allow org-level configuration of max expansion percentage
2. **Automatic Recovery**: Auto-trigger recovery on first storage error (with user notification)
3. **Storage Analytics**: Dashboard showing storage trends, cleanup suggestions
4. **Data Migration**: Improved overlay2 data copy mechanism using host filesystem access
5. **MicroVM Support**: Implement recovery for non-Docker runners (Firecracker, etc.)

## Security Considerations

1. **Resource Limits**: 10% expansion prevents unbounded growth
2. **Audit Logging**: All recovery attempts logged with user/org context
3. **Rate Limiting**: Consider adding rate limits to prevent recovery abuse
4. **Quota Enforcement**: Original quota stored in immutable label to prevent tampering

## Support and Troubleshooting

### Common Issues

**Q: Recovery fails with "failed to create new container"**

- A: Check Docker daemon logs, ensure XFS filesystem for storage-opt support

**Q: Data lost after recovery**

- A: Overlay2 data copy requires host access; manual recovery may be needed

**Q: Recovery succeeds but still shows storage full**

- A: Files may be held open by processes; restart sandbox after cleanup

**Q: Can't recover: expansion limit reached**

- A: User must clean up files or create new sandbox with larger storage quota
