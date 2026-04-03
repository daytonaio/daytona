# Phase 2: Decouple Sandbox State + Backup — Implementation Specification

## Overview

Split the single wide `sandbox` table (~35 columns, 14+ indexes, ~2KB/row) into 3 tables by write pattern:

- `sandbox` (config) — written once at creation, rare updates
- `sandbox_state` (hot) — updated 4-5x per lifecycle
- `sandbox_backup` — independent cadence, 2-3 writes per lifecycle

**Expected outcome**: ~70% WAL reduction on state transitions (from ~4KB to ~400B per UPDATE).

## Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Scope | Both `sandbox_state` + `sandbox_backup` in one phase | Single migration cycle, shared blast radius |
| Domain model | **Hybrid facade** — separate TypeORM entities, `SandboxAggregate` type, repository facade | Clean DB + manageable migration. Best balance. |
| Migration | 3-step simplified | Create + backfill → switch code → drop old columns |
| `runnerId` placement | `sandbox_state` | Per doc. Queries needing it JOIN sandbox_state. |
| Cross-table invariants | Repository-only | Each entity enforces its own. Repository handles cross-table (e.g., DESTROYED → backupState=NONE). |
| Validation | Per-entity | `SandboxStateEntity.enforceInvariants()` and `SandboxBackupEntity.enforceInvariants()` |
| Mixed updates | Lock state first | `updateWhere()` always acquires pessimistic lock on `sandbox_state`, then applies config/backup updates in same transaction. |

---

## 1. New Entities

### 1.1 `SandboxStateEntity` (`sandbox_state` table)

```typescript
// apps/api/src/sandbox/entities/sandbox-state.entity.ts

@Entity('sandbox_state')
@Index('ss_state_idx', ['state'])
@Index('ss_desiredstate_idx', ['desiredState'])
@Index('ss_runnerid_idx', ['runnerId'])
@Index('ss_runner_state_idx', ['runnerId', 'state'])
@Index('ss_runner_state_desired_idx', ['runnerId', 'state', 'desiredState'], {
  where: '"pending" = false',
})
@Index('ss_active_only_idx', ['sandboxId'], {
  where: `"state" <> ALL (ARRAY['destroyed'::sandbox_state_enum, 'archived'::sandbox_state_enum])`,
})
@Index('ss_pending_idx', ['sandboxId'], {
  where: `"pending" = true`,
})
export class SandboxStateEntity {
  @PrimaryColumn()
  sandboxId: string

  @Column({ type: 'enum', enum: SandboxState, default: SandboxState.UNKNOWN })
  state: SandboxState

  @Column({ type: 'enum', enum: SandboxDesiredState, default: SandboxDesiredState.STARTED })
  desiredState: SandboxDesiredState

  @Column({ default: false, type: 'boolean' })
  pending: boolean

  @Column({ nullable: true })
  errorReason?: string

  @Column({ default: false, type: 'boolean' })
  recoverable: boolean

  @Column({ type: 'uuid', nullable: true })
  runnerId?: string

  @Column({ type: 'uuid', nullable: true })
  prevRunnerId?: string

  @Column({ nullable: true })
  daemonVersion?: string

  @UpdateDateColumn({ type: 'timestamp with time zone' })
  updatedAt: Date

  @OneToOne(() => Sandbox, { onDelete: 'CASCADE' })
  @JoinColumn({ name: 'sandboxId' })
  sandbox?: Sandbox

  /**
   * Per-table invariant enforcement. Only handles fields within sandbox_state.
   * Cross-table invariants (e.g., DESTROYED → backupState=NONE) handled by repository.
   */
  enforceInvariants(): Partial<SandboxStateEntity> {
    const changes: Partial<SandboxStateEntity> = {}

    if (!this.pending && String(this.state) !== String(this.desiredState)) {
      changes.pending = true
    }
    if (this.pending && String(this.state) === String(this.desiredState)) {
      changes.pending = false
    }
    if (
      this.state === SandboxState.ERROR ||
      this.state === SandboxState.BUILD_FAILED ||
      this.desiredState === SandboxDesiredState.ARCHIVED
    ) {
      changes.pending = false
    }
    if (this.state === SandboxState.DESTROYED || this.state === SandboxState.ARCHIVED) {
      changes.runnerId = null
    }

    return changes
  }
}
```

