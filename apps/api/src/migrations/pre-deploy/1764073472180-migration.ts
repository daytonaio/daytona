/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'
import { configuration } from '../../config/configuration'

export class Migration1764073472180 implements MigrationInterface {
  name = 'Migration1764073472180'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Remove defaultRegion column from organization table
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "defaultRegion"`)

    // Remove region-specific quotas from organization table
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "total_cpu_quota"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "total_memory_quota"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "total_disk_quota"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Restore defaultRegion column to organization table
    await queryRunner.query(`ALTER TABLE "organization" ADD "defaultRegion" character varying NULL`)
    await queryRunner.query(`UPDATE "organization" SET "defaultRegion" = "defaultRegionId"`)
    await queryRunner.query(
      `ALTER TABLE "organization" ALTER COLUMN "defaultRegion" SET DEFAULT '${configuration.defaultRegion.id}'`,
    )
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "defaultRegion" SET NOT NULL`)

    // Restore region-specific quotas to organization table
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
