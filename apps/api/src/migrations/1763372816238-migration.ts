/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1763372816238 implements MigrationInterface {
  name = 'Migration1763372816238'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Remove region-specific quotas from organization table
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "total_cpu_quota"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "total_memory_quota"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "total_disk_quota"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" ADD "total_disk_quota" integer NOT NULL DEFAULT '30'`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "total_memory_quota" integer NOT NULL DEFAULT '10'`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "total_cpu_quota" integer NOT NULL DEFAULT '10'`)

    // For each organization, restore region-specific quotas by taking the maximum values among all region quotas
    await queryRunner.query(`
      UPDATE "organization" o
      SET 
        "total_cpu_quota" = COALESCE(q."total_cpu_quota", 10),
        "total_memory_quota" = COALESCE(q."total_memory_quota", 10),
        "total_disk_quota" = COALESCE(q."total_disk_quota", 30)
      FROM (
        SELECT 
          "organizationId",
          MAX("total_cpu_quota") as "total_cpu_quota",
          MAX("total_memory_quota") as "total_memory_quota",
          MAX("total_disk_quota") as "total_disk_quota"
        FROM "region_quota"
        GROUP BY "organizationId"
      ) q
      WHERE o."id" = q."organizationId"
    `)
  }
}
