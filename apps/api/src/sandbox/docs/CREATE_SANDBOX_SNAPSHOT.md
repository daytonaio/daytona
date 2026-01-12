# Create Snapshot from Sandbox

This document describes the implementation of the `POST /sandbox/:sandboxIdOrName/snapshot` endpoint, which creates a reusable snapshot from an existing sandbox.

## Overview

The Create Sandbox Snapshot feature allows users to capture the current state of a sandbox's filesystem and save it as a new snapshot. This snapshot can then be used to create new sandboxes with the same state.

## API Endpoint

```
POST /sandbox/:sandboxIdOrName/snapshot
```

### Request Body

```typescript
{
  "name": string,    // Required: Name for the new snapshot (e.g., "my-dev-env-v1")
  "live"?: boolean   // Optional: Use live mode (default: false)
}
```

### Response

Returns the sandbox DTO. The snapshot creation job runs asynchronously.

```typescript
{
  "id": "sandbox-uuid",
  "name": "my-sandbox",
  "state": "started",
  // ... other sandbox fields
}
```

## Snapshot Modes

| Mode | `live` value | Behavior | Consistency | Use Case |
|------|-------------|----------|-------------|----------|
| Safe (default) | `false` | Pauses VM during disk flatten | Consistent | Production snapshots |
| Optimistic | `true` | Uses `--force-share` to read disk while VM runs | May be inconsistent | Quick iterations |

### Safe Mode (Recommended)

1. VM is paused
2. Disk is flattened to a standalone qcow2 image
3. VM is resumed immediately after flatten
4. Flattened image is uploaded to S3
5. Snapshot entity is created in database

**Downtime**: Only during disk flattening (typically seconds to minutes depending on disk size)

### Live Mode

1. Disk is read using `qemu-img convert -U` while VM continues running
2. Image is uploaded to S3
3. Snapshot entity is created in database

**Warning**: Live mode may produce inconsistent snapshots if the VM is actively writing to disk during the operation.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              API Service                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   ┌─────────────────┐       ┌─────────────────┐       ┌──────────────┐  │
│   │ SandboxController│──────▶│  SandboxService │──────▶│RunnerAdapter │  │
│   │                  │       │                 │       │              │  │
│   │ POST /:id/snapshot│      │ - Find sandbox  │       │ V0: Direct   │  │
│   │                  │       │ - Check dupes   │       │     API call │  │
│   └─────────────────┘       │ - Create job    │       │              │  │
│                              └─────────────────┘       │ V2: Create   │  │
│                                                        │     job      │  │
│                                                        └──────────────┘  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
                                      │
                                      │ Job (for V2 runners)
                                      ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           Runner Service                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   ┌─────────────────┐       ┌─────────────────┐       ┌──────────────┐  │
│   │ SnapshotController│─────▶│    LibVirt     │──────▶│    MinIO     │  │
│   │                  │       │                │       │   (S3)       │  │
│   │ POST /create    │       │ CreateSnapshot │       │              │  │
│   └─────────────────┘       │ - Pause VM     │       │ PutSnapshot  │  │
│                              │ - Flatten disk │       └──────────────┘  │
│                              │ - Resume VM    │                         │
│                              │ - Upload       │                         │
│                              └─────────────────┘                        │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

## Implementation Details

### 1. Controller Layer (`sandbox.controller.ts`)

```typescript
@Post(':sandboxIdOrName/snapshot')
@ApiOperation({
  summary: 'Create a snapshot from a sandbox',
  operationId: 'createSandboxSnapshot',
})
async createSandboxSnapshot(
  @AuthContext() authContext: OrganizationAuthContext,
  @Param('sandboxIdOrName') sandboxIdOrName: string,
  @Body() dto: CreateSandboxSnapshotDto,
): Promise<SandboxDto>
```

### 2. Service Layer (`sandbox.service.ts`)

The service method performs the following steps:

1. **Find Sandbox**: Resolve sandbox by ID or name
2. **Validate State**: Ensure sandbox is in a valid state (STARTED or STOPPED)
3. **Check for Duplicates**: Query for existing active `CREATE_SANDBOX_SNAPSHOT` jobs
4. **Create Job**: Dispatch job to runner via RunnerAdapter

### 3. Duplicate Prevention

Instead of modifying the Sandbox entity state, we prevent duplicate requests by checking for active jobs:

