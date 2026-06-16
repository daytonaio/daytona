/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1779192781000 implements MigrationInterface {
  name = 'Migration1779192781000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
      CREATE TABLE "sandbox_lifecycle" (
        "sandboxId" character varying NOT NULL,
        "lifecyclePhase" text NOT NULL,
        "organizationId" uuid NOT NULL,
        "state" sandbox_state_enum NOT NULL DEFAULT 'unknown',
        "desiredState" sandbox_desiredstate_enum NOT NULL DEFAULT 'started',
        "pending" boolean NOT NULL DEFAULT false,
        "errorReason" character varying,
        "recoverable" boolean NOT NULL DEFAULT false,
        "daemonVersion" character varying,
        "runnerId" uuid,
        "prevRunnerId" uuid,
        "backupState" sandbox_backupstate_enum NOT NULL DEFAULT 'None',
        "lastBackupAt" timestamp with time zone,
        "backupSnapshot" character varying,
        "backupRegistryId" character varying,
        "backupErrorReason" text,
        "existingBackupSnapshots" jsonb NOT NULL DEFAULT '[]'::jsonb,
        "updatedAt" timestamp with time zone NOT NULL DEFAULT now(),
        CONSTRAINT "sandbox_lifecycle_pk" PRIMARY KEY ("sandboxId", "lifecyclePhase"),
        CONSTRAINT "sandbox_lifecycle_phase_check" CHECK ("lifecyclePhase" IN ('active', 'terminal'))
      ) PARTITION BY LIST ("lifecyclePhase")
    `)

    await queryRunner.query(`
      CREATE TABLE "sandbox_lifecycle_active"
        PARTITION OF "sandbox_lifecycle"
        FOR VALUES IN ('active')
    `)

    await queryRunner.query(`
      CREATE TABLE "sandbox_lifecycle_terminal"
        PARTITION OF "sandbox_lifecycle"
        FOR VALUES IN ('terminal')
    `)

    await queryRunner.query(`
      ALTER TABLE "sandbox_lifecycle"
        ADD CONSTRAINT "sandbox_lifecycle_sandboxId_fk"
        FOREIGN KEY ("sandboxId") REFERENCES "sandbox"("id")
        ON DELETE CASCADE ON UPDATE NO ACTION
    `)

    await queryRunner.query(`CREATE INDEX "sandbox_lifecycle_active_state_idx" ON "sandbox_lifecycle_active" ("state")`)
    await queryRunner.query(
      `CREATE INDEX "sandbox_lifecycle_active_runner_state_idx" ON "sandbox_lifecycle_active" ("runnerId", "state")`,
    )
    await queryRunner.query(`
      CREATE INDEX "sandbox_lifecycle_active_runner_state_desired_idx"
        ON "sandbox_lifecycle_active" ("runnerId", "state", "desiredState")
        WHERE "pending" = false
    `)
    await queryRunner.query(
      `CREATE INDEX "sandbox_lifecycle_active_backupstate_idx" ON "sandbox_lifecycle_active" ("backupState")`,
    )
    await queryRunner.query(`
      CREATE INDEX "sandbox_lifecycle_active_pending_idx"
        ON "sandbox_lifecycle_active" ("sandboxId")
        WHERE "pending" = true
    `)
    await queryRunner.query(`
      CREATE INDEX "sandbox_lifecycle_active_recoverable_idx"
        ON "sandbox_lifecycle_active" ("sandboxId")
        WHERE "recoverable" = true
    `)
    await queryRunner.query(
      `CREATE INDEX "sandbox_lifecycle_active_orgid_idx" ON "sandbox_lifecycle_active" ("organizationId")`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX IF EXISTS "sandbox_lifecycle_active_orgid_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "sandbox_lifecycle_active_recoverable_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "sandbox_lifecycle_active_pending_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "sandbox_lifecycle_active_backupstate_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "sandbox_lifecycle_active_runner_state_desired_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "sandbox_lifecycle_active_runner_state_idx"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "sandbox_lifecycle_active_state_idx"`)
    await queryRunner.query(
      `ALTER TABLE "sandbox_lifecycle" DROP CONSTRAINT IF EXISTS "sandbox_lifecycle_sandboxId_fk"`,
    )
    await queryRunner.query(`DROP TABLE IF EXISTS "sandbox_lifecycle_terminal"`)
    await queryRunner.query(`DROP TABLE IF EXISTS "sandbox_lifecycle_active"`)
    await queryRunner.query(`DROP TABLE IF EXISTS "sandbox_lifecycle"`)
  }
}
