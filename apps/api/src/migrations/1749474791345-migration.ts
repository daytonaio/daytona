/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1749474791345 implements MigrationInterface {
  name = 'Migration1749474791345'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // For sandbox_state_enum
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

    // For snapshot_state_enum
    await queryRunner.query(`ALTER TYPE "public"."snapshot_state_enum" RENAME TO "snapshot_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."snapshot_state_enum" AS ENUM('build_pending', 'building', 'pending', 'pulling', 'pending_validation', 'validating', 'active', 'error', 'build_failed', 'removing')`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "snapshot" ALTER COLUMN "state" TYPE "public"."snapshot_state_enum" USING "state"::"text"::"public"."snapshot_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" SET DEFAULT 'pending'`)
    await queryRunner.query(`DROP TYPE "public"."snapshot_state_enum_old"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // For snapshot_state_enum - recreate without build_failed
    await queryRunner.query(`UPDATE "snapshot" SET "state" = 'error' WHERE "state" = 'build_failed'`)

    await queryRunner.query(`ALTER TYPE "public"."snapshot_state_enum" RENAME TO "snapshot_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."snapshot_state_enum" AS ENUM('build_pending', 'building', 'pending', 'pulling', 'pending_validation', 'validating', 'active', 'error', 'removing')`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "snapshot" ALTER COLUMN "state" TYPE "public"."snapshot_state_enum" USING "state"::"text"::"public"."snapshot_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" SET DEFAULT 'pending'`)
    await queryRunner.query(`DROP TYPE "public"."snapshot_state_enum_old"`)

    // For sandbox_state_enum - recreate without build_failed
    await queryRunner.query(`UPDATE "sandbox" SET "state" = 'error' WHERE "state" = 'build_failed'`)

    await queryRunner.query(`ALTER TYPE "public"."sandbox_state_enum" RENAME TO "sandbox_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."sandbox_state_enum" AS ENUM('creating', 'restoring', 'destroyed', 'destroying', 'started', 'stopped', 'starting', 'stopping', 'error', 'pending_build', 'building_snapshot', 'unknown', 'pulling_snapshot', 'archiving', 'archived')`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "sandbox" ALTER COLUMN "state" TYPE "public"."sandbox_state_enum" USING "state"::"text"::"public"."sandbox_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "state" SET DEFAULT 'unknown'`)
    await queryRunner.query(`DROP TYPE "public"."sandbox_state_enum_old"`)
  }
}
