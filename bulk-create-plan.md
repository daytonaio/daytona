# Bulk Create Sandbox Endpoint

Add a POST /sandbox/bulk endpoint with bulkCreate service method, expose it through all 4 SDKs (TypeScript, Python, Go, Ruby), and add examples for each language demonstrating both awaitStarted modes.

## Todos

- [ ] Create BulkCreateItemDto and BulkCreateSandboxDto in apps/api/src/sandbox/dto/bulk-create-sandbox.dto.ts
- [ ] Add bulkCreateFromSnapshot() method to SandboxService
- [ ] Add POST /sandbox/bulk endpoint to SandboxController with awaitStarted wait logic
- [ ] Regenerate all 4 API clients (TS, Python, Go, Ruby) from updated OpenAPI spec
- [ ] Add BulkCreateItem type and bulkCreate() method to TypeScript SDK
- [ ] Add BulkCreateItem and bulk_create() to Python SDK (sync + async)
- [ ] Add BulkCreateItem type, options, and BulkCreate() method to Go SDK
- [ ] Add BulkCreateItem class and bulk_create method to Ruby SDK
- [ ] Create TypeScript bulk-create example
- [ ] Create Python bulk-create example (sync + async)
- [ ] Create Go bulk-create example
- [ ] Create Ruby bulk-create example

---

## 1. Backend DTOs

Create a new file [`apps/api/src/sandbox/dto/bulk-create-sandbox.dto.ts`](apps/api/src/sandbox/dto/bulk-create-sandbox.dto.ts) with two DTOs:

**`BulkCreateItemDto`** (OpenAPI name: `BulkCreateItem`):

- `snapshotIdOrName: string` (required) -- the snapshot to create sandboxes from
- `count: number` (required, integer, min 1) -- how many sandboxes to create
- `sandboxNamePrefix: string` (required) -- prefix for sandbox names
- `allowUnder?: boolean` (default false) -- if true, create as many as possible when quota/runner limits are hit instead of failing

**`BulkCreateSandboxDto`** (OpenAPI name: `BulkCreateSandbox`):

- `items: BulkCreateItemDto[]` (required, array) -- array of bulk create items
- `awaitStarted?: boolean` (default false) -- if true, the endpoint blocks until all sandboxes reach STARTED state

## 2. Backend Service Method

Add `bulkCreateFromSnapshot()` to [`apps/api/src/sandbox/services/sandbox.service.ts`](apps/api/src/sandbox/services/sandbox.service.ts):

```typescript
async bulkCreateFromSnapshot(
  dto: BulkCreateSandboxDto,
  organization: Organization,
): Promise<SandboxDto[]>
```

For each item in `dto.items`:

1. Resolve the snapshot once (reuse existing lookup logic from `createFromSnapshot` lines 382-426)
2. Validate the snapshot is ACTIVE and available in the target region
3. Calculate total resources needed: `count * (cpu, mem, disk)` from the snapshot
4. Call `validateOrganizationQuotas` -- if `allowUnder` is true, catch quota failures and reduce count to fit; otherwise fail
5. Select runner(s) via `runnerService.getRandomAvailableRunner` -- if `allowUnder` is true and no runner available, reduce count accordingly
6. Generate sandbox names: `{prefix}-00001` through `{prefix}-{count}` (zero-padded, minimum 5 digits, expanding if count > 99999)
7. Create `Sandbox` entities in a loop, setting all fields as in existing `createFromSnapshot` (lines 497-541)
8. Bulk insert using `sandboxRepository.insert(sandboxes)` (TypeORM supports array insert)
9. Emit `SandboxEvents.CREATED` for each sandbox to trigger state syncing

Return all created `SandboxDto[]`.

## 3. Backend Controller Endpoint

Add to [`apps/api/src/sandbox/controllers/sandbox.controller.ts`](apps/api/src/sandbox/controllers/sandbox.controller.ts), placed before the `/:sandboxIdOrName` GET route (to avoid route collision):

