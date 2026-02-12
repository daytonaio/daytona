/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1770300000000 implements MigrationInterface {
  name = 'Migration1770300000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Add 'snapshotting' to sandbox_state_enum
    await queryRunner.query(`ALTER TYPE "public"."sandbox_state_enum" ADD VALUE IF NOT EXISTS 'snapshotting'`)

    // Create checkpoint_state_enum
    await queryRunner.query(
      `CREATE TYPE "public"."checkpoint_state_enum" AS ENUM('creating', 'active', 'error', 'removing')`,
    )

    // Create checkpoint_runner_state_enum
    await queryRunner.query(`CREATE TYPE "public"."checkpoint_runner_state_enum" AS ENUM('pulling', 'ready', 'error')`)

    // Create checkpoint table
    await queryRunner.query(
      `CREATE TABLE "checkpoint" (
        "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
        "sandboxId" uuid NOT NULL,
        "organizationId" uuid NOT NULL,
        "name" character varying NOT NULL,
        "ref" character varying,
        "state" "public"."checkpoint_state_enum" NOT NULL DEFAULT 'creating',
        "errorReason" character varying,
        "sizeBytes" bigint,
        "hash" character varying,
        "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
        "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
        CONSTRAINT "UQ_checkpoint_sandbox_name" UNIQUE ("sandboxId", "name"),
        CONSTRAINT "PK_checkpoint" PRIMARY KEY ("id")
      )`,
    )

    // Create indexes for checkpoint
    await queryRunner.query(`CREATE INDEX "checkpoint_sandboxid_idx" ON "checkpoint" ("sandboxId")`)
    await queryRunner.query(`CREATE INDEX "checkpoint_organizationid_idx" ON "checkpoint" ("organizationId")`)
    await queryRunner.query(`CREATE INDEX "checkpoint_state_idx" ON "checkpoint" ("state")`)

    // Add foreign key from checkpoint to sandbox (CASCADE DELETE)
    await queryRunner.query(
      `ALTER TABLE "checkpoint" ADD CONSTRAINT "FK_checkpoint_sandbox" FOREIGN KEY ("sandboxId") REFERENCES "sandbox"("id") ON DELETE CASCADE ON UPDATE NO ACTION`,
    )

    // Create checkpoint_runner table
    await queryRunner.query(
      `CREATE TABLE "checkpoint_runner" (
        "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
        "checkpointId" uuid NOT NULL,
        "runnerId" character varying NOT NULL,
        "state" "public"."checkpoint_runner_state_enum" NOT NULL DEFAULT 'pulling',
        "errorReason" character varying,
        "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
        "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
        CONSTRAINT "PK_checkpoint_runner" PRIMARY KEY ("id")
      )`,
    )

    // Create indexes for checkpoint_runner
    await queryRunner.query(`CREATE INDEX "checkpoint_runner_checkpointid_idx" ON "checkpoint_runner" ("checkpointId")`)
    await queryRunner.query(`CREATE INDEX "checkpoint_runner_runnerid_idx" ON "checkpoint_runner" ("runnerId")`)
    await queryRunner.query(`CREATE INDEX "checkpoint_runner_state_idx" ON "checkpoint_runner" ("state")`)

    // Add foreign key from checkpoint_runner to checkpoint (CASCADE DELETE)
    await queryRunner.query(
      `ALTER TABLE "checkpoint_runner" ADD CONSTRAINT "FK_checkpoint_runner_checkpoint" FOREIGN KEY ("checkpointId") REFERENCES "checkpoint"("id") ON DELETE CASCADE ON UPDATE NO ACTION`,
    )

    // Add originSandboxId and originCheckpointId to snapshot table
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "originSandboxId" uuid`)
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "originCheckpointId" uuid`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Remove columns from snapshot
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "originCheckpointId"`)
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "originSandboxId"`)

    // Drop checkpoint_runner table and its constraints
    await queryRunner.query(`ALTER TABLE "checkpoint_runner" DROP CONSTRAINT "FK_checkpoint_runner_checkpoint"`)
    await queryRunner.query(`DROP INDEX "checkpoint_runner_state_idx"`)
    await queryRunner.query(`DROP INDEX "checkpoint_runner_runnerid_idx"`)
    await queryRunner.query(`DROP INDEX "checkpoint_runner_checkpointid_idx"`)
    await queryRunner.query(`DROP TABLE "checkpoint_runner"`)

    // Drop checkpoint table and its constraints
    await queryRunner.query(`ALTER TABLE "checkpoint" DROP CONSTRAINT "FK_checkpoint_sandbox"`)
    await queryRunner.query(`DROP INDEX "checkpoint_state_idx"`)
    await queryRunner.query(`DROP INDEX "checkpoint_organizationid_idx"`)
    await queryRunner.query(`DROP INDEX "checkpoint_sandboxid_idx"`)
    await queryRunner.query(`DROP TABLE "checkpoint"`)

    // Drop enums
    await queryRunner.query(`DROP TYPE "public"."checkpoint_runner_state_enum"`)
    await queryRunner.query(`DROP TYPE "public"."checkpoint_state_enum"`)

    // Remove 'snapshotting' from sandbox_state_enum
    // Note: PostgreSQL doesn't support removing enum values directly
    // Need to recreate the type
    await queryRunner.query(`UPDATE "sandbox" SET "state" = 'stopped' WHERE "state" = 'snapshotting'`)
    await queryRunner.query(`ALTER TYPE "public"."sandbox_state_enum" RENAME TO "sandbox_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."sandbox_state_enum" AS ENUM('creating', 'restoring', 'destroyed', 'destroying', 'started', 'stopped', 'starting', 'stopping', 'error', 'build_failed', 'pending_build', 'building_snapshot', 'unknown', 'pulling_snapshot', 'archived', 'archiving', 'resizing')`,
    )
    await queryRunner.query(
      `ALTER TABLE "sandbox" ALTER COLUMN "state" TYPE "public"."sandbox_state_enum" USING "state"::"text"::"public"."sandbox_state_enum"`,
    )
    await queryRunner.query(`DROP TYPE "public"."sandbox_state_enum_old"`)
  }
}
