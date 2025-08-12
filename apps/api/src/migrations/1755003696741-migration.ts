/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1755003696741 implements MigrationInterface {
  name = 'Migration1755003696741'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TYPE "public"."sandbox_state_enum" RENAME TO "sandbox_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."sandbox_state_enum" AS ENUM('creating', 'restoring', 'destroyed', 'destroying', 'started', 'stopped', 'starting', 'stopping', 'error', 'build_failed', 'pending_build', 'building_snapshot', 'unknown', 'pulling_snapshot', 'archived', 'pending_archive')`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "sandbox" ALTER COLUMN "state" TYPE "public"."sandbox_state_enum" USING "state"::"text"::"public"."sandbox_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "state" SET DEFAULT 'unknown'`)
    await queryRunner.query(`DROP TYPE "public"."sandbox_state_enum_old"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TYPE "public"."sandbox_state_enum_old" AS ENUM('archived', 'archiving', 'build_failed', 'building_snapshot', 'creating', 'destroyed', 'destroying', 'error', 'pending_build', 'pulling_snapshot', 'restoring', 'started', 'starting', 'stopped', 'stopping', 'unknown')`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "sandbox" ALTER COLUMN "state" TYPE "public"."sandbox_state_enum_old" USING "state"::"text"::"public"."sandbox_state_enum_old"`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "state" SET DEFAULT 'unknown'`)
    await queryRunner.query(`DROP TYPE "public"."sandbox_state_enum"`)
    await queryRunner.query(`ALTER TYPE "public"."sandbox_state_enum_old" RENAME TO "sandbox_state_enum"`)
  }
}
