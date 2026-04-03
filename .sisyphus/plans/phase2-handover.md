# Phase 2 Handover Context — Sandbox State Decoupling

> **Purpose**: Everything a new session needs to implement Phase 2 without re-exploring the codebase.
> **Spec file**: `.sisyphus/plans/phase2-sandbox-state-decoupling.md` — contains entity definitions, migration SQL, file change map, execution order.

---

## 1. Codebase Architecture (apps/api/src)

**Stack**: NestJS + TypeORM + PostgreSQL + Redis + EventEmitter2
**ORM**: TypeORM with custom `BaseRepository<T>` base class (`common/repositories/base.repository.ts`)
**Migrations**: Expand-and-contract pattern. Pre-deploy (additive) and post-deploy (breaking). See `migrations/README.md`.
**Entity registration**: `autoLoadEntities: true` in `app.module.ts` — any `@Entity()` in a module's `TypeOrmModule.forFeature()` is auto-discovered.

### Module Structure

```
sandbox/
├── controllers/          # HTTP endpoints
├── services/             # Business logic (SandboxService, SandboxActivityService)
├── entities/             # TypeORM entities (Sandbox, SandboxLastActivity, Runner, etc.)
├── repositories/         # Custom repos (SandboxRepository extends BaseRepository)
├── dto/                  # Request/response DTOs
├── enums/                # SandboxState, SandboxDesiredState, BackupState, etc.
├── events/               # Event classes (SandboxStateUpdatedEvent, etc.)
├── managers/             # Orchestration (SandboxManager, BackupManager, SnapshotManager)
│   └── sandbox-actions/  # State machine actions (Start, Stop, Destroy, Archive)
├── runner-adapter/       # Runner communication layer
├── common/               # Redis lock provider
├── constants/            # Event names, warm pool org ID
├── utils/                # Lock keys, cache utils, error sanitization
└── sandbox.module.ts     # Module definition
```

---

## 2. Key Files and Their Roles

### Entity Layer

| File | Lines | Role |
|---|---|---|
| `sandbox/entities/sandbox.entity.ts` | ~403 | Main entity. ~35 columns, 14+ indexes. Has `assertValid()`, `enforceInvariants()`, `getBackupStateUpdate()`, `getSoftDeleteUpdate()` static helpers. |
| `sandbox/entities/sandbox-last-activity.entity.ts` | ~25 | **Phase 1 reference pattern**. Separate table with OneToOne to Sandbox. FK with CASCADE delete. |

### Repository Layer

| File | Lines | Role |
|---|---|---|
| `common/repositories/base.repository.ts` | 136 | Abstract base with `findOne`, `find`, `count`, `delete`, abstract `insert`/`update`. |
| `sandbox/repositories/sandbox.repository.ts` | ~279 | Extends BaseRepository. Key methods: `insert()` (lines 33-53), `update()` (lines 61-119), `updateWhere()` (lines 132-176), `upsertLastActivity()` (lines 181-187), `emitUpdateEvents()` (lines 243-278). **This is the file that gets the major rewrite.** |

### State Machine

| File | Lines | Role |
|---|---|---|
| `managers/sandbox-actions/sandbox.action.ts` | 105 | Base class. `updateSandboxState()` (lines 33-104) is THE method for all state transitions. Validates lock code, builds updateData, calls `repository.update()`. |
| `managers/sandbox-actions/sandbox-start.action.ts` | 866 | Most complex action. Runner assignment, snapshot pulling, build orchestration, restore from backup. Heavily reads `runnerId`, `backupState`, `backupSnapshot`. |
| `managers/sandbox-actions/sandbox-stop.action.ts` | ~80 | Polls runner for stop confirmation. |
| `managers/sandbox-actions/sandbox-destroy.action.ts` | ~70 | Calls runner to destroy, transitions state. |
| `managers/sandbox-actions/sandbox-archive.action.ts` | ~100 | Waits for backup completion before archiving. |

### Managers (Cron Jobs)

