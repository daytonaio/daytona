# Phase 2c: Fix Raw SQL Query Builder Aliases

## What and Why

Phase 2b removed the 14 getter/setter accessor pairs from `Sandbox` and migrated all TypeScript call sites. The project now compiles with 0 tsc errors.

However, 21 raw SQL strings inside `.where()` / `.andWhere()` / `.addSelect()` / `.addOrderBy()` / `.orderBy()` / `.select()` calls still reference `sandbox.state`, `sandbox."desiredState"`, `sandbox."runnerId"`, `sandbox.backupState`, `sandbox.lastBackupAt`, etc. TypeScript cannot catch these — they are opaque string literals that resolve at **runtime** against PostgreSQL.

These queries currently work because the physical columns still exist on the `sandbox` table (the post-deploy migration has not yet dropped them). Once the migration drops the columns, every one of these queries will fail with a PostgreSQL runtime error.

## Goal

1. Switch each query from `createQueryBuilder('sandbox')` to `createAggregateQueryBuilder('sandbox')`
2. Change the SQL alias for state fields from `sandbox` to `ss`
3. Change the SQL alias for backup fields from `sandbox` to `sb`
4. Keep config field aliases as `sandbox`
5. Fix the `innerJoin('runner', 'r', 'r.id = sandbox.runnerId')` pattern — after the switch, `runnerId` lives in `sandbox_state`, so the join condition becomes `r.id = ss."runnerId"`

## Reference: `createAggregateQueryBuilder`

```typescript
// apps/api/src/sandbox/repositories/sandbox.repository.ts line 454
createAggregateQueryBuilder(alias = 'sandbox'): SelectQueryBuilder<Sandbox> {
  return this.repository
    .createQueryBuilder(alias)
    .innerJoin('sandbox_state', 'ss', `ss."sandboxId" = ${alias}."id"`)
    .innerJoin('sandbox_backup', 'sb', `sb."sandboxId" = ${alias}."id"`)
}
```

Aliases:

- `sandbox` — the `sandbox` config table (id, organizationId, name, region, class, snapshot, osUser, cpu, mem, disk, gpu, env, public, volumes, createdAt, updatedAt, autoStopInterval, autoArchiveInterval, autoDeleteInterval, authToken, networkBlockAll, networkAllowList, labels)
- `ss` — the `sandbox_state` table (sandboxId, state, desiredState, pending, errorReason, recoverable, runnerId, prevRunnerId, daemonVersion, updatedAt)
- `sb` — the `sandbox_backup` table (sandboxId, backupState, backupSnapshot, backupRegistryId, lastBackupAt, backupErrorReason, existingBackupSnapshots)

## Exact Sites to Fix (21 total, 5 files)

### File 1: `apps/api/src/sandbox/managers/backup.manager.ts` (7 sites, 3 query builders)

**QB 1 — `checkBackupStates` (line 150)**

```
Current:  .createQueryBuilder('sandbox')
          .innerJoin('runner', 'r', 'r.id = sandbox.runnerId')
          .where('sandbox.state IN (:...states)', ...)
          .andWhere('sandbox.backupState IN (:...backupStates)', ...)
          .addSelect(`CASE sandbox.state WHEN ... END`, 'state_priority')
          .addOrderBy('sandbox.lastBackupAt', 'ASC', 'NULLS FIRST')

Change to: .createAggregateQueryBuilder('sandbox')
           .innerJoin('runner', 'r', 'r.id = ss."runnerId"')
           .where('ss."state" IN (:...states)', ...)
           .andWhere('sb."backupState" IN (:...backupStates)', ...)
           .addSelect(`CASE ss."state" WHEN ... END`, 'state_priority')
           .addOrderBy('sb."lastBackupAt"', 'ASC', 'NULLS FIRST')
```

Note: The `sandbox.createdAt` in `.addOrderBy('sandbox.createdAt', 'ASC')` stays as-is (config field).

**QB 2 — `checkBackupStatesForErroredDraining` (line 253)**

