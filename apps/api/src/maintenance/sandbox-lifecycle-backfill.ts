/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DataSource } from 'typeorm'
import { baseDataSourceOptions } from '../migrations/data-source'

const DEFAULT_BATCH_SIZE = 10_000
const DEFAULT_PROGRESS_INTERVAL = 10

function parseEnvInt(name: string, fallback: number): number {
  const raw = process.env[name]
  if (!raw) return fallback
  const n = parseInt(raw, 10)
  if (!Number.isFinite(n) || n <= 0) {
    throw new Error(`Invalid value for ${name}: ${raw}`)
  }
  return n
}

function parseEnvBool(name: string, fallback: boolean): boolean {
  const raw = process.env[name]
  if (!raw) return fallback
  const normalized = raw.trim().toLowerCase()
  return normalized === 'true' || normalized === '1' || normalized === 'yes'
}

async function run(): Promise<void> {
  const batchSize = parseEnvInt('BACKFILL_BATCH_SIZE', DEFAULT_BATCH_SIZE)
  const progressInterval = parseEnvInt('BACKFILL_PROGRESS_INTERVAL', DEFAULT_PROGRESS_INTERVAL)
  const dryRun = parseEnvBool('BACKFILL_DRY_RUN', false)

  const dataSource = new DataSource(baseDataSourceOptions)
  await dataSource.initialize()
  const startTime = Date.now()

  try {
    const totalRows: number = await dataSource
      .query(`SELECT COUNT(*)::int AS count FROM "sandbox"`)
      .then((rows: Array<{ count: number }>) => rows[0]?.count ?? 0)

    const existingLifecycleRows: number = await dataSource
      .query(`SELECT COUNT(*)::int AS count FROM "sandbox_lifecycle"`)
      .then((rows: Array<{ count: number }>) => rows[0]?.count ?? 0)

    console.log(
      `[sandbox-lifecycle-backfill] starting${dryRun ? ' (DRY RUN — no writes)' : ''}: ` +
        `${totalRows} rows in sandbox, ${existingLifecycleRows} rows in sandbox_lifecycle, ` +
        `batch size ${batchSize}`,
    )

    let lastId: string | null = null
    let processed = 0
    let inserted = 0
    let batchNumber = 0

    for (;;) {
      const rows: Array<{ inserted_id: string }> = dryRun
        ? // Dry-run mode: count rows that would be inserted (sandboxes not yet in
          // lifecycle), without writing anything.
          await dataSource.query(
            `
              SELECT s.id AS inserted_id
              FROM "sandbox" s
              WHERE ($1::varchar IS NULL OR s.id > $1::varchar)
                AND NOT EXISTS (
                  SELECT 1 FROM "sandbox_lifecycle" l WHERE l."sandboxId" = s.id
                )
              ORDER BY s.id
              LIMIT $2
            `,
            [lastId, batchSize],
          )
        : await dataSource.query(
            `
              INSERT INTO "sandbox_lifecycle" (
                "sandboxId", "lifecyclePhase", "organizationId",
                "state", "desiredState", "pending", "errorReason", "recoverable",
                "daemonVersion", "runnerId", "prevRunnerId",
                "backupState", "lastBackupAt", "backupSnapshot",
                "backupRegistryId", "backupErrorReason", "existingBackupSnapshots",
                "updatedAt"
              )
              SELECT
                s.id,
                CASE WHEN s."state" IN ('destroyed'::sandbox_state_enum, 'archived'::sandbox_state_enum)
                  THEN 'terminal' ELSE 'active' END,
                s."organizationId",
                s."state", s."desiredState", s."pending", s."errorReason", s."recoverable",
                s."daemonVersion", s."runnerId", s."prevRunnerId",
                s."backupState", s."lastBackupAt", s."backupSnapshot",
                s."backupRegistryId", s."backupErrorReason", s."existingBackupSnapshots",
                COALESCE(s."updatedAt", now())
              FROM "sandbox" s
              WHERE ($1::varchar IS NULL OR s.id > $1::varchar)
              ORDER BY s.id
              LIMIT $2
              ON CONFLICT ("sandboxId", "lifecyclePhase") DO NOTHING
              RETURNING "sandboxId" AS inserted_id
            `,
            [lastId, batchSize],
          )

      const lastIdFromSandbox: Array<{ id: string }> = await dataSource.query(
        `
          SELECT id FROM "sandbox"
          WHERE ($1::varchar IS NULL OR id > $1::varchar)
          ORDER BY id
          LIMIT $2
        `,
        [lastId, batchSize],
      )

      if (lastIdFromSandbox.length === 0) {
        break
      }

      const advance = lastIdFromSandbox[lastIdFromSandbox.length - 1].id
      processed += lastIdFromSandbox.length
      inserted += rows.length
      batchNumber += 1
      lastId = advance

      if (batchNumber % progressInterval === 0 || lastIdFromSandbox.length < batchSize) {
        const elapsedMs = Date.now() - startTime
        const rate = processed / (elapsedMs / 1000)
        const remaining = Math.max(0, totalRows - processed)
        const etaSec = rate > 0 ? Math.round(remaining / rate) : 0
        console.log(
          `[sandbox-lifecycle-backfill] batch=${batchNumber} processed=${processed}/${totalRows} ` +
            `inserted=${inserted} skipped=${processed - inserted} ` +
            `rate=${rate.toFixed(0)} rows/s eta=${etaSec}s`,
        )
      }

      if (lastIdFromSandbox.length < batchSize) {
        break
      }
    }

    const finalLifecycleRows: number = await dataSource
      .query(`SELECT COUNT(*)::int AS count FROM "sandbox_lifecycle"`)
      .then((rows: Array<{ count: number }>) => rows[0]?.count ?? 0)

    const elapsedSec = ((Date.now() - startTime) / 1000).toFixed(1)
    console.log(
      `[sandbox-lifecycle-backfill] done in ${elapsedSec}s${dryRun ? ' (DRY RUN)' : ''}. ` +
        `Processed ${processed} sandbox rows, ${dryRun ? 'would insert' : 'inserted'} ${inserted} ` +
        `new lifecycle rows. Total lifecycle rows: ${finalLifecycleRows} (was ${existingLifecycleRows}).`,
    )

    if (!dryRun && finalLifecycleRows !== totalRows) {
      console.warn(
        `[sandbox-lifecycle-backfill] WARNING: row-count parity not yet satisfied ` +
          `(sandbox=${totalRows}, sandbox_lifecycle=${finalLifecycleRows}). ` +
          `This is expected if writes happened during backfill; re-run to converge.`,
      )
    }
  } finally {
    await dataSource.destroy()
  }
}

run()
  .then(() => process.exit(0))
  .catch((err) => {
    console.error('[sandbox-lifecycle-backfill] fatal:', err)
    process.exit(1)
  })