**Table structure** (~100 bytes/row):

| Column | Type | Default | Nullable |
|---|---|---|---|
| `sandboxId` | uuid PK, FK→sandbox(id) ON DELETE CASCADE | — | no |
| `state` | `sandbox_state_enum` | UNKNOWN | no |
| `desiredState` | `sandbox_desired_state_enum` | STARTED | no |
| `pending` | boolean | false | no |
| `errorReason` | varchar | — | yes |
| `recoverable` | boolean | false | no |
| `runnerId` | uuid | — | yes |
| `prevRunnerId` | uuid | — | yes |
| `daemonVersion` | varchar | — | yes |
| `updatedAt` | timestamp with time zone | now() | no |

### 1.2 `SandboxBackupEntity` (`sandbox_backup` table)

```typescript
// apps/api/src/sandbox/entities/sandbox-backup.entity.ts

@Entity('sandbox_backup')
@Index('sb_backupstate_idx', ['backupState'])
export class SandboxBackupEntity {
  @PrimaryColumn()
  sandboxId: string

  @Column({ type: 'enum', enum: BackupState, default: BackupState.NONE })
  backupState: BackupState

  @Column({ nullable: true })
  backupSnapshot: string | null

  @Column({ nullable: true })
  backupRegistryId: string | null

  @Column({ nullable: true, type: 'timestamp with time zone' })
  lastBackupAt: Date | null

  @Column({ type: 'text', nullable: true })
  backupErrorReason: string | null

  @Column({ type: 'jsonb', default: [] })
  existingBackupSnapshots: Array<{ snapshotName: string; createdAt: Date }>

  @UpdateDateColumn({ type: 'timestamp with time zone' })
  updatedAt: Date

  @OneToOne(() => Sandbox, { onDelete: 'CASCADE' })
  @JoinColumn({ name: 'sandboxId' })
  sandbox?: Sandbox

  /**
   * Per-table invariant enforcement. Only handles fields within sandbox_backup.
   */
  enforceInvariants(): Partial<SandboxBackupEntity> {
    const changes: Partial<SandboxBackupEntity> = {}
    // No per-table invariants currently needed
    // Cross-table invariant (DESTROYED → backupState=NONE) handled by repository
    return changes
  }
}
```

**Table structure** (~200 bytes/row):

| Column | Type | Default | Nullable |
|---|---|---|---|
| `sandboxId` | uuid PK, FK→sandbox(id) ON DELETE CASCADE | — | no |
| `backupState` | `backup_state_enum` | NONE | no |
| `backupSnapshot` | varchar | — | yes |
| `backupRegistryId` | varchar | — | yes |
| `lastBackupAt` | timestamp with time zone | — | yes |
| `backupErrorReason` | text | — | yes |
| `existingBackupSnapshots` | jsonb | `[]` | no |
| `updatedAt` | timestamp with time zone | now() | no |

### 1.3 Modified `Sandbox` Entity (config-only)

**Remove** from `sandbox.entity.ts`:

- Fields: `state`, `desiredState`, `pending`, `errorReason`, `recoverable`, `runnerId`, `prevRunnerId`, `daemonVersion`, `backupState`, `backupSnapshot`, `backupRegistryId`, `lastBackupAt`, `backupErrorReason`, `existingBackupSnapshots`
- Indexes: `sandbox_state_idx`, `sandbox_desiredstate_idx`, `sandbox_runnerid_idx`, `sandbox_runner_state_idx`, `sandbox_runner_state_desired_idx`, `sandbox_pending_idx`, `sandbox_active_only_idx`, `sandbox_backupstate_idx`
- Static methods: `getBackupStateUpdate()`, `getSoftDeleteUpdate()` — move to repository facade or `SandboxBackupEntity`
- Instance methods: `assertValid()`, `enforceInvariants()` — replace with per-entity versions

**Keep** on `sandbox.entity.ts`:

