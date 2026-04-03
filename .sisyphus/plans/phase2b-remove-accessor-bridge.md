# Phase 2b: Remove Accessor Bridge from Sandbox Entity

## What and Why

The Sandbox entity currently has 14 getter/setter pairs (lines 140-306 of `sandbox.entity.ts`) that proxy property access to the `sandboxState` and `sandboxBackup` OneToOne relations. These were added as a transitional bridge so existing callers could keep writing `sandbox.state` instead of `sandbox.sandboxState.state`.

This bridge should be removed because:

1. It hides the data source — callers don't know reads/writes hit a different table
2. If the relation isn't loaded, getters silently return `undefined` cast to the return type (e.g., `undefined as SandboxState`)
3. Setters auto-create empty entities when the relation is null, masking load failures
4. It defeats the purpose of the table decomposition — the entity looks like it still has all fields

## Goal

Delete all 14 getter/setter pairs from `sandbox.entity.ts` and update every call site to access the relation directly:

- `sandbox.state` → `sandbox.sandboxState.state`
- `sandbox.backupState` → `sandbox.sandboxBackup.backupState`

After this change, the Sandbox entity is a clean config-only TypeORM entity with two `!`-asserted relations. Callers that need state/backup data explicitly access the relation.

## Current State (baseline: 0 tsc errors)

### Accessor bridge (to remove)

File: `apps/api/src/sandbox/entities/sandbox.entity.ts`, lines 140-306

| Getter/Setter | Proxies to |
|---|---|
| `state` | `sandboxState.state` |
| `desiredState` | `sandboxState.desiredState` |
| `pending` | `sandboxState.pending` |
| `errorReason` | `sandboxState.errorReason` |
| `recoverable` | `sandboxState.recoverable` |
| `runnerId` | `sandboxState.runnerId` |
| `prevRunnerId` | `sandboxState.prevRunnerId` |
| `daemonVersion` | `sandboxState.daemonVersion` |
| `backupState` | `sandboxBackup.backupState` |
| `backupSnapshot` | `sandboxBackup.backupSnapshot` |
| `backupRegistryId` | `sandboxBackup.backupRegistryId` |
| `lastBackupAt` | `sandboxBackup.lastBackupAt` |
| `backupErrorReason` | `sandboxBackup.backupErrorReason` |
| `existingBackupSnapshots` | `sandboxBackup.existingBackupSnapshots` |

Also remove the enum imports that only the accessors use:

- `SandboxState` (from `../enums/sandbox-state.enum`)
- `SandboxDesiredState` (from `../enums/sandbox-desired-state.enum`)
- `BackupState` (from `../enums/backup-state.enum`)

### Repository (already correct — no changes needed)

- `sandbox.repository.ts` already loads relations when returning entities
- `insert()` already accepts `stateData` and `backupData` params and attaches them to `sandbox.sandboxState` / `sandbox.sandboxBackup`
- `update()` / `updateWhere()` / `updateState()` already accept `Partial<SandboxAggregate>` and partition writes
- The repository is the access boundary — it guarantees relations are loaded

### DTO (already migrated — no changes needed)

- `sandbox.dto.ts` `fromSandbox()` already reads `sandbox.sandboxState?.desiredState`, `sandbox.sandboxBackup?.backupState` etc.
- `getSandboxState()` already reads `sandbox.sandboxState?.state`

## Call Sites to Migrate

199 references across 15 files in `apps/api/src/sandbox/` plus 3 files outside.

### By file (descending by count)