```typescript
const existingJob = await jobRepository.findOne({
  where: {
    resourceType: ResourceType.SANDBOX,
    resourceId: sandbox.id,
    type: JobType.CREATE_SANDBOX_SNAPSHOT,
    status: In([JobStatus.PENDING, JobStatus.IN_PROGRESS]),
  },
})

if (existingJob) {
  throw new ConflictException('Snapshot creation already in progress')
}
```

### 4. Job Types

A new job type is added:

```typescript
export enum JobType {
  // ... existing types
  CREATE_SANDBOX_SNAPSHOT = 'CREATE_SANDBOX_SNAPSHOT',
}
```

### 5. Runner Adapter

#### V2 Runners (Job-based)

Creates a job that the runner polls and processes:

```typescript
async createSnapshotFromSandbox(
  sandboxId: string,
  snapshotName: string,
  live?: boolean,
): Promise<void> {
  await this.jobService.createJob(
    null,
    JobType.CREATE_SANDBOX_SNAPSHOT,
    this.runner.id,
    ResourceType.SANDBOX,
    sandboxId,
    { sandboxId, name: snapshotName, live },
  )
}
```

#### V0 Runners (Direct API)

Makes a direct API call to the runner:

```typescript
async createSnapshotFromSandbox(
  sandboxId: string,
  snapshotName: string,
  live?: boolean,
): Promise<void> {
  await this.snapshotApiClient.createSnapshot({
    sandboxId,
    name: snapshotName,
    live,
  })
}
```

### 6. Job Completion Handler

When the job completes, `JobStateHandlerService` creates the Snapshot entity:

```typescript
private async handleCreateSandboxSnapshotJobCompletion(job: Job): Promise<void> {
  if (job.status === JobStatus.COMPLETED) {
    const metadata = job.getResultMetadata()
    // Create Snapshot entity with metadata from runner
    const snapshot = new Snapshot({
      name: metadata.name,
      ref: metadata.snapshotPath,
      organizationId: sandbox.organizationId,
      // ... other fields
    })
    await this.snapshotRepository.save(snapshot)
  }
}
```

## Runner Implementation (Windows/LibVirt)

The runner's `CreateSnapshot` function in `snapshot_create.go`:

1. **Lookup Domain**: Find the VM by sandbox ID
2. **Pause (Safe Mode)**: If `live=false`, pause the VM
3. **Flatten Disk**: Convert overlay qcow2 to standalone image
   - Safe: `qemu-img convert -O qcow2 source dest`
   - Live: `qemu-img convert -U -O qcow2 source dest`
4. **Resume**: Resume VM immediately after flattening (before upload)
5. **Upload**: Stream the flattened image to S3
6. **Cleanup**: Remove temporary flattened image

### Remote Host Support

The implementation handles both local and remote libvirt hosts:

- **Local**: Direct file operations
- **Remote**: SSH-based file operations and streaming

## Error Handling

| Error | HTTP Status | Cause |
|-------|-------------|-------|
| `NotFoundException` | 404 | Sandbox not found |
| `ConflictException` | 409 | Snapshot creation already in progress |
| `BadRequestException` | 400 | Invalid sandbox state |
| `ForbiddenException` | 403 | Insufficient permissions |

## Permissions

Requires `WRITE_SANDBOXES` permission on the organization.

## Example Usage

### Create a snapshot (safe mode)

```bash
curl -X POST "https://api.daytona.io/sandbox/my-sandbox/snapshot" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Organization-Id: org-123" \
  -d '{"name": "my-snapshot-v1"}'
```

### Create a snapshot (live mode)

```bash
curl -X POST "https://api.daytona.io/sandbox/my-sandbox/snapshot" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Organization-Id: org-123" \
  -d '{"name": "my-snapshot-v1", "live": true}'
```

## Related Files

- `apps/api/src/sandbox/controllers/sandbox.controller.ts` - API endpoint
- `apps/api/src/sandbox/services/sandbox.service.ts` - Business logic
- `apps/api/src/sandbox/dto/create-sandbox-snapshot.dto.ts` - Request DTO
- `apps/api/src/sandbox/enums/job-type.enum.ts` - Job type enum
- `apps/api/src/sandbox/runner-adapter/runnerAdapter.ts` - Runner adapter interface
- `apps/api/src/sandbox/services/job-state-handler.service.ts` - Job completion handling
- `apps/runner-win/pkg/libvirt/snapshot_create.go` - Runner implementation
- `apps/runner-win/docs/SNAPSHOTS.md` - Runner-side documentation