- `id`, `organizationId`, `name`, `region`, `class`, `snapshot`, `osUser`
- `cpu`, `gpu`, `mem`, `disk`, `env`, `labels`, `volumes`
- `public`, `networkBlockAll`, `networkAllowList`, `authToken`
- `autoStopInterval`, `autoArchiveInterval`, `autoDeleteInterval`
- `createdAt`, `updatedAt`
- Relations: `buildInfo` (ManyToOne), `lastActivityAt` (OneToOne)

**Add** relations:

- `sandboxState: SandboxStateEntity` (OneToOne)
- `sandboxBackup: SandboxBackupEntity` (OneToOne)

**Keep** indexes: `Unique(organizationId, name)`, `sandbox_snapshot_idx`, `sandbox_organizationid_idx`, `sandbox_region_idx`, `sandbox_resources_idx`, `idx_sandbox_authtoken`, `sandbox_labels_gin_full_idx`, `idx_sandbox_volumes_gin`

---

## 2. SandboxAggregate Type

```typescript
// apps/api/src/sandbox/types/sandbox-aggregate.type.ts

export interface SandboxStateFields {
  state: SandboxState
  desiredState: SandboxDesiredState
  pending: boolean
  errorReason?: string
  recoverable: boolean
  runnerId?: string
  prevRunnerId?: string
  daemonVersion?: string
}

export interface SandboxBackupFields {
  backupState: BackupState
  backupSnapshot?: string | null
  backupRegistryId?: string | null
  lastBackupAt?: Date | null
  backupErrorReason?: string | null
  existingBackupSnapshots: Array<{ snapshotName: string; createdAt: Date }>
}

/**
 * Assembled domain aggregate combining data from all 3 tables.
 * Preserves the same shape as the old Sandbox entity for API compatibility.
 * NOT a TypeORM entity — assembled by the repository facade.
 */
export type SandboxAggregate = Sandbox & SandboxStateFields & SandboxBackupFields
```

This means `SandboxDto.fromSandbox(aggregate, toolboxProxyUrl)` works unchanged.

---

## 3. Repository Facade

### 3.1 Key Methods

```
SandboxRepository
├── insert(aggregate)              → transaction: INSERT into 3 tables
├── update(id, {updateData})       → partition + transaction: UPDATE 1-3 tables
├── updateWhere(id, {updateData, whereCondition})
│                                  → pessimistic lock on sandbox_state, then update all
├── updateState(id, updateData, whereCondition)
│                                  → FAST PATH: sandbox_state only, pessimistic lock
├── updateBackup(id, updateData)   → FAST PATH: sandbox_backup only
├── findOne(options)               → JOIN 3 tables → assemble SandboxAggregate
├── find(options)                  → JOIN 3 tables → assemble SandboxAggregate[]
├── findOneBy(where)               → JOIN 3 tables → assemble SandboxAggregate
├── createAggregateQueryBuilder()  → pre-JOINed SelectQueryBuilder
├── delete(criteria)               → DELETE from sandbox (CASCADE to state + backup)
└── emitUpdateEvents()             → compare old vs new state/desiredState/public/orgId
```

### 3.2 Partitioning Logic

```typescript
const STATE_KEYS = new Set([
  'state', 'desiredState', 'pending', 'errorReason', 'recoverable',
  'runnerId', 'prevRunnerId', 'daemonVersion'
])

const BACKUP_KEYS = new Set([
  'backupState', 'backupSnapshot', 'backupRegistryId',
  'lastBackupAt', 'backupErrorReason', 'existingBackupSnapshots'
])

function partition(updateData: Partial<SandboxAggregate>): {
  stateFields: Partial<SandboxStateEntity>
  backupFields: Partial<SandboxBackupEntity>
  configFields: Partial<Sandbox>
} {
  // Split updateData keys into the 3 buckets
}
```

### 3.3 Cross-Table Invariants (Repository)

Applied within the transaction after per-entity invariants:

```typescript
// In repository.update() / updateWhere() transaction:
if (newStateFields.state === SandboxState.DESTROYED) {
  await em.update(SandboxBackupEntity, { sandboxId: id }, { backupState: BackupState.NONE })
}
```

### 3.4 `updateWhere()` Locking Strategy