| File | Lines | Key Crons |
|---|---|---|
| `managers/sandbox.manager.ts` | 961 | `autostopCheck` (10s), `autoArchiveCheck` (10s), `autoDeleteCheck` (10s), `drainingRunnerSandboxesCheck` (10s), `syncStates` (10s), `syncArchivedDesiredStates` (10s), `syncArchivedCompletedStates` (10s). Also `syncInstanceState()` — the core state machine loop. |
| `managers/backup.manager.ts` | 549 | `adHocBackupCheck` (5min), `checkBackupStates` (10s), `checkBackupStatesForErroredDraining` (10s), `syncStopStateCreateBackups` (10s). Event handlers for ARCHIVED, DESTROYED, BACKUP_CREATED. |
| `managers/snapshot.manager.ts` | ~1200 | 8 cron jobs for snapshot lifecycle. Reads `backupSnapshot`, `backupRegistryId` for draining runner migration. |

### Service Layer

| File | Lines | Role |
|---|---|---|
| `services/sandbox.service.ts` | ~2200 | Main service. `createFromSnapshot()`, `start()`, `stop()`, `destroy()`, `archive()`, `updateState()`, `updateSandboxBackupState()`. DTO assembly via `toSandboxDto()` → `SandboxDto.fromSandbox()`. |
| `services/sandbox-activity.service.ts` | 179 | **Phase 1 implementation**. Redis ZADD buffering + batch flush to `sandbox_last_activity` table. Reference pattern for how the codebase handles split tables. |
| `services/job-state-handler.service.ts` | ~300 | Handles job completion from v2 runners. Reads `backupState` to determine archiving eligibility. |

### DTO / Controller

| File | Lines | Role |
|---|---|---|
| `dto/sandbox.dto.ts` | 364 | `fromSandbox()` (line 279) maps entity → API response. `getSandboxState()` (line 324) computes display state from state + desiredState combo. **Unchanged by this refactoring** (receives SandboxAggregate with same shape). |
| `controllers/sandbox.controller.ts` | ~1300 | HTTP endpoints. `waitForSandboxStarted()` (line 1271) uses in-memory Map + Redis pub/sub. Calls `service.toSandboxDto()`. **Unchanged by this refactoring.** |

---

## 3. Critical Patterns to Follow

### Pattern: SandboxLastActivity (Phase 1 Reference)

This is the exact pattern already established for splitting data out of the sandbox table:

- **Entity**: `sandbox-last-activity.entity.ts` — OneToOne with Sandbox, FK with CASCADE delete
- **Service**: `sandbox-activity.service.ts` — Redis buffer (ZADD) + batch flush cron (ZRANGEBYSCORE → bulk UPSERT)
- **Repository integration**: `sandbox.repository.ts` line 181-187 — `upsertLastActivity()` called within update transactions
- **Cron queries**: Already use `innerJoin('sandbox_last_activity', 'activity', ...)` in query builders

### Pattern: Repository Update with Event Emission

```
1. Begin transaction
2. Apply updateData to entity
3. Call entity.assertValid() / enforceInvariants()
4. Execute UPDATE
5. Upsert lastActivityAt (if state/org changed)
6. Commit
7. Emit events (SandboxStateUpdatedEvent, SandboxDesiredStateUpdatedEvent) AFTER commit
8. Invalidate lookup cache
```

Events MUST be emitted after the transaction commits, not inside it. See `emitUpdateEvents()` at line 243.

### Pattern: Optimistic Concurrency (updateWhere)

```typescript
// Acquires pessimistic_write lock, verifies WHERE condition matches
const existing = await em.findOne(Entity, {
  where: { id, ...whereCondition },
  lock: { mode: 'pessimistic_write' },
})
if (!existing) throw new SandboxConflictError()
```

The `whereCondition` is typically `{ pending: false, state: currentState }` or `{ pending: sandbox.pending, state: sandbox.state }`.

### Pattern: Redis Lock Provider

Every cron job and state transition acquires a distributed lock:

```typescript
const lockKey = getStateChangeLockKey(sandboxId) // 'sandbox-state-change:{sandboxId}'
const acquired = await this.redisLockProvider.lock(lockKey, ttlSeconds, lockCode?)
if (!acquired) return
try { /* work */ } finally { await this.redisLockProvider.unlock(lockKey) }
```

### Pattern: Cron Query Builder (current style)

```typescript
const sandboxes = await this.sandboxRepository
  .createQueryBuilder('sandbox')
  .innerJoin('sandbox_last_activity', 'activity', 'activity."sandboxId" = sandbox.id')
  .where('sandbox.state = :state', { state: SandboxState.STARTED })
  .andWhere('sandbox."desiredState" = :desiredState', { desiredState: SandboxDesiredState.STARTED })
  .andWhere('activity."lastActivityAt" < NOW() - INTERVAL \'1 minute\' * sandbox."autoStopInterval"')
  .getMany()
```

