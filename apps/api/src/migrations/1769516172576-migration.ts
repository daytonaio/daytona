/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1769516172576 implements MigrationInterface {
  name = 'Migration1769516172576'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Drop index with explicit enum type cast in WHERE clause (required for enum swap)
    await queryRunner.query(`DROP INDEX IF EXISTS "public"."sandbox_active_only_idx"`)

    // For sandbox_state_enum - add 'resizing' value using hotswap approach
    await queryRunner.query(`ALTER TYPE "public"."sandbox_state_enum" RENAME TO "sandbox_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."sandbox_state_enum" AS ENUM('creating', 'restoring', 'destroyed', 'destroying', 'started', 'stopped', 'starting', 'stopping', 'error', 'build_failed', 'pending_build', 'building_snapshot', 'unknown', 'pulling_snapshot', 'archiving', 'archived', 'resizing')`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "sandbox" ALTER COLUMN "state" TYPE "public"."sandbox_state_enum" USING "state"::"text"::"public"."sandbox_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "state" SET DEFAULT 'unknown'`)
    await queryRunner.query(`DROP TYPE "public"."sandbox_state_enum_old"`)

    // For sandbox_desiredstate_enum - remove 'resized' value using hotswap approach
    await queryRunner.query(`ALTER TYPE "public"."sandbox_desiredstate_enum" RENAME TO "sandbox_desiredstate_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."sandbox_desiredstate_enum" AS ENUM('destroyed', 'started', 'stopped', 'archived')`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "desiredState" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "sandbox" ALTER COLUMN "desiredState" TYPE "public"."sandbox_desiredstate_enum" USING "desiredState"::"text"::"public"."sandbox_desiredstate_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "desiredState" SET DEFAULT 'started'`)
    await queryRunner.query(`DROP TYPE "public"."sandbox_desiredstate_enum_old"`)

    // Recreate the index that was dropped
    await queryRunner.query(
      `CREATE INDEX "sandbox_active_only_idx" ON "sandbox" ("id") WHERE "state" <> ALL (ARRAY['destroyed'::sandbox_state_enum, 'archived'::sandbox_state_enum])`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Drop index with explicit enum type cast in WHERE clause (required for enum swap)
    await queryRunner.query(`DROP INDEX IF EXISTS "public"."sandbox_active_only_idx"`)

    // For sandbox_desiredstate_enum - re-add 'resized' value
    await queryRunner.query(`ALTER TYPE "public"."sandbox_desiredstate_enum" RENAME TO "sandbox_desiredstate_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."sandbox_desiredstate_enum" AS ENUM('destroyed', 'started', 'stopped', 'resized', 'archived')`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "desiredState" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "sandbox" ALTER COLUMN "desiredState" TYPE "public"."sandbox_desiredstate_enum" USING "desiredState"::"text"::"public"."sandbox_desiredstate_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "desiredState" SET DEFAULT 'started'`)
    await queryRunner.query(`DROP TYPE "public"."sandbox_desiredstate_enum_old"`)

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

    // Recreate the index that was dropped
    await queryRunner.query(
      `CREATE INDEX "sandbox_active_only_idx" ON "sandbox" ("id") WHERE "state" <> ALL (ARRAY['destroyed'::sandbox_state_enum, 'archived'::sandbox_state_enum])`,
    )
  }
}
