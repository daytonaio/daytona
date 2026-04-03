/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1775100000001 implements MigrationInterface {
  name = 'Migration1775100000001'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP TRIGGER IF EXISTS sandbox_backup_sync_to_old ON "sandbox_backup"`)
    await queryRunner.query(`DROP TRIGGER IF EXISTS sandbox_backup_sync_to_new ON "sandbox"`)
    await queryRunner.query(`DROP TRIGGER IF EXISTS sandbox_state_sync_to_old ON "sandbox_state"`)
    await queryRunner.query(`DROP TRIGGER IF EXISTS sandbox_state_sync_to_new ON "sandbox"`)
    await queryRunner.query(`DROP FUNCTION IF EXISTS sync_sandbox_backup_columns()`)
    await queryRunner.query(`DROP FUNCTION IF EXISTS sync_sandbox_state_columns()`)

    await queryRunner.query(`DROP INDEX IF EXISTS "sandbox_state_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "sandbox_desiredstate_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "sandbox_runnerid_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "sandbox_runner_state_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "sandbox_runner_state_desired_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "sandbox_pending_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "sandbox_active_only_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "sandbox_backupstate_idx"`)

    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "state"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "desiredState"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "pending"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "errorReason"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "recoverable"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "runnerId"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "prevRunnerId"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "daemonVersion"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "backupState"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "backupSnapshot"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "backupRegistryId"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "lastBackupAt"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "backupErrorReason"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "existingBackupSnapshots"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "state" sandbox_state_enum NOT NULL DEFAULT 'unknown'`)
    await queryRunner.query(
      `ALTER TABLE "sandbox" ADD "desiredState" sandbox_desired_state_enum NOT NULL DEFAULT 'started'`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "pending" boolean NOT NULL DEFAULT false`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "errorReason" character varying`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "recoverable" boolean NOT NULL DEFAULT false`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "runnerId" uuid`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "prevRunnerId" uuid`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "daemonVersion" character varying`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "backupState" backup_state_enum NOT NULL DEFAULT 'None'`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "backupSnapshot" character varying`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "backupRegistryId" character varying`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "lastBackupAt" TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "backupErrorReason" text`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "existingBackupSnapshots" jsonb NOT NULL DEFAULT '[]'::jsonb`)

    await queryRunner.query(`
      UPDATE "sandbox" s SET
        "state" = ss."state",
        "desiredState" = ss."desiredState",
        "pending" = ss."pending",
        "errorReason" = ss."errorReason",
        "recoverable" = ss."recoverable",
        "runnerId" = ss."runnerId",
        "prevRunnerId" = ss."prevRunnerId",
        "daemonVersion" = ss."daemonVersion"
      FROM "sandbox_state" ss WHERE s."id" = ss."sandboxId"
    `)

    await queryRunner.query(`
      UPDATE "sandbox" s SET
        "backupState" = sb."backupState",
        "backupSnapshot" = sb."backupSnapshot",
        "backupRegistryId" = sb."backupRegistryId",
        "lastBackupAt" = sb."lastBackupAt",
        "backupErrorReason" = sb."backupErrorReason",
        "existingBackupSnapshots" = sb."existingBackupSnapshots"
      FROM "sandbox_backup" sb WHERE s."id" = sb."sandboxId"
    `)

    await queryRunner.query(`CREATE INDEX "sandbox_state_idx" ON "sandbox" ("state")`)
    await queryRunner.query(`CREATE INDEX "sandbox_desiredstate_idx" ON "sandbox" ("desiredState")`)
    await queryRunner.query(`CREATE INDEX "sandbox_runnerid_idx" ON "sandbox" ("runnerId")`)
    await queryRunner.query(`CREATE INDEX "sandbox_runner_state_idx" ON "sandbox" ("runnerId", "state")`)
    await queryRunner.query(
      `CREATE INDEX "sandbox_runner_state_desired_idx" ON "sandbox" ("runnerId", "state", "desiredState") WHERE "pending" = false`,
    )
    await queryRunner.query(
      `CREATE INDEX "sandbox_active_only_idx" ON "sandbox" ("id") WHERE "state" <> ALL (ARRAY['destroyed'::sandbox_state_enum, 'archived'::sandbox_state_enum])`,
    )
    await queryRunner.query(`CREATE INDEX "sandbox_pending_idx" ON "sandbox" ("id") WHERE "pending" = true`)
    await queryRunner.query(`CREATE INDEX "sandbox_backupstate_idx" ON "sandbox" ("backupState")`)

    await queryRunner.query(`
      CREATE OR REPLACE FUNCTION sync_sandbox_state_columns()
      RETURNS TRIGGER AS $$
      BEGIN
        IF TG_TABLE_NAME = 'sandbox' THEN
          INSERT INTO sandbox_state ("sandboxId", "state", "desiredState", "pending", "errorReason", "recoverable", "runnerId", "prevRunnerId", "daemonVersion", "updatedAt")
          VALUES (NEW.id, NEW.state, NEW."desiredState", NEW.pending, NEW."errorReason", NEW.recoverable, NEW."runnerId", NEW."prevRunnerId", NEW."daemonVersion", NOW())
          ON CONFLICT ("sandboxId") DO UPDATE SET
            "state" = EXCLUDED."state", "desiredState" = EXCLUDED."desiredState", "pending" = EXCLUDED."pending",
            "errorReason" = EXCLUDED."errorReason", "recoverable" = EXCLUDED."recoverable", "runnerId" = EXCLUDED."runnerId",
            "prevRunnerId" = EXCLUDED."prevRunnerId", "daemonVersion" = EXCLUDED."daemonVersion", "updatedAt" = EXCLUDED."updatedAt"
          WHERE sandbox_state."state" IS DISTINCT FROM EXCLUDED."state"
            OR sandbox_state."desiredState" IS DISTINCT FROM EXCLUDED."desiredState"
            OR sandbox_state."pending" IS DISTINCT FROM EXCLUDED."pending"
            OR sandbox_state."runnerId" IS DISTINCT FROM EXCLUDED."runnerId";
        ELSIF TG_TABLE_NAME = 'sandbox_state' THEN
          UPDATE sandbox SET "state" = NEW."state", "desiredState" = NEW."desiredState", "pending" = NEW."pending",
            "errorReason" = NEW."errorReason", "recoverable" = NEW."recoverable", "runnerId" = NEW."runnerId",
            "prevRunnerId" = NEW."prevRunnerId", "daemonVersion" = NEW."daemonVersion"
          WHERE "id" = NEW."sandboxId" AND ("state" IS DISTINCT FROM NEW."state" OR "desiredState" IS DISTINCT FROM NEW."desiredState");
        END IF;
        RETURN NEW;
      END;
      $$ LANGUAGE plpgsql
    `)

    await queryRunner.query(`
      CREATE OR REPLACE FUNCTION sync_sandbox_backup_columns()
      RETURNS TRIGGER AS $$
      BEGIN
        IF TG_TABLE_NAME = 'sandbox' THEN
          INSERT INTO sandbox_backup ("sandboxId", "backupState", "backupSnapshot", "backupRegistryId", "lastBackupAt", "backupErrorReason", "existingBackupSnapshots", "updatedAt")
          VALUES (NEW.id, NEW."backupState", NEW."backupSnapshot", NEW."backupRegistryId", NEW."lastBackupAt", NEW."backupErrorReason", NEW."existingBackupSnapshots", NOW())
          ON CONFLICT ("sandboxId") DO UPDATE SET
            "backupState" = EXCLUDED."backupState", "backupSnapshot" = EXCLUDED."backupSnapshot",
            "backupRegistryId" = EXCLUDED."backupRegistryId", "lastBackupAt" = EXCLUDED."lastBackupAt",
            "backupErrorReason" = EXCLUDED."backupErrorReason", "existingBackupSnapshots" = EXCLUDED."existingBackupSnapshots",
            "updatedAt" = EXCLUDED."updatedAt"
          WHERE sandbox_backup."backupState" IS DISTINCT FROM EXCLUDED."backupState"
            OR sandbox_backup."backupSnapshot" IS DISTINCT FROM EXCLUDED."backupSnapshot";
        ELSIF TG_TABLE_NAME = 'sandbox_backup' THEN
          UPDATE sandbox SET "backupState" = NEW."backupState", "backupSnapshot" = NEW."backupSnapshot",
            "backupRegistryId" = NEW."backupRegistryId", "lastBackupAt" = NEW."lastBackupAt",
            "backupErrorReason" = NEW."backupErrorReason", "existingBackupSnapshots" = NEW."existingBackupSnapshots"
          WHERE "id" = NEW."sandboxId" AND "backupState" IS DISTINCT FROM NEW."backupState";
        END IF;
        RETURN NEW;
      END;
      $$ LANGUAGE plpgsql
    `)

    await queryRunner.query(
      `CREATE TRIGGER sandbox_state_sync_to_new AFTER INSERT OR UPDATE OF "state", "desiredState", "pending", "errorReason", "recoverable", "runnerId", "prevRunnerId", "daemonVersion" ON "sandbox" FOR EACH ROW EXECUTE FUNCTION sync_sandbox_state_columns()`,
    )
    await queryRunner.query(
      `CREATE TRIGGER sandbox_state_sync_to_old AFTER INSERT OR UPDATE ON "sandbox_state" FOR EACH ROW EXECUTE FUNCTION sync_sandbox_state_columns()`,
    )
    await queryRunner.query(
      `CREATE TRIGGER sandbox_backup_sync_to_new AFTER INSERT OR UPDATE OF "backupState", "backupSnapshot", "backupRegistryId", "lastBackupAt", "backupErrorReason", "existingBackupSnapshots" ON "sandbox" FOR EACH ROW EXECUTE FUNCTION sync_sandbox_backup_columns()`,
    )
    await queryRunner.query(
      `CREATE TRIGGER sandbox_backup_sync_to_old AFTER INSERT OR UPDATE ON "sandbox_backup" FOR EACH ROW EXECUTE FUNCTION sync_sandbox_backup_columns()`,
    )
  }
}