```typescript
@Post('bulk')
@HttpCode(200)
@ApiOperation({ summary: 'Bulk create sandboxes', operationId: 'bulkCreateSandbox' })
@ApiResponse({ status: 200, type: [SandboxDto] })
@RequiredOrganizationResourcePermissions([OrganizationResourcePermission.WRITE_SANDBOXES])
async bulkCreateSandbox(
  @AuthContext() authContext: OrganizationAuthContext,
  @Body() dto: BulkCreateSandboxDto,
): Promise<SandboxDto[]>
```

- Call `sandboxService.bulkCreateFromSnapshot(dto, organization)`
- If `awaitStarted` is false, return immediately after creation
- If `awaitStarted` is true, wait for all sandboxes to reach STARTED state using a multi-sandbox version of the existing `waitForSandboxStarted` pattern (Redis pub/sub with a timeout of 5 minutes)

Add a private helper `waitForAllSandboxesStarted(sandboxIds: string[], timeoutSeconds: number)` that:

- Registers callbacks for all IDs via `sandboxCallbacks`
- Resolves when all reach STARTED or ERROR/BUILD_FAILED
- Times out after `timeoutSeconds`, returning whatever state was reached

## 4. Regenerate API Clients

After the backend changes, regenerate all API clients:

- `nx run api-client:generate:api-client` (TypeScript)
- `nx run api-client-python:generate:api-client` (Python)
- `nx run api-client-go:generate:api-client` (Go)
- `nx run api-client-ruby:generate:api-client` (Ruby)

This auto-generates `bulkCreateSandbox()` methods with the `BulkCreateSandbox` and `BulkCreateItem` model types in all API client libraries.

## 5. TypeScript SDK

Add to [`libs/sdk-typescript/src/Daytona.ts`](libs/sdk-typescript/src/Daytona.ts):

**New types:**

```typescript
export interface BulkCreateItem {
  snapshotIdOrName: string
  count: number
  sandboxNamePrefix: string
  allowUnder?: boolean
}
```

**New method on `Daytona` class:**

```typescript
public async bulkCreate(
  items: BulkCreateItem[],
  options?: { awaitStarted?: boolean; timeout?: number },
): Promise<Sandbox[]>
```

Implementation:

- Calls `this.sandboxApi.bulkCreateSandbox({ items, awaitStarted })` with timeout
- Maps response `SandboxDto[]` into `Sandbox[]` instances (creating toolbox clients for each)
- If `awaitStarted` is false, the API returns immediately and sandboxes may not be started yet

Export the new type from [`libs/sdk-typescript/src/index.ts`](libs/sdk-typescript/src/index.ts).

## 6. Python SDK

**New type** in [`libs/sdk-python/src/daytona/common/daytona.py`](libs/sdk-python/src/daytona/common/daytona.py):

```python
class BulkCreateItem(BaseModel):
    snapshot_id_or_name: str
    count: int
    sandbox_name_prefix: str
    allow_under: bool = False
```

**New method** in sync [`libs/sdk-python/src/daytona/_sync/daytona.py`](libs/sdk-python/src/daytona/_sync/daytona.py):

```python
def bulk_create(self, items: list[BulkCreateItem], *, await_started: bool = False, timeout: float = 300) -> list[Sandbox]
```

**New method** in async [`libs/sdk-python/src/daytona/_async/daytona.py`](libs/sdk-python/src/daytona/_async/daytona.py):

```python
async def bulk_create(self, items: list[BulkCreateItem], *, await_started: bool = False, timeout: float = 300) -> list[AsyncSandbox]
```

Export from [`libs/sdk-python/src/daytona/__init__.py`](libs/sdk-python/src/daytona/__init__.py).

## 7. Go SDK

**New type** in [`libs/sdk-go/pkg/types/types.go`](libs/sdk-go/pkg/types/types.go):

```go
type BulkCreateItem struct {
    SnapshotIdOrName  string
    Count             int
    SandboxNamePrefix string
    AllowUnder        bool
}
```

