/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1765544432484 implements MigrationInterface {
  name = 'Migration1765544432484'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "domain" DROP NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "apiUrl" DROP NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "proxyUrl" DROP NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "cpu" SET DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "memoryGiB" SET DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "diskGiB" SET DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "gpu" DROP NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "gpuType" DROP NOT NULL`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "gpuType" SET NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "gpu" SET NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "diskGiB" DROP DEFAULT`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "memoryGiB" DROP DEFAULT`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "cpu" DROP DEFAULT`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "proxyUrl" SET NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "apiUrl" SET NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "domain" SET NOT NULL`)
  }
}
