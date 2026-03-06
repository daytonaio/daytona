/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1772460237544 implements MigrationInterface {
  name = 'Migration1772460237544'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Drop dual-write triggers and function (no longer needed after deployment)
    await queryRunner.query(`DROP TRIGGER IF EXISTS sandbox_activity_sync_to_old ON "sandbox_last_activity"`)
    await queryRunner.query(`DROP TRIGGER IF EXISTS sandbox_activity_sync_to_new ON "sandbox"`)
    await queryRunner.query(`DROP FUNCTION IF EXISTS sync_sandbox_last_activity()`)

    // Drop the old column
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "lastActivityAt"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Restore the old column
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "lastActivityAt" TIMESTAMP WITH TIME ZONE`)

    // Backfill data from sandbox_last_activity to restored column
    await queryRunner.query(`
      UPDATE "sandbox" s
      SET "lastActivityAt" = sla."lastActivityAt"
      FROM "sandbox_last_activity" sla
      WHERE s.id = sla."sandboxId"
    `)

    // Recreate sync function for dual-write
    await queryRunner.query(`
      CREATE OR REPLACE FUNCTION sync_sandbox_last_activity()
      RETURNS TRIGGER AS $$
      BEGIN
        IF TG_TABLE_NAME = 'sandbox' THEN
          INSERT INTO sandbox_last_activity ("sandboxId", "lastActivityAt")
          VALUES (NEW.id, NEW."lastActivityAt")
          ON CONFLICT ("sandboxId") DO UPDATE SET "lastActivityAt" = EXCLUDED."lastActivityAt"
          WHERE sandbox_last_activity."lastActivityAt" IS DISTINCT FROM EXCLUDED."lastActivityAt";
        ELSIF TG_TABLE_NAME = 'sandbox_last_activity' THEN
          UPDATE sandbox SET "lastActivityAt" = NEW."lastActivityAt"
          WHERE id = NEW."sandboxId" AND "lastActivityAt" IS DISTINCT FROM NEW."lastActivityAt";
        END IF;
        RETURN NEW;
      END;
      $$ LANGUAGE plpgsql;
    `)

    // Recreate trigger on sandbox table (old -> new)
    await queryRunner.query(`
      CREATE TRIGGER sandbox_activity_sync_to_new
      AFTER INSERT OR UPDATE OF "lastActivityAt" ON "sandbox"
      FOR EACH ROW EXECUTE FUNCTION sync_sandbox_last_activity();
    `)

    // Recreate trigger on sandbox_last_activity table (new -> old)
    await queryRunner.query(`
      CREATE TRIGGER sandbox_activity_sync_to_old
      AFTER INSERT OR UPDATE OF "lastActivityAt" ON "sandbox_last_activity"
      FOR EACH ROW EXECUTE FUNCTION sync_sandbox_last_activity();
    `)

    // Re-sync any rows that changed during the window between backfill and trigger creation
    await queryRunner.query(`
      UPDATE "sandbox" s
      SET "lastActivityAt" = sla."lastActivityAt"
      FROM "sandbox_last_activity" sla
      WHERE s.id = sla."sandboxId"
        AND s."lastActivityAt" IS DISTINCT FROM sla."lastActivityAt"
    `)
  }
}
