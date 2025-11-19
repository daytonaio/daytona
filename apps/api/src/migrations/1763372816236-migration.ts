/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'
import { configuration } from '../config/configuration'

export class Migration1763372816236 implements MigrationInterface {
  name = 'Migration1763372816236'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Create region_quota table
    await queryRunner.query(
      `CREATE TABLE "region_quota" ("organizationId" uuid NOT NULL, "regionId" character varying NOT NULL, "total_cpu_quota" integer NOT NULL DEFAULT '10', "total_memory_quota" integer NOT NULL DEFAULT '10', "total_disk_quota" integer NOT NULL DEFAULT '30', "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "region_quota_organizationId_regionId_pk" PRIMARY KEY ("organizationId", "regionId"))`,
    )
    await queryRunner.query(
      `ALTER TABLE "region_quota" ADD CONSTRAINT "region_quota_organizationId_fk" FOREIGN KEY ("organizationId") REFERENCES "organization"("id") ON DELETE CASCADE ON UPDATE NO ACTION`,
    )

    // For existing organizations, migrate their region-specific quotas to their default region
    await queryRunner.query(`
        INSERT INTO "region_quota" ("organizationId", "regionId", "total_cpu_quota", "total_memory_quota", "total_disk_quota")
        SELECT 
          o."id" as "organizationId",
          o."defaultRegionId" as "regionId",
          o."total_cpu_quota",
          o."total_memory_quota",
          o."total_disk_quota"
        FROM "organization" o
      `)

    // For all other regions where quotas are enforced, assign default quotas from configuration
    await queryRunner.query(`
        INSERT INTO "region_quota" ("organizationId", "regionId", "total_cpu_quota", "total_memory_quota", "total_disk_quota")
        SELECT 
          o."id" as "organizationId",
          r."id" as "regionId",
          ${configuration.defaultOrganizationQuota.totalCpuQuota} as "total_cpu_quota",
          ${configuration.defaultOrganizationQuota.totalMemoryQuota} as "total_memory_quota", 
          ${configuration.defaultOrganizationQuota.totalDiskQuota} as "total_disk_quota"
        FROM "organization" o
        CROSS JOIN "region" r
        WHERE r."enforceQuotas" = true 
          AND r."id" != o."defaultRegionId"
      `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP TABLE "region_quota"`)
  }
}