After refactoring, `sandbox.state` becomes `ss.state` and requires JOIN to `sandbox_state`.

---

## 4. Pre-existing Conditions

### Already-wired references to split properties

These files ALREADY reference `sandbox.sandboxState` and `sandbox.sandboxBackup` (properties that don't exist yet on the entity). They show PRE-EXISTING LSP errors — someone started preparing:

- `sandbox/runner-adapter/runnerAdapter.v2.ts` — lines 88, 113, 114, 132, 241, 242
- `sandbox/services/runner.service.ts` — lines 266, 270
- `sandbox/services/toolbox.deprecated.service.ts` — lines 96, 98
- `sandbox/services/snapshot.service.ts` — line 661

These files will naturally resolve once the OneToOne relations are added to the Sandbox entity.

### lastActivityAt side-effect in repository

`sandbox.repository.ts` calls `upsertLastActivity()` on:

1. `insert()` — with `createdAt` timestamp (line 47)
2. `update()` — with `updatedAt` if `state` or `organizationId` changed (line 110-111)
3. `updateWhere()` — with `updatedAt` if `state` or `organizationId` changed (line 167-168)

This side-effect MUST be preserved in the facade. The `updateState()` fast path should also call `upsertLastActivity()` when state changes.

### Event emission carries full entity

Events like `SandboxStateUpdatedEvent` carry the full `Sandbox` entity reference (which will become `SandboxAggregate`). The controller's `waitForSandboxStarted()` subscribes to these via Redis pub/sub. The event shape doesn't need to change since `SandboxAggregate` has the same fields.

### getSoftDeleteUpdate() static method

`Sandbox.getSoftDeleteUpdate(sandbox)` (line 278-285) builds the update payload for soft deletes:

```typescript
{ pending: true, desiredState: DESTROYED, backupState: NONE, name: `DESTROYED_${id}_${name}` }
```

This spans state + backup + config fields. Must be moved to the repository facade's partition logic.

### enforceInvariants() cross-table logic

Current `enforceInvariants()` (line 370-402) does:

- `!pending && state !== desiredState` → `pending = true` (state-only)
- `pending && state === desiredState` → `pending = false` (state-only)
- ERROR/BUILD_FAILED or desiredState=ARCHIVED → `pending = false` (state-only)
- DESTROYED/ARCHIVED → `runnerId = null` (state-only)
- **DESTROYED → `backupState = NONE`** (CROSS-TABLE — repository handles this)

---

## 5. Gotchas and Edge Cases

### 1. SandboxAction.updateSandboxState() writes backupState

At line 91-96 of `sandbox.action.ts`:

```typescript
if (state == SandboxState.DESTROYED) {
  updateData.backupState = BackupState.NONE
}
if (backupState !== undefined) {
  Object.assign(updateData, Sandbox.getBackupStateUpdate(sandbox, backupState))
}
```

After the split, this method should use the `updateState()` fast path for state fields, and the repository handles the cross-table cascade (DESTROYED → backup).

### 2. Sandbox.getBackupStateUpdate() is complex

This static method (lines 233-273) manages backup state transitions with side effects:

- COMPLETED → appends to `existingBackupSnapshots` array, sets `lastBackupAt`
- ERROR → sets `backupErrorReason`
- PENDING → sets `backupSnapshot`, `backupRegistryId`
- NONE → clears `backupSnapshot`
This logic should move to `SandboxBackupEntity` or a helper on the repository.

### 3. BackupManager queries span state + backup fields

`checkBackupStates()` (backup.manager.ts:148-179) queries:

```sql
WHERE sandbox.state IN ('archiving', 'started', 'stopped')
  AND sandbox.backupState IN ('pending', 'in_progress')
```

After split, this needs JOINs to both `sandbox_state` (for `state`) and `sandbox_backup` (for `backupState`).

### 4. drainingRunnerSandboxesCheck() queries across all 3 domains

It queries `state`, `desiredState`, `backupState`, `backupSnapshot`, `runnerId`, `recoverable` — spanning state + backup fields. Needs JOINs to both new tables.

### 5. SandboxStartAction.checkTimeoutError() reads from SandboxActivityService

Line 654-666: Uses `sandboxActivityService.getLastActivityAt()` (Redis first, then DB) to check if sandbox has timed out. This is separate from the table split but shows the activity pattern.

### 6. TypeORM `find()` with WHERE on split fields

Callers like `this.sandboxRepository.find({ where: { state: SandboxState.STOPPED, backupState: BackupState.COMPLETED } })` won't work with TypeORM's native `find()` after the split because `state` and `backupState` aren't columns on the `sandbox` table anymore. The facade must intercept these and translate to query builder JOINs.

### 7. Module registration

New entities (`SandboxStateEntity`, `SandboxBackupEntity`) MUST be added to `TypeOrmModule.forFeature()` in `sandbox.module.ts` (line ~73) for TypeORM to recognize them.

### 8. Migration batch size

For the backfill in Step 1, consider batching the INSERT...SELECT if the sandbox table is large. The spec shows a single INSERT but production may need `LIMIT`/`OFFSET` batching or `INSERT...SELECT...ORDER BY id` with cursor pagination.

---

## 6. Implementation Guidance

### Recommended approach: Delegate to `deep` category

This is a large, interconnected refactoring. Recommended delegation:

```
task(category="deep", load_skills=[], prompt="
Read .sisyphus/plans/phase2-sandbox-state-decoupling.md and 
.sisyphus/plans/phase2-handover.md for full context.

Implement Phase 2 of the sandbox model refactoring following the 
14-step execution order in the spec. Key principles:
- Hybrid facade pattern: separate entities, unified SandboxAggregate
- Repository facade routes writes to correct tables
- Per-entity enforceInvariants(), cross-table invariants in repository
- updateState() fast path for state machine (sandbox_state only)
- updateBackup() fast path for backup manager (sandbox_backup only)
- 3-step migration: create+backfill → switch code → drop old columns

Start with steps 1-4 (entities, migration, repository rewrite, sandbox entity modification).
Build and run diagnostics after each major step.
")
```

### Alternative: Step-by-step in conversation

If implementing directly:

1. Start with entity files (low risk, foundation)
2. Write pre-deploy migration (independent, testable)
3. Repository rewrite (highest complexity — do this with full attention)
4. Update sandbox.entity.ts (depends on repository being ready)
5. Update action classes (depend on repository)
6. Update managers/services (depend on repository + actions)
7. Post-deploy migration (last, after everything works)

### Testing strategy

1. After entity creation: `npm run build` should pass (new files, no references yet)
2. After repository rewrite: `lsp_diagnostics` on sandbox.repository.ts
3. After entity modification: `npm run build` — expect errors in ~15 files that reference removed fields
4. Fix all dependent files
5. `npm run build` — should pass
6. `npm run test` — should pass
7. Manual QA: verify API responses match before/after

---

## 7. Quick Reference: Column Ownership

### sandbox_state columns (move FROM sandbox)

`state`, `desiredState`, `pending`, `errorReason`, `recoverable`, `runnerId`, `prevRunnerId`, `daemonVersion`

### sandbox_backup columns (move FROM sandbox)

`backupState`, `backupSnapshot`, `backupRegistryId`, `lastBackupAt`, `backupErrorReason`, `existingBackupSnapshots`

### sandbox columns (KEEP)

`id`, `organizationId`, `name`, `region`, `class`, `snapshot`, `osUser`, `cpu`, `gpu`, `mem`, `disk`, `env`, `labels`, `volumes`, `public`, `networkBlockAll`, `networkAllowList`, `authToken`, `autoStopInterval`, `autoArchiveInterval`, `autoDeleteInterval`, `createdAt`, `updatedAt`, `buildInfoSnapshotRef`

### Indexes that MOVE to sandbox_state

`sandbox_state_idx`, `sandbox_desiredstate_idx`, `sandbox_runnerid_idx`, `sandbox_runner_state_idx`, `sandbox_runner_state_desired_idx`, `sandbox_pending_idx`, `sandbox_active_only_idx`

### Indexes that MOVE to sandbox_backup

`sandbox_backupstate_idx`

### Indexes that STAY on sandbox

`Unique(organizationId, name)`, `sandbox_snapshot_idx`, `sandbox_organizationid_idx`, `sandbox_region_idx`, `sandbox_resources_idx`, `idx_sandbox_authtoken`, `sandbox_labels_gin_full_idx`, `idx_sandbox_volumes_gin`
