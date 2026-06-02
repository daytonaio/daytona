/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1780000000000 implements MigrationInterface {
  name = 'Migration1780000000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "gpu_type" character varying`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "gpu_type" character varying`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" ADD "gpu_type" character varying`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods_archive" ADD "gpu_type" character varying`)
    await queryRunner.query(`ALTER TABLE "region_quota" ADD "allowed_gpu_types" text array`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "region_quota" DROP COLUMN "allowed_gpu_types"`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods_archive" DROP COLUMN "gpu_type"`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" DROP COLUMN "gpu_type"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "gpu_type"`)
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "gpu_type"`)
  }
}
