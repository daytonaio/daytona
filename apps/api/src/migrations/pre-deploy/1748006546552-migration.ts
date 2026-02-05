/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1748006546552 implements MigrationInterface {
  name = 'Migration1748006546552'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "max_concurrent_workspaces"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "workspace_quota"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "total_image_size"`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "total_memory_quota" SET DEFAULT '10'`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "total_disk_quota" SET DEFAULT '30'`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "max_cpu_per_workspace" SET DEFAULT '4'`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "max_memory_per_workspace" SET DEFAULT '8'`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "max_image_size" SET DEFAULT '20'`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "image_quota" SET DEFAULT '100'`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "volume_quota" SET DEFAULT '100'`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "volume_quota" SET DEFAULT '10'`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "image_quota" SET DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "max_image_size" SET DEFAULT '2'`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "max_memory_per_workspace" SET DEFAULT '4'`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "max_cpu_per_workspace" SET DEFAULT '2'`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "total_disk_quota" SET DEFAULT '100'`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "total_memory_quota" SET DEFAULT '40'`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "total_image_size" integer NOT NULL DEFAULT '5'`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "workspace_quota" integer NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "max_concurrent_workspaces" integer NOT NULL DEFAULT '10'`)
  }
}
