/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1769516172576 implements MigrationInterface {
  name = 'Migration1769516172576'

  public async up(queryRunner: QueryRunner): Promise<void> {
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

    // For sandbox_desired_state_enum - remove 'resized' value using hotswap approach
    await queryRunner.query(
      `ALTER TYPE "public"."sandbox_desired_state_enum" RENAME TO "sandbox_desired_state_enum_old"`,
    )
    await queryRunner.query(
      `CREATE TYPE "public"."sandbox_desired_state_enum" AS ENUM('destroyed', 'started', 'stopped', 'archived')`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "desiredState" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "sandbox" ALTER COLUMN "desiredState" TYPE "public"."sandbox_desired_state_enum" USING "desiredState"::"text"::"public"."sandbox_desired_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "desiredState" SET DEFAULT 'started'`)
    await queryRunner.query(`DROP TYPE "public"."sandbox_desired_state_enum_old"`)

    // Drop the resizing boolean column (no longer needed since we use state)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "resizing"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Re-add the resizing column
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "resizing" boolean NOT NULL DEFAULT false`)

    // For sandbox_desired_state_enum - re-add 'resized' value
    await queryRunner.query(
      `ALTER TYPE "public"."sandbox_desired_state_enum" RENAME TO "sandbox_desired_state_enum_old"`,
    )
    await queryRunner.query(
      `CREATE TYPE "public"."sandbox_desired_state_enum" AS ENUM('destroyed', 'started', 'stopped', 'resized', 'archived')`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "desiredState" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "sandbox" ALTER COLUMN "desiredState" TYPE "public"."sandbox_desired_state_enum" USING "desiredState"::"text"::"public"."sandbox_desired_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "desiredState" SET DEFAULT 'started'`)
    await queryRunner.query(`DROP TYPE "public"."sandbox_desired_state_enum_old"`)

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
  }
}
