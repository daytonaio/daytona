/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1781262090000 implements MigrationInterface {
  name = 'Migration1781262090000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TYPE "public"."snapshot_state_enum" ADD VALUE IF NOT EXISTS 'snapshotting'`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Move any in-flight snapshot-from-sandbox entries to a state the old enum knows about
    await queryRunner.query(`UPDATE "snapshot" SET "state" = 'error' WHERE "state" = 'snapshotting'`)

    await queryRunner.query(`ALTER TYPE "public"."snapshot_state_enum" RENAME TO "snapshot_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."snapshot_state_enum" AS ENUM('pending', 'pulling', 'active', 'inactive', 'building', 'error', 'build_failed', 'removing')`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "snapshot" ALTER COLUMN "state" TYPE "public"."snapshot_state_enum" USING "state"::"text"::"public"."snapshot_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" SET DEFAULT 'pending'`)
    await queryRunner.query(`DROP TYPE "public"."snapshot_state_enum_old"`)
  }
}