```
Current:  .createQueryBuilder('sandbox')
          .innerJoin('runner', 'r', 'r.id = sandbox.runnerId')
          .where('sandbox.state = :error', ...)
          .andWhere('sandbox.backupState IN (:...backupStates)', ...)
          .addOrderBy('sandbox.lastBackupAt', 'ASC', 'NULLS FIRST')

Change to: .createAggregateQueryBuilder('sandbox')
           .innerJoin('runner', 'r', 'r.id = ss."runnerId"')
           .where('ss."state" = :error', ...)
           .andWhere('sb."backupState" IN (:...backupStates)', ...)
           .addOrderBy('sb."lastBackupAt"', 'ASC', 'NULLS FIRST')
```

**QB 3 — `syncStopStateCreateBackups` (line 338)**

```
Current:  .createQueryBuilder('sandbox')
          .innerJoin('runner', 'r', 'r.id = sandbox.runnerId')
          .where('sandbox.state IN (:...states)', ...)
          .andWhere('sandbox.backupState = :none', ...)

Change to: .createAggregateQueryBuilder('sandbox')
           .innerJoin('runner', 'r', 'r.id = ss."runnerId"')
           .where('ss."state" IN (:...states)', ...)
           .andWhere('sb."backupState" = :none', ...)
```

### File 2: `apps/api/src/sandbox/managers/sandbox.manager.ts` (11 sites, 4 query builders)

**QB 1 — `autostopCheck` (line 106)**

```
Current:  .createQueryBuilder('sandbox')
          .innerJoin('sandbox_last_activity', 'activity', ...)
          .where('sandbox."runnerId" = :runnerId', ...)
          .andWhere('sandbox.state = :state', ...)
          .andWhere('sandbox."desiredState" = :desiredState', ...)
          .andWhere('sandbox.pending != true')
          .orderBy('sandbox."lastBackupAt"', 'ASC')

Change to: .createAggregateQueryBuilder('sandbox')
           .innerJoin('sandbox_last_activity', 'activity', ...)
           .where('ss."runnerId" = :runnerId', ...)
           .andWhere('ss."state" = :state', ...)
           .andWhere('ss."desiredState" = :desiredState', ...)
           .andWhere('ss."pending" != true')
           .orderBy('sb."lastBackupAt"', 'ASC')
```

Note: `sandbox."organizationId"`, `sandbox."autoStopInterval"` stay as-is (config fields).

**QB 2 — `autoArchiveCheck` (line 180)**

```
Current:  .createQueryBuilder('sandbox')
          .innerJoin('sandbox_last_activity', 'activity', ...)
          .andWhere('sandbox.state = :state', ...)
          .andWhere('sandbox."desiredState" = :desiredState', ...)
          .andWhere('sandbox.pending != true')
          .orderBy('sandbox."lastBackupAt"', 'ASC')

Change to: .createAggregateQueryBuilder('sandbox')
           .innerJoin('sandbox_last_activity', 'activity', ...)
           .andWhere('ss."state" = :state', ...)
           .andWhere('ss."desiredState" = :desiredState', ...)
           .andWhere('ss."pending" != true')
           .orderBy('sb."lastBackupAt"', 'ASC')
```

**QB 3 — `autoDeleteCheck` (line 243)**

```
Current:  .createQueryBuilder('sandbox')
          .innerJoin('sandbox_last_activity', 'activity', ...)
          .where('sandbox."runnerId" = :runnerId', ...)
          .andWhere('sandbox.state = :state', ...)
          .andWhere('sandbox."desiredState" = :desiredState', ...)
          .andWhere('sandbox.pending != true')

Change to: .createAggregateQueryBuilder('sandbox')
           .innerJoin('sandbox_last_activity', 'activity', ...)
           .where('ss."runnerId" = :runnerId', ...)
           .andWhere('ss."state" = :state', ...)
           .andWhere('ss."desiredState" = :desiredState', ...)
           .andWhere('ss."pending" != true')
```

**QB 4 — `syncStates` (line 700)**

```
Current:  .createQueryBuilder('sandbox')
          .select(['sandbox.id'])
          .leftJoin('sandbox_last_activity', 'activity', ...)
          .where('sandbox.state NOT IN (:...excludedStates)', ...)
          .andWhere('sandbox."desiredState"::text != sandbox.state::text')
          .andWhere('sandbox."desiredState"::text != :archived', ...)

Change to: .createAggregateQueryBuilder('sandbox')
           .select(['sandbox.id'])
           .leftJoin('sandbox_last_activity', 'activity', ...)
           .where('ss."state" NOT IN (:...excludedStates)', ...)
           .andWhere('ss."desiredState"::text != ss."state"::text')
           .andWhere('ss."desiredState"::text != :archived', ...)
```

