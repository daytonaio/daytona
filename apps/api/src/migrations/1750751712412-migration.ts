/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1750751712412 implements MigrationInterface {
  name = 'Migration1750751712412'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TYPE "public"."snapshot_state_enum" RENAME TO "snapshot_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."snapshot_state_enum" AS ENUM('build_pending', 'building', 'pending', 'pulling', 'pending_validation', 'validating', 'active', 'inactive', 'error', 'build_failed', 'removing')`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "snapshot" ALTER COLUMN "state" TYPE "public"."snapshot_state_enum" USING "state"::"text"::"public"."snapshot_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" SET DEFAULT 'pending'`)
    await queryRunner.query(`DROP TYPE "public"."snapshot_state_enum_old"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TYPE "public"."snapshot_state_enum_old" AS ENUM('active', 'build_failed', 'build_pending', 'building', 'error', 'pending', 'pending_validation', 'pulling', 'removing', 'validating')`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "snapshot" ALTER COLUMN "state" TYPE "public"."snapshot_state_enum_old" USING "state"::"text"::"public"."snapshot_state_enum_old"`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" SET DEFAULT 'pending'`)
    await queryRunner.query(`DROP TYPE "public"."snapshot_state_enum"`)
    await queryRunner.query(`ALTER TYPE "public"."snapshot_state_enum_old" RENAME TO "snapshot_state_enum"`)
  }
}
