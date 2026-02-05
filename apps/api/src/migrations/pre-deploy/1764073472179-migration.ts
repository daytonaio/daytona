/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'
import { configuration } from '../../config/configuration'

export class Migration1764073472179 implements MigrationInterface {
  name = 'Migration1764073472179'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Create region table
    await queryRunner.query(
      `CREATE TABLE "region" ("id" character varying NOT NULL, "name" character varying NOT NULL, "organizationId" uuid, "hidden" boolean NOT NULL DEFAULT false, "enforceQuotas" boolean NOT NULL DEFAULT true, "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "region_id_pk" PRIMARY KEY ("id"))`,
    )

    // Add unique constraints for region name
    await queryRunner.query(
      `CREATE UNIQUE INDEX "region_organizationId_null_name_unique" ON "region" ("name") WHERE "organizationId" IS NULL`,
    )
    await queryRunner.query(
      `CREATE UNIQUE INDEX "region_organizationId_name_unique" ON "region" ("organizationId", "name") WHERE "organizationId" IS NOT NULL`,
    )

    // Expand organization table with defaultRegionId column (make it nullable)
    await queryRunner.query(`ALTER TABLE "organization" ADD "defaultRegionId" character varying NULL`)
    await queryRunner.query(`UPDATE "organization" SET "defaultRegionId" = "defaultRegion"`)

    // Add default value for required defaultRegion column before dropping it in the contract migration
    await queryRunner.query(
      `ALTER TABLE "organization" ALTER COLUMN "defaultRegion" SET DEFAULT '${configuration.defaultRegion.id}'`,
    )

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
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Drop region table
    await queryRunner.query(`DROP TABLE "region"`)

    // Drop defaultRegionId column from organization table
    await queryRunner.dropColumn('organization', 'defaultRegionId')

    // Drop region_quota table
    await queryRunner.query(`DROP TABLE "region_quota"`)
  }
}