| File | Count | Error types |
|---|---|---|
| `services/sandbox.service.ts` | 53 | READ (`sandbox.state`, `.runnerId`, `.pending`, `.backupState`), WRITE (`Partial<Sandbox>` with state fields), WHERE (`FindOptionsWhere<Sandbox>` with state fields) |
| `managers/backup.manager.ts` | 34 | READ (`.backupState`, `.runnerId`, `.state`, `.backupSnapshot`, `.backupRegistryId`), QB WHERE (`sandbox.state` in SQL strings — see note below), TypeORM `find()` WHERE with state fields |
| `managers/sandbox-actions/sandbox-start.action.ts` | 31 | READ (`.state`, `.runnerId`, `.prevRunnerId`, `.backupState`, `.backupSnapshot`) |
| `managers/sandbox.manager.ts` | 26 | READ (`.state`, `.pending`, `.desiredState`, `.backupSnapshot`, `.backupRegistryId`), WRITE (`Partial<Sandbox>` with state fields), WHERE, QB WHERE |
| `services/job-state-handler.service.ts` | 18 | READ (`.state`), WRITE (`Partial<Sandbox>` with state fields) |
| `managers/snapshot.manager.ts` | 10 | READ (`.backupSnapshot`), QB SQL references |
| `controllers/sandbox.controller.ts` | 7 | READ (`.state`, `.runnerId`, `.errorReason`) |
| `runner-adapter/runnerAdapter.v2.ts` | 6 | READ (`.state`, `.runnerId`, `.prevRunnerId`) |
| `services/runner.service.ts` | 5 | READ (`.runnerId`) |
| `services/toolbox.deprecated.service.ts` | 2 | READ (`.sandboxState` — pre-existing pattern) |
| `services/sandbox-warm-pool.service.ts` | 2 | READ (`.state`) |
| `runner-adapter/runnerAdapter.v0.ts` | 2 | READ (`.state`) |
| `services/snapshot.service.ts` | 1 | READ (`.backupSnapshot`) |
| `services/volume.service.ts` | 1 | (false positive — volume.state, not sandbox) |
| `organization/services/organization-usage.service.ts` | 7 | READ (`.state`), QB SQL (`sandbox.state` in raw SQL — see note) |
| `usage/services/usage.service.ts` | 3 | READ (`.state`) |
| `webhook/dto/webhook-event-payloads.dto.ts` | 1 | READ (`event.sandbox.state`) |

## Transformation Patterns

### Pattern 1 — Direct property READ (most common, ~130 sites)

```typescript
// Before
sandbox.state
sandbox.runnerId
sandbox.backupState

// After
sandbox.sandboxState.state
sandbox.sandboxState.runnerId
sandbox.sandboxBackup.backupState
```

Variable isn't always `sandbox` — check context for: `s`, `existingSandbox`, `updatedSandbox`, `sandboxToUpdate`, `excludedSandbox`, `event.sandbox`.

### Pattern 2 — Update data objects (~30 sites)

```typescript
// Before
const updateData: Partial<Sandbox> = { state: SandboxState.STARTED, runnerId: '...' }

// After — remove type annotation (or use Partial<SandboxAggregate>)
const updateData: Partial<SandboxAggregate> = { state: SandboxState.STARTED, runnerId: '...' }
// Import: import { SandboxAggregate } from '../types/sandbox-aggregate.type'
```

The repository already accepts `Partial<SandboxAggregate>` for `update()` and `updateWhere()`, so inline objects passed directly work without annotation.

### Pattern 3 — WHERE conditions in updateWhere calls (~15 sites)

```typescript
// Before
whereCondition: { pending: false, state: sandbox.state }

// After
whereCondition: { pending: false, state: sandbox.sandboxState.state }
```

The repository's `updateWhere()` accepts `Partial<SandboxAggregate>` for whereCondition — the partitioning routes state fields to the sandbox_state table's WHERE clause.

### Pattern 4 — TypeORM `find()` / `findOne()` WHERE with state fields (~10 sites)

```typescript
// Before
this.sandboxRepository.find({
  where: { state: In([SandboxState.STARTED]), runnerId: Not(IsNull()) }
})

// After — use createAggregateQueryBuilder
this.sandboxRepository.createAggregateQueryBuilder('sandbox')
  .where('ss."state" IN (:...states)', { states: [SandboxState.STARTED] })
  .andWhere('ss."runnerId" IS NOT NULL')
  .getMany()
```

`createAggregateQueryBuilder()` returns a SelectQueryBuilder with `sandbox_state` joined as alias `ss` and `sandbox_backup` joined as alias `sb`. Callers reference state columns as `ss."columnName"` and backup columns as `sb."columnName"`.

