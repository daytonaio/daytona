/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1769516172576 implements MigrationInterface {
  name = 'Migration1769516172576'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // For sandbox_state_enum - add 'resizing' value
    await queryRunner.query(`ALTER TYPE "public"."sandbox_state_enum" ADD VALUE IF NOT EXISTS 'resizing'`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Drop any index with explicit desiredState enum type cast in WHERE clause (required for enum swap)

    // For sandbox_state_enum - remove 'resizing' value
    await queryRunner.query(`UPDATE "sandbox" SET "state" = 'stopped' WHERE "state" = 'resizing'`)
    await queryRunner.query(`ALTER TYPE "public"."sandbox_state_enum" RENAME TO "sandbox_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."sandbox_state_enum" AS ENUM('creating', 'restoring', 'destroyed', 'destroying', 'started', 'stopped', 'starting', 'stopping', 'error', 'build_failed', 'pending_build', 'building_snapshot', 'unknown', 'pulling_snapshot', 'archiving', 'archived')`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "sandbox" ALTER COLUMN "state" TYPE "public"."sandbox_state_enum" USING "state"::"text"::"public"."sandbox_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "state" SET DEFAULT 'unknown'`)
    await queryRunner.query(`DROP TYPE "public"."sandbox_state_enum_old"`)

    // Recreate the indices that were dropped
  }
}
