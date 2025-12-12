/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1765538824750 implements MigrationInterface {
  name = 'Migration1765538824750'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "domain" DROP NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" DROP CONSTRAINT "UQ_330d74ac3d0e349b4c73c62ad6d"`)
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
    await queryRunner.query(`ALTER TABLE "runner" ADD CONSTRAINT "UQ_330d74ac3d0e349b4c73c62ad6d" UNIQUE ("domain")`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "domain" SET NOT NULL`)
  }
}
