/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1780531200000 implements MigrationInterface {
  name = 'Migration1780531200000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "region_quota" ALTER COLUMN "max_cpu_per_gpu_sandbox" SET DEFAULT 16`)
    await queryRunner.query(`ALTER TABLE "region_quota" ALTER COLUMN "max_memory_per_gpu_sandbox" SET DEFAULT 192`)
    await queryRunner.query(`ALTER TABLE "region_quota" ALTER COLUMN "max_disk_per_gpu_sandbox" SET DEFAULT 512`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "region_quota" ALTER COLUMN "max_disk_per_gpu_sandbox" DROP DEFAULT`)
    await queryRunner.query(`ALTER TABLE "region_quota" ALTER COLUMN "max_memory_per_gpu_sandbox" DROP DEFAULT`)
    await queryRunner.query(`ALTER TABLE "region_quota" ALTER COLUMN "max_cpu_per_gpu_sandbox" DROP DEFAULT`)
  }
}
