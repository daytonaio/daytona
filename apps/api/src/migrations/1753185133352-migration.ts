/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1753185133352 implements MigrationInterface {
  name = 'Migration1753185133352'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.renameColumn('snapshot', 'internalName', 'ref')
    await queryRunner.renameColumn('snapshot', 'buildRunnerId', 'initialRunnerId')
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "skipValidation" boolean NOT NULL DEFAULT false`)

    // Update snapshot states
    await queryRunner.query(`ALTER TYPE "public"."snapshot_state_enum" RENAME TO "snapshot_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."snapshot_state_enum" AS ENUM('pending', 'pulling', 'pending_validation', 'validating', 'active', 'inactive', 'building', 'warming_up', 'error', 'build_failed', 'removing')`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "snapshot" ALTER COLUMN "state" TYPE "public"."snapshot_state_enum" USING "state"::"text"::"public"."snapshot_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" SET DEFAULT 'pending'`)
    await queryRunner.query(`DROP TYPE "public"."snapshot_state_enum_old"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Revert snapshot states
    await queryRunner.query(`ALTER TYPE "public"."snapshot_state_enum" RENAME TO "snapshot_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."snapshot_state_enum" AS ENUM('pending', 'pulling', 'pending_validation', 'validating', 'active', 'inactive', 'building', 'warming_up', 'error', 'build_failed', 'removing')`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "snapshot" ALTER COLUMN "state" TYPE "public"."snapshot_state_enum" USING "state"::"text"::"public"."snapshot_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" SET DEFAULT 'pending'`)
    await queryRunner.query(`DROP TYPE "public"."snapshot_state_enum_old"`)

    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "skipValidation"`)
    await queryRunner.renameColumn('snapshot', 'initialRunnerId', 'buildRunnerId')
    await queryRunner.renameColumn('snapshot', 'ref', 'internalName')
  }
}
