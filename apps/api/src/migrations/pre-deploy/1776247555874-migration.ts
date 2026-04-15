/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1776247555874 implements MigrationInterface {
  name = 'Migration1776247555874'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "region_quota" ADD "max_cpu_per_sandbox" integer`)
    await queryRunner.query(`ALTER TABLE "region_quota" ADD "max_memory_per_sandbox" integer`)
    await queryRunner.query(`ALTER TABLE "region_quota" ADD "max_disk_per_sandbox" integer`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "region_quota" DROP COLUMN "max_disk_per_sandbox"`)
    await queryRunner.query(`ALTER TABLE "region_quota" DROP COLUMN "max_memory_per_sandbox"`)
    await queryRunner.query(`ALTER TABLE "region_quota" DROP COLUMN "max_cpu_per_sandbox"`)
  }
}