### File 3: `apps/api/src/sandbox/services/sandbox-warm-pool.service.ts` (2 sites, 1 query builder)

**QB 1 — `fetchWarmPoolSandbox` (line 127)**

```
Current:  .createQueryBuilder('sandbox')
          .andWhere('sandbox.state = :state', { state: SandboxState.STARTED })
          .andWhere(`sandbox.runnerId NOT IN (${excludedRunnersSubquery.getQuery()})`)

Change to: .createAggregateQueryBuilder('sandbox')
           .andWhere('ss."state" = :state', { state: SandboxState.STARTED })
           .andWhere(`ss."runnerId" NOT IN (${excludedRunnersSubquery.getQuery()})`)
```

Note: `sandbox.class`, `sandbox.cpu`, `sandbox.mem`, `sandbox.disk`, `sandbox.snapshot`, `sandbox.osUser`, `sandbox.env`, `sandbox.organizationId`, `sandbox.region` stay as-is (config fields).

### File 4: `apps/api/src/sandbox/services/runner.service.ts` (3 sites, 1 query builder)

**QB 1 — `getRunnersWithMultipleSnapshotsBuilding` (line 816)**

```
Current:  .createQueryBuilder('sandbox')
          .select('sandbox.runnerId', 'runnerId')
          .where('sandbox.state = :state', { state: SandboxState.BUILDING_SNAPSHOT })
          .groupBy('sandbox.runnerId')
          .having('COUNT(DISTINCT sandbox.buildInfoSnapshotRef) > :maxSnapshotCount', ...)

Change to: .createAggregateQueryBuilder('sandbox')
           .select('ss."runnerId"', 'runnerId')
           .where('ss."state" = :state', { state: SandboxState.BUILDING_SNAPSHOT })
           .groupBy('ss."runnerId"')
           .having('COUNT(DISTINCT sandbox."buildInfoSnapshotRef") > :maxSnapshotCount', ...)
```

Note: `sandbox.buildInfoSnapshotRef` stays as-is — it is a column on the `sandbox` table (FK to build_info).

### File 5: `apps/api/src/organization/services/organization-usage.service.ts` (6 sites, 1 query builder)

**QB 1 — `fetchSandboxUsageFromDb` (line 614)**
This is the most complex one. The query uses `sandbox.state` and `sandbox."desiredState"` in CASE expressions.

```
Current:  .createQueryBuilder('sandbox')
          .select([
            `SUM(CASE WHEN sandbox.state IN (:...statesConsumingCompute) OR (sandbox.state = :resizingState AND sandbox."desiredState" = :startedDesiredState) THEN sandbox.cpu ELSE 0 END) as used_cpu`,
            `SUM(CASE WHEN sandbox.state IN (:...statesConsumingCompute) OR (sandbox.state = :resizingState AND sandbox."desiredState" = :startedDesiredState) THEN sandbox.mem ELSE 0 END) as used_mem`,
            'SUM(CASE WHEN sandbox.state IN (:...statesConsumingDisk) THEN sandbox.disk ELSE 0 END) as used_disk',
          ])
          .where('sandbox.organizationId = :organizationId', ...)
          .andWhere('sandbox.region = :regionId', ...)

Change to: .createAggregateQueryBuilder('sandbox')
           .select([
             `SUM(CASE WHEN ss."state" IN (:...statesConsumingCompute) OR (ss."state" = :resizingState AND ss."desiredState" = :startedDesiredState) THEN sandbox.cpu ELSE 0 END) as used_cpu`,
             `SUM(CASE WHEN ss."state" IN (:...statesConsumingCompute) OR (ss."state" = :resizingState AND ss."desiredState" = :startedDesiredState) THEN sandbox.mem ELSE 0 END) as used_mem`,
             'SUM(CASE WHEN ss."state" IN (:...statesConsumingDisk) THEN sandbox.disk ELSE 0 END) as used_disk',
           ])
           .where('sandbox."organizationId" = :organizationId', ...)
           .andWhere('sandbox.region = :regionId', ...)
```

Note: `sandbox.cpu`, `sandbox.mem`, `sandbox.disk` stay as-is (config table). Only state/desiredState references change to `ss`.