After `.getMany()`, the returned Sandbox entities do NOT have `sandboxState`/`sandboxBackup` loaded (they're raw JOINs not relation loads). If the caller needs the relations, add `.leftJoinAndSelect('sandbox.sandboxState', 'sandboxState')` instead of the raw innerJoin.

### Pattern 5 — QueryBuilder SQL strings (~15 sites)

```typescript
// Before (in sandbox.manager.ts, backup.manager.ts, organization-usage.service.ts)
.where('sandbox.state = :state', { state: SandboxState.STARTED })
.andWhere('sandbox."runnerId" IS NOT NULL')

// After
.where('ss."state" = :state', { state: SandboxState.STARTED })
.andWhere('ss."runnerId" IS NOT NULL')
```

These are raw SQL strings in `.where()` / `.andWhere()` calls on query builders. The table alias changes from `sandbox` to `ss` (for state fields) or `sb` (for backup fields). Config fields stay as `sandbox."columnName"`.

IMPORTANT: The query builder must be `createAggregateQueryBuilder()` (which JOINs the tables), not `createQueryBuilder()` (which only queries the sandbox table). If the caller already uses `createAggregateQueryBuilder()`, just change the alias. If it uses `createQueryBuilder()`, switch to `createAggregateQueryBuilder()`.

### Pattern 6 — insert() call sites (~2 sites)

```typescript
// Before (in sandbox.service.ts)
sandbox.state = SandboxState.CREATING
sandbox.desiredState = SandboxDesiredState.STARTED
sandbox.pending = true
await this.sandboxRepository.insert(sandbox)

// After
await this.sandboxRepository.insert(sandbox, {
  state: SandboxState.CREATING,
  desiredState: SandboxDesiredState.STARTED,
  pending: true,
})
```

Remove all direct assignments of state/backup fields on the sandbox object before insert. Pass them as the second/third params to `insert()`.

## Execution Order

1. Delete the 14 getter/setter pairs from `sandbox.entity.ts` (lines 140-306)
2. Delete the now-unused enum imports (`SandboxState`, `SandboxDesiredState`, `BackupState`)
3. Run `npx tsc --noEmit --project apps/api/tsconfig.app.json` — expect ~300 errors
4. Fix files in order (largest first): sandbox.service.ts, backup.manager.ts, sandbox-start.action.ts, sandbox.manager.ts, job-state-handler.service.ts, then the rest
5. After each file: `lsp_diagnostics` to verify 0 errors
6. After all files: `npx tsc --noEmit` to verify 0 total errors

## Gotchas

1. **Variable naming**: The Sandbox variable isn't always `sandbox`. Look for: `s`, `existingSandbox`, `updatedSandbox`, `sandboxToUpdate`, `excludedSandbox`, `event.sandbox`, `result.sandbox`.

2. **Raw SQL strings**: QueryBuilder `.where()` uses SQL strings like `'sandbox.state = :state'`. These need the alias changed to `ss` for state fields, `sb` for backup fields. These are NOT caught by TypeScript — they're runtime SQL. Grep for `sandbox.state`, `sandbox."desiredState"`, `sandbox."runnerId"`, `sandbox."backupState"` etc. in string literals.

3. **Relation loading**: After migration, if a caller gets a Sandbox without the relation loaded (e.g., from `entityManager.findOne(Sandbox, ...)` without `relations: ['sandboxState']`), accessing `sandbox.sandboxState.state` throws at runtime. The repository loads relations, but code inside transaction callbacks that does its own entity fetches needs `relations: ['sandboxState', 'sandboxBackup']`.

4. **Event payloads**: Event classes like `SandboxStateUpdatedEvent` carry a `Sandbox` object. Listeners access `event.sandbox.state` — needs `event.sandbox.sandboxState.state`. Check: `webhook-event-payloads.dto.ts` line 78.

5. **organization-usage.service.ts raw SQL**: Lines 617-619 have raw SQL strings referencing `sandbox.state` and `sandbox."desiredState"`. After post-deploy migration drops these columns, this SQL breaks at runtime. Must change to JOIN sandbox_state and reference `ss."state"`.

6. **Partial<Sandbox> type annotations**: Some callers explicitly annotate `const updateData: Partial<Sandbox> = { ... }` with state fields. TypeScript will error because `state` isn't on `Sandbox`. Either remove the annotation (let inference handle it) or change to `Partial<SandboxAggregate>` and add the import.