```
1. BEGIN TRANSACTION
2. SELECT sandbox_state FOR UPDATE WHERE sandboxId=:id AND state-level conditions
3. If not found → throw SandboxConflictError
4. Apply state updateData + enforceInvariants() on SandboxStateEntity
5. UPDATE sandbox_state
6. If backup fields in updateData → UPDATE sandbox_backup
7. If config fields in updateData → UPDATE sandbox
8. Apply cross-table invariants
9. COMMIT
10. Emit events (outside transaction)
```

### 3.5 `updateState()` Fast Path

Used by `SandboxAction.updateSandboxState()` — the state machine hot path:

```
1. BEGIN TRANSACTION
2. SELECT sandbox_state FOR UPDATE WHERE sandboxId=:id AND whereCondition
3. Apply updateData + enforceInvariants()
4. UPDATE sandbox_state
5. Apply cross-table invariants (DESTROYED → backup)
6. Upsert sandbox_last_activity if state changed
7. COMMIT
8. Emit SandboxStateUpdatedEvent / SandboxDesiredStateUpdatedEvent
```

This is the primary performance win — locks only the ~100B `sandbox_state` row instead of the ~2KB `sandbox` row.

---

## 4. Migration Plan

### Step 1: Pre-deploy Migration (Create + Backfill)

```sql
-- Create sandbox_state table
CREATE TABLE sandbox_state (
  "sandboxId" uuid NOT NULL PRIMARY KEY REFERENCES sandbox(id) ON DELETE CASCADE,
  state sandbox_state_enum NOT NULL DEFAULT 'unknown',
  "desiredState" sandbox_desired_state_enum NOT NULL DEFAULT 'started',
  pending boolean NOT NULL DEFAULT false,
  "errorReason" character varying,
  recoverable boolean NOT NULL DEFAULT false,
  "runnerId" uuid,
  "prevRunnerId" uuid,
  "daemonVersion" character varying,
  "updatedAt" timestamp with time zone NOT NULL DEFAULT now()
);

-- Create sandbox_backup table
CREATE TABLE sandbox_backup (
  "sandboxId" uuid NOT NULL PRIMARY KEY REFERENCES sandbox(id) ON DELETE CASCADE,
  "backupState" backup_state_enum NOT NULL DEFAULT 'none',
  "backupSnapshot" character varying,
  "backupRegistryId" character varying,
  "lastBackupAt" timestamp with time zone,
  "backupErrorReason" text,
  "existingBackupSnapshots" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "updatedAt" timestamp with time zone NOT NULL DEFAULT now()
);

-- Backfill in batches of 100K
INSERT INTO sandbox_state ("sandboxId", state, "desiredState", pending, "errorReason", recoverable, "runnerId", "prevRunnerId", "daemonVersion", "updatedAt")
SELECT id, state, "desiredState", pending, "errorReason", recoverable, "runnerId", "prevRunnerId", "daemonVersion", "updatedAt"
FROM sandbox
ORDER BY id
ON CONFLICT DO NOTHING;

INSERT INTO sandbox_backup ("sandboxId", "backupState", "backupSnapshot", "backupRegistryId", "lastBackupAt", "backupErrorReason", "existingBackupSnapshots", "updatedAt")
SELECT id, "backupState", "backupSnapshot", "backupRegistryId", "lastBackupAt", "backupErrorReason", "existingBackupSnapshots", "updatedAt"
FROM sandbox
ORDER BY id
ON CONFLICT DO NOTHING;

-- Create indexes on new tables
CREATE INDEX ss_state_idx ON sandbox_state (state);
CREATE INDEX ss_desiredstate_idx ON sandbox_state ("desiredState");
CREATE INDEX ss_runnerid_idx ON sandbox_state ("runnerId");
CREATE INDEX ss_runner_state_idx ON sandbox_state ("runnerId", state);
CREATE INDEX ss_runner_state_desired_idx ON sandbox_state ("runnerId", state, "desiredState") WHERE pending = false;
CREATE INDEX ss_active_only_idx ON sandbox_state ("sandboxId") WHERE state <> ALL (ARRAY['destroyed'::sandbox_state_enum, 'archived'::sandbox_state_enum]);
CREATE INDEX ss_pending_idx ON sandbox_state ("sandboxId") WHERE pending = true;
CREATE INDEX sb_backupstate_idx ON sandbox_backup ("backupState");
```

