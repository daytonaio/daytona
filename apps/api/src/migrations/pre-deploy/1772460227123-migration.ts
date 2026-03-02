/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1772460227123 implements MigrationInterface {
  name = 'Migration1772460227123'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Create new table
    await queryRunner.query(
      `CREATE TABLE "sandbox_last_activity" ("sandboxId" character varying NOT NULL, "lastActivityAt" TIMESTAMP WITH TIME ZONE, CONSTRAINT "sandbox_last_activity_sandboxId_pk" PRIMARY KEY ("sandboxId"))`,
    )
    await queryRunner.query(
      `ALTER TABLE "sandbox_last_activity" ADD CONSTRAINT "sandbox_last_activity_sandboxId_fk" FOREIGN KEY ("sandboxId") REFERENCES "sandbox"("id") ON DELETE CASCADE ON UPDATE NO ACTION`,
    )

    // Copy existing data from sandbox.lastActivityAt to new table
    await queryRunner.query(`
      INSERT INTO "sandbox_last_activity" ("sandboxId", "lastActivityAt")
      SELECT "id", COALESCE("lastActivityAt", "createdAt", NOW())
      FROM "sandbox"
      WHERE "state" != 'destroyed'
    `)

    // Create sync function for dual-write
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

    // Create trigger on sandbox table (old -> new)
    await queryRunner.query(`
      CREATE TRIGGER sandbox_activity_sync_to_new
      AFTER INSERT OR UPDATE OF "lastActivityAt" ON "sandbox"
      FOR EACH ROW EXECUTE FUNCTION sync_sandbox_last_activity();
    `)

    // Create trigger on sandbox_last_activity table (new -> old)
    await queryRunner.query(`
      CREATE TRIGGER sandbox_activity_sync_to_old
      AFTER INSERT OR UPDATE OF "lastActivityAt" ON "sandbox_last_activity"
      FOR EACH ROW EXECUTE FUNCTION sync_sandbox_last_activity();
    `)

    // Re-sync any rows that changed during the window between initial copy and trigger creation
    await queryRunner.query(`
      UPDATE "sandbox_last_activity" sla
      SET "lastActivityAt" = s."lastActivityAt"
      FROM "sandbox" s
      WHERE sla."sandboxId" = s.id
        AND sla."lastActivityAt" IS DISTINCT FROM s."lastActivityAt"
    `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Drop dual-write triggers and function
    await queryRunner.query(`DROP TRIGGER IF EXISTS sandbox_activity_sync_to_old ON "sandbox_last_activity"`)
    await queryRunner.query(`DROP TRIGGER IF EXISTS sandbox_activity_sync_to_new ON "sandbox"`)
    await queryRunner.query(`DROP FUNCTION IF EXISTS sync_sandbox_last_activity()`)

    // Drop table
    await queryRunner.query(`ALTER TABLE "sandbox_last_activity" DROP CONSTRAINT "sandbox_last_activity_sandboxId_fk"`)
    await queryRunner.query(`DROP TABLE "sandbox_last_activity"`)
  }
}
