/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1779192782000 implements MigrationInterface {
  name = 'Migration1779192782000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Trigger propagates state-machine columns from sandbox to sandbox_lifecycle
    // on INSERT or UPDATE of the watched columns. Goes dormant once the app
    // stops writing those columns directly to sandbox.
    //
    // Phase mapping (DESTROYED/ARCHIVED -> terminal, else -> active) MUST stay
    // in sync with SandboxLifecycle.phaseFor() in TypeScript.
    //
    // The UPDATE branch uses DELETE-then-INSERT to handle cross-phase relocation:
    // if the row's phase changes, the old partition's row is deleted and a new
    // row inserted in the correct partition. Otherwise the DELETE matches
    // nothing and the INSERT's ON CONFLICT DO UPDATE updates in place.
    //
    // ON CONFLICT DO UPDATE also handles races with the catch-up backfill below.
    //
    // Risk: once the app has switched to writing sandbox_lifecycle directly,
    // any remaining code path that writes a watched column to sandbox would
    // fire this trigger and overwrite the fresh lifecycle.X with the
    // now-stale sandbox.X. The repository's splitUpdateFields routing
    // prevents this — raw SQL or unmigrated code paths must not violate it.
    await queryRunner.query(`
      CREATE OR REPLACE FUNCTION sync_sandbox_lifecycle()
      RETURNS TRIGGER AS $$
      DECLARE
        new_phase text;
      BEGIN
        IF NEW."state" IN ('destroyed', 'archived') THEN
          new_phase := 'terminal';
        ELSE
          new_phase := 'active';
        END IF;

        IF TG_OP = 'INSERT' THEN
          INSERT INTO sandbox_lifecycle (
            "sandboxId", "lifecyclePhase", "organizationId",
            "state", "desiredState", "pending", "errorReason", "recoverable",
            "daemonVersion", "runnerId", "prevRunnerId",
            "backupState", "lastBackupAt", "backupSnapshot",
            "backupRegistryId", "backupErrorReason", "existingBackupSnapshots",
            "updatedAt"
          )
          VALUES (
            NEW.id, new_phase, NEW."organizationId",
            NEW."state", NEW."desiredState", NEW."pending", NEW."errorReason", NEW."recoverable",
            NEW."daemonVersion", NEW."runnerId", NEW."prevRunnerId",
            NEW."backupState", NEW."lastBackupAt", NEW."backupSnapshot",
            NEW."backupRegistryId", NEW."backupErrorReason", NEW."existingBackupSnapshots",
            COALESCE(NEW."updatedAt", now())
          )
          ON CONFLICT ("sandboxId", "lifecyclePhase") DO UPDATE SET
            "organizationId"          = EXCLUDED."organizationId",
            "state"                   = EXCLUDED."state",
            "desiredState"            = EXCLUDED."desiredState",
            "pending"                 = EXCLUDED."pending",
            "errorReason"             = EXCLUDED."errorReason",
            "recoverable"             = EXCLUDED."recoverable",
            "daemonVersion"           = EXCLUDED."daemonVersion",
            "runnerId"                = EXCLUDED."runnerId",
            "prevRunnerId"            = EXCLUDED."prevRunnerId",
            "backupState"             = EXCLUDED."backupState",
            "lastBackupAt"            = EXCLUDED."lastBackupAt",
            "backupSnapshot"          = EXCLUDED."backupSnapshot",
            "backupRegistryId"        = EXCLUDED."backupRegistryId",
            "backupErrorReason"       = EXCLUDED."backupErrorReason",
            "existingBackupSnapshots" = EXCLUDED."existingBackupSnapshots",
            "updatedAt"               = EXCLUDED."updatedAt";
          RETURN NEW;
        END IF;

        IF TG_OP = 'UPDATE' THEN
          DELETE FROM sandbox_lifecycle
            WHERE "sandboxId" = NEW.id AND "lifecyclePhase" <> new_phase;

          INSERT INTO sandbox_lifecycle (
            "sandboxId", "lifecyclePhase", "organizationId",
            "state", "desiredState", "pending", "errorReason", "recoverable",
            "daemonVersion", "runnerId", "prevRunnerId",
            "backupState", "lastBackupAt", "backupSnapshot",
            "backupRegistryId", "backupErrorReason", "existingBackupSnapshots",
            "updatedAt"
          )
          VALUES (
            NEW.id, new_phase, NEW."organizationId",
            NEW."state", NEW."desiredState", NEW."pending", NEW."errorReason", NEW."recoverable",
            NEW."daemonVersion", NEW."runnerId", NEW."prevRunnerId",
            NEW."backupState", NEW."lastBackupAt", NEW."backupSnapshot",
            NEW."backupRegistryId", NEW."backupErrorReason", NEW."existingBackupSnapshots",
            COALESCE(NEW."updatedAt", now())
          )
          ON CONFLICT ("sandboxId", "lifecyclePhase") DO UPDATE SET
            "organizationId"          = EXCLUDED."organizationId",
            "state"                   = EXCLUDED."state",
            "desiredState"            = EXCLUDED."desiredState",
            "pending"                 = EXCLUDED."pending",
            "errorReason"             = EXCLUDED."errorReason",
            "recoverable"             = EXCLUDED."recoverable",
            "daemonVersion"           = EXCLUDED."daemonVersion",
            "runnerId"                = EXCLUDED."runnerId",
            "prevRunnerId"            = EXCLUDED."prevRunnerId",
            "backupState"             = EXCLUDED."backupState",
            "lastBackupAt"            = EXCLUDED."lastBackupAt",
            "backupSnapshot"          = EXCLUDED."backupSnapshot",
            "backupRegistryId"        = EXCLUDED."backupRegistryId",
            "backupErrorReason"       = EXCLUDED."backupErrorReason",
            "existingBackupSnapshots" = EXCLUDED."existingBackupSnapshots",
            "updatedAt"               = EXCLUDED."updatedAt";
          RETURN NEW;
        END IF;

        RETURN NEW;
      END;
      $$ LANGUAGE plpgsql;
    `)

    await queryRunner.query(`
      CREATE TRIGGER sandbox_lifecycle_sync
      AFTER INSERT OR UPDATE OF
        "state", "desiredState", "pending", "errorReason", "recoverable",
        "daemonVersion", "runnerId", "prevRunnerId",
        "backupState", "lastBackupAt", "backupSnapshot",
        "backupRegistryId", "backupErrorReason", "existingBackupSnapshots",
        "organizationId"
      ON "sandbox"
      FOR EACH ROW EXECUTE FUNCTION sync_sandbox_lifecycle();
    `)

    // Final catch-up backfill — covers only the delta since the standalone backfill ran.
    // DO NOT skip the standalone backfill: running this on an unbackfilled
    // table would lock sandbox for an extended period.
    // ON CONFLICT DO NOTHING respects rows already inserted by the trigger.
    await queryRunner.query(`
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
        CASE WHEN s."state" IN ('destroyed', 'archived') THEN 'terminal' ELSE 'active' END,
        s."organizationId",
        s."state", s."desiredState", s."pending", s."errorReason", s."recoverable",
        s."daemonVersion", s."runnerId", s."prevRunnerId",
        s."backupState", s."lastBackupAt", s."backupSnapshot",
        s."backupRegistryId", s."backupErrorReason", s."existingBackupSnapshots",
        COALESCE(s."updatedAt", now())
      FROM "sandbox" s
      ON CONFLICT ("sandboxId", "lifecyclePhase") DO NOTHING
    `)
  }

  // WARNING: safe to run only while the app is still writing state-machine
  // columns to `sandbox` (the trigger above keeps `sandbox_lifecycle` in sync).
  // Once the app has switched to writing `sandbox_lifecycle` directly,
  // `sandbox.state` and its sibling columns freeze in place and
  // `sandbox_lifecycle` becomes the only source of truth — at that point
  // the TRUNCATE below would discard every state transition that happened
  // post-switch. See the deploy runbook's rollback section before running.
  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP TRIGGER IF EXISTS sandbox_lifecycle_sync ON "sandbox"`)
    await queryRunner.query(`DROP FUNCTION IF EXISTS sync_sandbox_lifecycle()`)
    await queryRunner.query(`TRUNCATE "sandbox_lifecycle"`)
  }
}