## Alias Mapping Cheat Sheet

| Old alias | New alias | Fields |
|---|---|---|
| `sandbox.state` | `ss."state"` | |
| `sandbox."desiredState"` | `ss."desiredState"` | |
| `sandbox.pending` | `ss."pending"` | |
| `sandbox."runnerId"` | `ss."runnerId"` | |
| `sandbox."prevRunnerId"` | `ss."prevRunnerId"` | |
| `sandbox."errorReason"` | `ss."errorReason"` | |
| `sandbox."recoverable"` | `ss."recoverable"` | |
| `sandbox."daemonVersion"` | `ss."daemonVersion"` | |
| `sandbox.backupState` | `sb."backupState"` | |
| `sandbox."backupSnapshot"` | `sb."backupSnapshot"` | |
| `sandbox."backupRegistryId"` | `sb."backupRegistryId"` | |
| `sandbox.lastBackupAt` | `sb."lastBackupAt"` | |
| `sandbox."backupErrorReason"` | `sb."backupErrorReason"` | |
| `sandbox."existingBackupSnapshots"` | `sb."existingBackupSnapshots"` | |

Config fields that stay as `sandbox."..."`:
`id`, `organizationId`, `name`, `region`, `class`, `snapshot`, `osUser`, `env`, `public`, `networkBlockAll`, `networkAllowList`, `labels`, `cpu`, `gpu`, `mem`, `disk`, `volumes`, `createdAt`, `updatedAt`, `autoStopInterval`, `autoArchiveInterval`, `autoDeleteInterval`, `authToken`, `buildInfoSnapshotRef`

## Execution Order

1. Fix each query builder site (21 total across 5 files)
2. For each: change `createQueryBuilder('sandbox')` → `createAggregateQueryBuilder('sandbox')`
3. For each: update SQL alias for state fields (`sandbox.X` → `ss."X"`) and backup fields (`sandbox.X` → `sb."X"`)
4. For each: fix runner JOIN conditions from `r.id = sandbox.runnerId` → `r.id = ss."runnerId"`
5. After all files: run `npx tsc --noEmit --project apps/api/tsconfig.app.json` (should still be 0 errors — these are string changes)
6. Test runtime: the queries must return the same results as before

## Gotchas

1. **Runner JOINs**: Three query builders in `backup.manager.ts` have `.innerJoin('runner', 'r', 'r.id = sandbox.runnerId')`. After the switch, `runnerId` is on `ss`, so the condition becomes `r.id = ss."runnerId"`. If you use `createAggregateQueryBuilder`, the `ss` alias is already available from the automatic inner join.

2. **CASE expressions**: The `checkBackupStates` QB in `backup.manager.ts` line 163 has `CASE sandbox.state WHEN :archiving THEN 1 ...`. This must change to `CASE ss."state" WHEN :archiving THEN 1 ...`.

3. **Self-referencing comparison**: `syncStates` in `sandbox.manager.ts` line 712 has `sandbox."desiredState"::text != sandbox.state::text`. Both sides reference state fields: `ss."desiredState"::text != ss."state"::text`.

4. **Subquery in warm pool**: `sandbox-warm-pool.service.ts` line 141 has `` sandbox.runnerId NOT IN (${excludedRunnersSubquery.getQuery()}) ``. The subquery selects runner IDs, and this filter is against the state table's `runnerId`: `` ss."runnerId" NOT IN (${excludedRunnersSubquery.getQuery()}) ``.

5. **GROUP BY + HAVING in runner.service.ts**: `getRunnersWithMultipleSnapshotsBuilding` uses `sandbox.runnerId` in both `.select()`, `.groupBy()`, and implicitly in the raw query. All need to change to `ss."runnerId"`.

6. **Double innerJoin conflict**: `createAggregateQueryBuilder` already does `innerJoin('sandbox_state', 'ss', ...)`. If any QB also has its own join to `sandbox_state`, that will conflict. Verify none of the 10 QBs do their own `sandbox_state` join. (They don't — they only join `runner` and `sandbox_last_activity`.)

7. **No TypeScript safety net**: All of these changes are string edits. tsc will NOT catch regressions. The only way to verify is to run the actual queries against the database, or to write integration tests that exercise these code paths.
