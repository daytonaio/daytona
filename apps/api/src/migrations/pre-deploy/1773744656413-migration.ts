/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1773744656413 implements MigrationInterface {
  name = 'Migration1773744656413'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "currentAllocatedCpu" TYPE double precision`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "currentAllocatedMemoryGiB" TYPE double precision`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "currentAllocatedDiskGiB" TYPE double precision`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `ALTER TABLE "runner" ALTER COLUMN "currentAllocatedDiskGiB" TYPE integer USING ROUND("currentAllocatedDiskGiB")::integer`,
    )
    await queryRunner.query(
      `ALTER TABLE "runner" ALTER COLUMN "currentAllocatedMemoryGiB" TYPE integer USING ROUND("currentAllocatedMemoryGiB")::integer`,
    )
    await queryRunner.query(
      `ALTER TABLE "runner" ALTER COLUMN "currentAllocatedCpu" TYPE integer USING ROUND("currentAllocatedCpu")::integer`,
    )
  }
}