**New options** in [`libs/sdk-go/pkg/options/client.go`](libs/sdk-go/pkg/options/client.go):

```go
type BulkCreateSandbox struct {
    AwaitStarted bool
    Timeout      *time.Duration
}
```

**New method** in [`libs/sdk-go/pkg/daytona/client.go`](libs/sdk-go/pkg/daytona/client.go):

```go
func (c *Client) BulkCreate(ctx context.Context, items []types.BulkCreateItem, opts ...func(*options.BulkCreateSandbox)) ([]*Sandbox, error)
```

## 8. Ruby SDK

**New type** in [`libs/sdk-ruby/lib/daytona/common/daytona.rb`](libs/sdk-ruby/lib/daytona/common/daytona.rb):

```ruby
class BulkCreateItem
  attr_accessor :snapshot_id_or_name, :count, :sandbox_name_prefix, :allow_under
end
```

**New method** in [`libs/sdk-ruby/lib/daytona/daytona.rb`](libs/sdk-ruby/lib/daytona/daytona.rb):

```ruby
def bulk_create(items, await_started: false, timeout: 300)
```

## 9. Examples

Create one example per language in a new `bulk-create/` directory:

### TypeScript: [`examples/typescript/bulk-create/index.ts`](examples/typescript/bulk-create/index.ts)

```typescript
import { Daytona, SandboxState } from '@daytonaio/sdk'

async function main() {
  const daytona = new Daytona()

  // Mode 1: Fire and forget (awaitStarted = false), then poll
  console.log('Bulk creating sandboxes (fire and forget)...')
  const sandboxes = await daytona.bulkCreate(
    [{ snapshotIdOrName: 'default', count: 3, sandboxNamePrefix: 'batch-a' }],
    { awaitStarted: false },
  )
  // Poll until all are started
  for (const sandbox of sandboxes) {
    while (sandbox.state !== 'started') {
      const updated = await daytona.get(sandbox.id)
      if (updated.state === 'started') break
      await new Promise((r) => setTimeout(r, 2000))
    }
    const result = await sandbox.process.executeCommand('echo hello')
    console.log(`${sandbox.name}: ${result.result}`)
  }

  // Mode 2: Await started
  console.log('Bulk creating sandboxes (await started)...')
  const readySandboxes = await daytona.bulkCreate(
    [{ snapshotIdOrName: 'default', count: 3, sandboxNamePrefix: 'batch-b' }],
    { awaitStarted: true, timeout: 300 },
  )
  for (const sandbox of readySandboxes) {
    const result = await sandbox.process.executeCommand('echo hello')
    console.log(`${sandbox.name}: ${result.result}`)
  }

  // Cleanup
  for (const s of [...sandboxes, ...readySandboxes]) {
    await daytona.delete(s)
  }
}
main().catch(console.error)
```

### Python: [`examples/python/bulk-create/bulk_create.py`](examples/python/bulk-create/bulk_create.py) (sync) and `_async/` variant

### Go: [`examples/go/bulk_create/main.go`](examples/go/bulk_create/main.go)

### Ruby: [`examples/ruby/bulk-create/bulk_create.rb`](examples/ruby/bulk-create/bulk_create.rb)

Each example demonstrates both modes: fire-and-forget with polling, and await-started with immediate exec.

## Key Design Decisions

- **Naming format**: `{prefix}-00001` with zero-padding of `max(5, digits_in_count)` width
- **`allowUnder` behavior**: On quota or runner exhaustion, silently reduce count instead of failing. The response includes only the sandboxes that were actually created, so the caller can check `response.length` vs requested count
- **Timeout for `awaitStarted`**: Handled server-side with a 5-minute default. The SDK also applies a client-side HTTP timeout
- **No build-info support in bulk**: Only snapshot-based creation is supported (no image builds). This keeps bulk operations fast and predictable
- **Runner distribution**: For large counts, the service should distribute sandboxes across available runners (round-robin or random assignment per sandbox)
