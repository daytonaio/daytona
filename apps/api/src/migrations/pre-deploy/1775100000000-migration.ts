/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1775100000000 implements MigrationInterface {
  name = 'Migration1775100000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // ── Create sandbox_state table ──────────────────────────────────────
    await queryRunner.query(`
      CREATE TABLE "sandbox_state" (
        "sandboxId" uuid NOT NULL PRIMARY KEY REFERENCES "sandbox"("id") ON DELETE CASCADE,
        "state" sandbox_state_enum NOT NULL DEFAULT 'unknown',
        "desiredState" sandbox_desired_state_enum NOT NULL DEFAULT 'started',
        "pending" boolean NOT NULL DEFAULT false,
        "errorReason" character varying,
        "recoverable" boolean NOT NULL DEFAULT false,
        "runnerId" uuid,
        "prevRunnerId" uuid,
        "daemonVersion" character varying,
        "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
      )
    `)

    // ── Create sandbox_backup table ─────────────────────────────────────
    await queryRunner.query(`
      CREATE TABLE "sandbox_backup" (
        "sandboxId" uuid NOT NULL PRIMARY KEY REFERENCES "sandbox"("id") ON DELETE CASCADE,
        "backupState" backup_state_enum NOT NULL DEFAULT 'None',
        "backupSnapshot" character varying,
        "backupRegistryId" character varying,
        "lastBackupAt" TIMESTAMP WITH TIME ZONE,
        "backupErrorReason" text,
        "existingBackupSnapshots" jsonb NOT NULL DEFAULT '[]'::jsonb,
        "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
      )
    `)

    // ── Backfill sandbox_state from sandbox ─────────────────────────────
    await queryRunner.query(`
      INSERT INTO "sandbox_state" (
        "sandboxId", "state", "desiredState", "pending", "errorReason",
        "recoverable", "runnerId", "prevRunnerId", "daemonVersion", "updatedAt"
      )
      SELECT
        "id", "state", "desiredState", "pending", "errorReason",
        "recoverable", "runnerId", "prevRunnerId", "daemonVersion", "updatedAt"
      FROM "sandbox"
      ORDER BY "id"
      ON CONFLICT DO NOTHING
    `)

    // ── Backfill sandbox_backup from sandbox ────────────────────────────
    await queryRunner.query(`
      INSERT INTO "sandbox_backup" (
        "sandboxId", "backupState", "backupSnapshot", "backupRegistryId",
        "lastBackupAt", "backupErrorReason", "existingBackupSnapshots", "updatedAt"
      )
      SELECT
        "id", "backupState", "backupSnapshot", "backupRegistryId",
        "lastBackupAt", "backupErrorReason", "existingBackupSnapshots", "updatedAt"
      FROM "sandbox"
      ORDER BY "id"
      ON CONFLICT DO NOTHING
    `)

    // ── Create indexes on sandbox_state ─────────────────────────────────
    await queryRunner.query(`CREATE INDEX "ss_state_idx" ON "sandbox_state" ("state")`)
    await queryRunner.query(`CREATE INDEX "ss_desiredstate_idx" ON "sandbox_state" ("desiredState")`)
    await queryRunner.query(`CREATE INDEX "ss_runnerid_idx" ON "sandbox_state" ("runnerId")`)
    await queryRunner.query(`CREATE INDEX "ss_runner_state_idx" ON "sandbox_state" ("runnerId", "state")`)
    await queryRunner.query(
      `CREATE INDEX "ss_runner_state_desired_idx" ON "sandbox_state" ("runnerId", "state", "desiredState") WHERE "pending" = false`,
    )
    await queryRunner.query(
      `CREATE INDEX "ss_active_only_idx" ON "sandbox_state" ("sandboxId") WHERE "state" <> ALL (ARRAY['destroyed'::sandbox_state_enum, 'archived'::sandbox_state_enum])`,
    )
    await queryRunner.query(`CREATE INDEX "ss_pending_idx" ON "sandbox_state" ("sandboxId") WHERE "pending" = true`)

    // ── Create indexes on sandbox_backup ────────────────────────────────
    await queryRunner.query(`CREATE INDEX "sb_backupstate_idx" ON "sandbox_backup" ("backupState")`)

    // ── Dual-write sync function for sandbox_state ──────────────────────
    await queryRunner.query(`
      CREATE OR REPLACE FUNCTION sync_sandbox_state_columns()
      RETURNS TRIGGER AS $$
      BEGIN
        IF TG_TABLE_NAME = 'sandbox' THEN
          INSERT INTO sandbox_state (
            "sandboxId", "state", "desiredState", "pending", "errorReason",
            "recoverable", "runnerId", "prevRunnerId", "daemonVersion", "updatedAt"
          ) VALUES (
            NEW.id, NEW.state, NEW."desiredState", NEW.pending, NEW."errorReason",
            NEW.recoverable, NEW."runnerId", NEW."prevRunnerId", NEW."daemonVersion", NOW()
          )
          ON CONFLICT ("sandboxId") DO UPDATE SET
            "state" = EXCLUDED."state",
            "desiredState" = EXCLUDED."desiredState",
            "pending" = EXCLUDED."pending",
            "errorReason" = EXCLUDED."errorReason",
            "recoverable" = EXCLUDED."recoverable",
            "runnerId" = EXCLUDED."runnerId",
            "prevRunnerId" = EXCLUDED."prevRunnerId",
            "daemonVersion" = EXCLUDED."daemonVersion",
            "updatedAt" = EXCLUDED."updatedAt"
          WHERE sandbox_state."state" IS DISTINCT FROM EXCLUDED."state"
            OR sandbox_state."desiredState" IS DISTINCT FROM EXCLUDED."desiredState"
            OR sandbox_state."pending" IS DISTINCT FROM EXCLUDED."pending"
            OR sandbox_state."errorReason" IS DISTINCT FROM EXCLUDED."errorReason"
            OR sandbox_state."recoverable" IS DISTINCT FROM EXCLUDED."recoverable"
            OR sandbox_state."runnerId" IS DISTINCT FROM EXCLUDED."runnerId"
            OR sandbox_state."prevRunnerId" IS DISTINCT FROM EXCLUDED."prevRunnerId"
            OR sandbox_state."daemonVersion" IS DISTINCT FROM EXCLUDED."daemonVersion";
        ELSIF TG_TABLE_NAME = 'sandbox_state' THEN
          UPDATE sandbox SET
            "state" = NEW."state",
            "desiredState" = NEW."desiredState",
            "pending" = NEW."pending",
            "errorReason" = NEW."errorReason",
            "recoverable" = NEW."recoverable",
            "runnerId" = NEW."runnerId",
            "prevRunnerId" = NEW."prevRunnerId",
            "daemonVersion" = NEW."daemonVersion"
          WHERE "id" = NEW."sandboxId"
            AND ("state" IS DISTINCT FROM NEW."state"
              OR "desiredState" IS DISTINCT FROM NEW."desiredState"
              OR "pending" IS DISTINCT FROM NEW."pending"
              OR "errorReason" IS DISTINCT FROM NEW."errorReason"
              OR "recoverable" IS DISTINCT FROM NEW."recoverable"
              OR "runnerId" IS DISTINCT FROM NEW."runnerId"
              OR "prevRunnerId" IS DISTINCT FROM NEW."prevRunnerId"
              OR "daemonVersion" IS DISTINCT FROM NEW."daemonVersion");
        END IF;
        RETURN NEW;
      END;
      $$ LANGUAGE plpgsql
    `)

    // ── Dual-write sync function for sandbox_backup ─────────────────────
    await queryRunner.query(`
      CREATE OR REPLACE FUNCTION sync_sandbox_backup_columns()
      RETURNS TRIGGER AS $$
      BEGIN
        IF TG_TABLE_NAME = 'sandbox' THEN
          INSERT INTO sandbox_backup (
            "sandboxId", "backupState", "backupSnapshot", "backupRegistryId",
            "lastBackupAt", "backupErrorReason", "existingBackupSnapshots", "updatedAt"
          ) VALUES (
            NEW.id, NEW."backupState", NEW."backupSnapshot", NEW."backupRegistryId",
            NEW."lastBackupAt", NEW."backupErrorReason", NEW."existingBackupSnapshots", NOW()
          )
          ON CONFLICT ("sandboxId") DO UPDATE SET
            "backupState" = EXCLUDED."backupState",
            "backupSnapshot" = EXCLUDED."backupSnapshot",
            "backupRegistryId" = EXCLUDED."backupRegistryId",
            "lastBackupAt" = EXCLUDED."lastBackupAt",
            "backupErrorReason" = EXCLUDED."backupErrorReason",
            "existingBackupSnapshots" = EXCLUDED."existingBackupSnapshots",
            "updatedAt" = EXCLUDED."updatedAt"
          WHERE sandbox_backup."backupState" IS DISTINCT FROM EXCLUDED."backupState"
            OR sandbox_backup."backupSnapshot" IS DISTINCT FROM EXCLUDED."backupSnapshot"
            OR sandbox_backup."backupRegistryId" IS DISTINCT FROM EXCLUDED."backupRegistryId"
            OR sandbox_backup."lastBackupAt" IS DISTINCT FROM EXCLUDED."lastBackupAt"
            OR sandbox_backup."backupErrorReason" IS DISTINCT FROM EXCLUDED."backupErrorReason"
            OR sandbox_backup."existingBackupSnapshots" IS DISTINCT FROM EXCLUDED."existingBackupSnapshots";
        ELSIF TG_TABLE_NAME = 'sandbox_backup' THEN
          UPDATE sandbox SET
            "backupState" = NEW."backupState",
            "backupSnapshot" = NEW."backupSnapshot",
            "backupRegistryId" = NEW."backupRegistryId",
            "lastBackupAt" = NEW."lastBackupAt",
            "backupErrorReason" = NEW."backupErrorReason",
            "existingBackupSnapshots" = NEW."existingBackupSnapshots"
          WHERE "id" = NEW."sandboxId"
            AND ("backupState" IS DISTINCT FROM NEW."backupState"
              OR "backupSnapshot" IS DISTINCT FROM NEW."backupSnapshot"
              OR "backupRegistryId" IS DISTINCT FROM NEW."backupRegistryId"
              OR "lastBackupAt" IS DISTINCT FROM NEW."lastBackupAt"
              OR "backupErrorReason" IS DISTINCT FROM NEW."backupErrorReason"
              OR "existingBackupSnapshots" IS DISTINCT FROM NEW."existingBackupSnapshots");
        END IF;
        RETURN NEW;
      END;
      $$ LANGUAGE plpgsql
    `)

    // ── Triggers: sandbox → sandbox_state (old → new) ───────────────────
    await queryRunner.query(`
      CREATE TRIGGER sandbox_state_sync_to_new
      AFTER INSERT OR UPDATE OF "state", "desiredState", "pending", "errorReason", "recoverable", "runnerId", "prevRunnerId", "daemonVersion"
      ON "sandbox"
      FOR EACH ROW EXECUTE FUNCTION sync_sandbox_state_columns()
    `)

    // ── Triggers: sandbox_state → sandbox (new → old) ───────────────────
    await queryRunner.query(`
      CREATE TRIGGER sandbox_state_sync_to_old
      AFTER INSERT OR UPDATE ON "sandbox_state"
      FOR EACH ROW EXECUTE FUNCTION sync_sandbox_state_columns()
    `)

    // ── Triggers: sandbox → sandbox_backup (old → new) ──────────────────
    await queryRunner.query(`
      CREATE TRIGGER sandbox_backup_sync_to_new
      AFTER INSERT OR UPDATE OF "backupState", "backupSnapshot", "backupRegistryId", "lastBackupAt", "backupErrorReason", "existingBackupSnapshots"
      ON "sandbox"
      FOR EACH ROW EXECUTE FUNCTION sync_sandbox_backup_columns()
    `)

    // ── Triggers: sandbox_backup → sandbox (new → old) ──────────────────
    await queryRunner.query(`
      CREATE TRIGGER sandbox_backup_sync_to_old
      AFTER INSERT OR UPDATE ON "sandbox_backup"
      FOR EACH ROW EXECUTE FUNCTION sync_sandbox_backup_columns()
    `)

    // ── Re-sync to handle race window between backfill and trigger creation
    await queryRunner.query(`
      INSERT INTO "sandbox_state" (
        "sandboxId", "state", "desiredState", "pending", "errorReason",
        "recoverable", "runnerId", "prevRunnerId", "daemonVersion", "updatedAt"
      )
      SELECT
        "id", "state", "desiredState", "pending", "errorReason",
        "recoverable", "runnerId", "prevRunnerId", "daemonVersion", "updatedAt"
      FROM "sandbox"
      ON CONFLICT ("sandboxId") DO UPDATE SET
        "state" = EXCLUDED."state",
        "desiredState" = EXCLUDED."desiredState",
        "pending" = EXCLUDED."pending",
        "errorReason" = EXCLUDED."errorReason",
        "recoverable" = EXCLUDED."recoverable",
        "runnerId" = EXCLUDED."runnerId",
        "prevRunnerId" = EXCLUDED."prevRunnerId",
        "daemonVersion" = EXCLUDED."daemonVersion",
        "updatedAt" = EXCLUDED."updatedAt"
      WHERE sandbox_state."state" IS DISTINCT FROM EXCLUDED."state"
        OR sandbox_state."desiredState" IS DISTINCT FROM EXCLUDED."desiredState"
        OR sandbox_state."pending" IS DISTINCT FROM EXCLUDED."pending"
        OR sandbox_state."errorReason" IS DISTINCT FROM EXCLUDED."errorReason"
        OR sandbox_state."recoverable" IS DISTINCT FROM EXCLUDED."recoverable"
        OR sandbox_state."runnerId" IS DISTINCT FROM EXCLUDED."runnerId"
        OR sandbox_state."prevRunnerId" IS DISTINCT FROM EXCLUDED."prevRunnerId"
        OR sandbox_state."daemonVersion" IS DISTINCT FROM EXCLUDED."daemonVersion"
    `)

    await queryRunner.query(`
      INSERT INTO "sandbox_backup" (
        "sandboxId", "backupState", "backupSnapshot", "backupRegistryId",
        "lastBackupAt", "backupErrorReason", "existingBackupSnapshots", "updatedAt"
      )
      SELECT
        "id", "backupState", "backupSnapshot", "backupRegistryId",
        "lastBackupAt", "backupErrorReason", "existingBackupSnapshots", "updatedAt"
      FROM "sandbox"
      ON CONFLICT ("sandboxId") DO UPDATE SET
        "backupState" = EXCLUDED."backupState",
        "backupSnapshot" = EXCLUDED."backupSnapshot",
        "backupRegistryId" = EXCLUDED."backupRegistryId",
        "lastBackupAt" = EXCLUDED."lastBackupAt",
        "backupErrorReason" = EXCLUDED."backupErrorReason",
        "existingBackupSnapshots" = EXCLUDED."existingBackupSnapshots",
        "updatedAt" = EXCLUDED."updatedAt"
      WHERE sandbox_backup."backupState" IS DISTINCT FROM EXCLUDED."backupState"
        OR sandbox_backup."backupSnapshot" IS DISTINCT FROM EXCLUDED."backupSnapshot"
        OR sandbox_backup."backupRegistryId" IS DISTINCT FROM EXCLUDED."backupRegistryId"
        OR sandbox_backup."lastBackupAt" IS DISTINCT FROM EXCLUDED."lastBackupAt"
        OR sandbox_backup."backupErrorReason" IS DISTINCT FROM EXCLUDED."backupErrorReason"
        OR sandbox_backup."existingBackupSnapshots" IS DISTINCT FROM EXCLUDED."existingBackupSnapshots"
    `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Drop dual-write triggers
    await queryRunner.query(`DROP TRIGGER IF EXISTS sandbox_backup_sync_to_old ON "sandbox_backup"`)
    await queryRunner.query(`DROP TRIGGER IF EXISTS sandbox_backup_sync_to_new ON "sandbox"`)
    await queryRunner.query(`DROP TRIGGER IF EXISTS sandbox_state_sync_to_old ON "sandbox_state"`)
    await queryRunner.query(`DROP TRIGGER IF EXISTS sandbox_state_sync_to_new ON "sandbox"`)

    // Drop sync functions
    await queryRunner.query(`DROP FUNCTION IF EXISTS sync_sandbox_backup_columns()`)
    await queryRunner.query(`DROP FUNCTION IF EXISTS sync_sandbox_state_columns()`)

    // Drop indexes
    await queryRunner.query(`DROP INDEX IF EXISTS "sb_backupstate_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "ss_pending_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "ss_active_only_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "ss_runner_state_desired_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "ss_runner_state_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "ss_runnerid_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "ss_desiredstate_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "ss_state_idx"`)

    // Drop tables (FK constraints are dropped automatically with the table)
    await queryRunner.query(`DROP TABLE IF EXISTS "sandbox_backup"`)
    await queryRunner.query(`DROP TABLE IF EXISTS "sandbox_state"`)
  }
}