### Step 2: Code Deploy

Deploy new code that reads/writes new tables. Old columns still exist but are unused.

### Step 3: Post-deploy Migration (Drop Old Columns)

```sql
-- Drop old columns from sandbox table
ALTER TABLE sandbox DROP COLUMN IF EXISTS state;
ALTER TABLE sandbox DROP COLUMN IF EXISTS "desiredState";
ALTER TABLE sandbox DROP COLUMN IF EXISTS pending;
ALTER TABLE sandbox DROP COLUMN IF EXISTS "errorReason";
ALTER TABLE sandbox DROP COLUMN IF EXISTS recoverable;
ALTER TABLE sandbox DROP COLUMN IF EXISTS "runnerId";
ALTER TABLE sandbox DROP COLUMN IF EXISTS "prevRunnerId";
ALTER TABLE sandbox DROP COLUMN IF EXISTS "daemonVersion";
ALTER TABLE sandbox DROP COLUMN IF EXISTS "backupState";
ALTER TABLE sandbox DROP COLUMN IF EXISTS "backupSnapshot";
ALTER TABLE sandbox DROP COLUMN IF EXISTS "backupRegistryId";
ALTER TABLE sandbox DROP COLUMN IF EXISTS "lastBackupAt";
ALTER TABLE sandbox DROP COLUMN IF EXISTS "backupErrorReason";
ALTER TABLE sandbox DROP COLUMN IF EXISTS "existingBackupSnapshots";

-- Drop old indexes (these referenced columns that no longer exist)
DROP INDEX IF EXISTS sandbox_state_idx;
DROP INDEX IF EXISTS sandbox_desiredstate_idx;
DROP INDEX IF EXISTS sandbox_runnerid_idx;
DROP INDEX IF EXISTS sandbox_runner_state_idx;
DROP INDEX IF EXISTS sandbox_runner_state_desired_idx;
DROP INDEX IF EXISTS sandbox_pending_idx;
DROP INDEX IF EXISTS sandbox_active_only_idx;
DROP INDEX IF EXISTS sandbox_backupstate_idx;
```

---

## 5. File Change Map

### New Files

| File | Description |
|---|---|
| `sandbox/entities/sandbox-state.entity.ts` | SandboxStateEntity |
| `sandbox/entities/sandbox-backup.entity.ts` | SandboxBackupEntity |
| `sandbox/types/sandbox-aggregate.type.ts` | SandboxAggregate type + interfaces |
| `migrations/{ts1}-migration.ts` | Pre-deploy: create tables + backfill |
| `migrations/{ts2}-migration.ts` | Post-deploy: drop old columns |

### Modified Files

