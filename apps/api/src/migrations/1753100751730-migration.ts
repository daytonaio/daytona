/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1753100751730 implements MigrationInterface {
  name = 'Migration1753100751730'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.renameColumn('runner', 'memory', 'memoryGiB')
    await queryRunner.renameColumn('runner', 'disk', 'diskGiB')
    await queryRunner.query(
      `ALTER TABLE "runner" ADD "currentCpuUsagePercentage" double precision NOT NULL DEFAULT '0'`,
    )
    await queryRunner.query(
      `ALTER TABLE "runner" ADD "currentMemoryUsagePercentage" double precision NOT NULL DEFAULT '0'`,
    )
    await queryRunner.query(
      `ALTER TABLE "runner" ADD "currentDiskUsagePercentage" double precision NOT NULL DEFAULT '0'`,
    )
    await queryRunner.query(`ALTER TABLE "runner" ADD "currentAllocatedCpu" integer NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "currentAllocatedMemoryGiB" integer NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "currentAllocatedDiskGiB" integer NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "currentSnapshotCount" integer NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "availabilityScore" integer NOT NULL DEFAULT '0'`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "availabilityScore"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "currentSnapshotCount"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "currentAllocatedDiskGiB"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "currentAllocatedMemoryGiB"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "currentAllocatedCpu"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "currentDiskUsagePercentage"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "currentMemoryUsagePercentage"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "currentCpuUsagePercentage"`)
    await queryRunner.renameColumn('runner', 'diskGiB', 'disk')
    await queryRunner.renameColumn('runner', 'memoryGiB', 'memory')
  }
}