| File | Change Type | Details |
|---|---|---|
| **Entity layer** | | |
| `sandbox/entities/sandbox.entity.ts` | Major | Remove state/backup fields + indexes. Add OneToOne relations. Remove assertValid/enforceInvariants (replaced by per-entity). Remove static helpers (moved to repo). |
| **Repository layer** | | |
| `sandbox/repositories/sandbox.repository.ts` | Major rewrite | Implement facade: partition logic, multi-table transactions, fast paths, aggregate assembly, cross-table invariants. |
| **State machine** | | |
| `sandbox/managers/sandbox-actions/sandbox.action.ts` | Moderate | `updateSandboxState()` → call `repository.updateState()` fast path. |
| `sandbox/managers/sandbox-actions/sandbox-start.action.ts` | Moderate | Read `runnerId`, `backupState` etc. from aggregate. Use `updateState()`. |
| `sandbox/managers/sandbox-actions/sandbox-stop.action.ts` | Minor | Use `updateState()`. |
| `sandbox/managers/sandbox-actions/sandbox-destroy.action.ts` | Minor | Use `updateState()`. |
| `sandbox/managers/sandbox-actions/sandbox-archive.action.ts` | Minor | Use `updateState()`. Read `backupState` from aggregate. |
| **Managers** | | |
| `sandbox/managers/sandbox.manager.ts` | Moderate | Cron queries switch to `createAggregateQueryBuilder()` with JOINs. `syncStates()`, `autostopCheck()`, `autoArchiveCheck()`, `autoDeleteCheck()`, `drainingRunnerSandboxesCheck()` all need query refactor. |
| `sandbox/managers/backup.manager.ts` | Moderate | Cron queries add JOINs for state fields. `setBackupPending()` uses `updateBackup()`. `checkBackupProgress()` uses `updateBackup()`. |
| `sandbox/managers/snapshot.manager.ts` | Minor | Queries that read backup fields need JOINs. |
| **Services** | | |
| `sandbox/services/sandbox.service.ts` | Moderate | `updateSandboxBackupState()` → use `repository.updateBackup()`. `updateState()` → use `repository.updateState()`. Cleanup cron queries. |
| `sandbox/services/job-state-handler.service.ts` | Minor | Read state/backup from aggregate. |
| `sandbox/services/runner.service.ts` | Minor | If it queries sandbox state fields directly. |
| **Module** | | |
| `sandbox/sandbox.module.ts` | Minor | Register `SandboxStateEntity`, `SandboxBackupEntity` in `TypeOrmModule.forFeature()`. |
| **DTO** | | |
| `sandbox/dto/sandbox.dto.ts` | None | `fromSandbox()` receives SandboxAggregate (same shape). Unchanged. |
| **Controller** | | |
| `sandbox/controllers/sandbox.controller.ts` | None | Calls service methods that return SandboxAggregate. Unchanged. |
| **Events** | | |
| `sandbox/events/*.ts` | None | Events carry SandboxAggregate (same shape as old Sandbox). |

### Files NOT Changed

- `sandbox/dto/*.ts` (DTOs unchanged)
- `sandbox/controllers/*.ts` (Controllers call services, not repos)
- `sandbox/events/*.ts` (Events carry aggregate)
- `sandbox/enums/*.ts` (Enums unchanged)
- `sandbox/guards/*.ts` (Guards unchanged)
- `sandbox/proxy/*.ts` (Proxy unchanged)
- `sandbox/subscribers/*.ts` (May need minor if they read state)
- `common/repositories/base.repository.ts` (Unchanged — SandboxRepository may no longer extend it, or extends it for config table only)

---

## 6. Verification Criteria

### Per Phase 2 refactor doc

- [ ] `EXPLAIN ANALYZE` cron queries on split tables vs single table — no regression
- [ ] `SandboxDto` API responses unchanged (compare before/after payloads)
- [ ] No index regressions on cron job queries
- [ ] State transition UPDATE WAL: ~100B row + 7 indexes (was ~2KB + 14 indexes)
- [ ] All existing tests pass without modification (or with minimal adaptation)
- [ ] Cross-table invariants verified: DESTROYED → backupState=NONE, ARCHIVED → runnerId=null

### Build/Test

- [ ] `npm run build` passes
- [ ] `npm run test` passes
- [ ] `npm run migration:generate` produces no diff after Step 3

---

## 7. Execution Order

Suggested task breakdown for implementation:

1. **Create new entity files** (sandbox-state.entity.ts, sandbox-backup.entity.ts, sandbox-aggregate.type.ts)
2. **Write pre-deploy migration** (create tables, backfill, indexes)
3. **Rewrite SandboxRepository** (facade with partition, multi-table transactions, fast paths)
4. **Update sandbox.entity.ts** (remove state/backup fields, add relations)
5. **Update SandboxAction base class** (use updateState() fast path)
6. **Update all sandbox action classes** (start, stop, destroy, archive)
7. **Update SandboxManager cron queries** (aggregate query builder)
8. **Update BackupManager** (cron queries + updateBackup())
9. **Update SandboxService** (updateSandboxBackupState, updateState, cron jobs)
10. **Update SnapshotManager** (backup field JOINs)
11. **Update sandbox.module.ts** (register new entities)
12. **Write post-deploy migration** (drop old columns)
13. **Run lsp_diagnostics, build, tests** — verify everything passes
14. **Manual QA** — verify API responses, state transitions, cron behavior
