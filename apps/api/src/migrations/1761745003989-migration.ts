/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1761745003989 implements MigrationInterface {
  name = 'Migration1761745003989'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Create region table
    await queryRunner.query(
      `CREATE TABLE "region" ("id" character varying NOT NULL, "name" character varying NOT NULL, "organizationId" uuid NOT NULL, "enforceQuotas" boolean NOT NULL DEFAULT true, "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "region_organizationId_name_unique" UNIQUE ("organizationId", "name"), CONSTRAINT "region_id_pk" PRIMARY KEY ("id"))`,
    )

    // Create region quota table
    await queryRunner.query(
      `CREATE TABLE "region_quota" ("organizationId" uuid NOT NULL, "region" character varying NOT NULL, "total_cpu_quota" integer NOT NULL DEFAULT '10', "total_memory_quota" integer NOT NULL DEFAULT '10', "total_disk_quota" integer NOT NULL DEFAULT '30', "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "region_quota_organizationId_region_pk" PRIMARY KEY ("organizationId", "region"))`,
    )
    await queryRunner.query(
      `ALTER TABLE "region_quota" ADD CONSTRAINT "region_quota_organizationId_fk" FOREIGN KEY ("organizationId") REFERENCES "organization"("id") ON DELETE CASCADE ON UPDATE NO ACTION`,
    )

    // For existing organizations, migrate their region-specific quotas to each region where quotas are enforced
    await queryRunner.query(`
        INSERT INTO "region_quota" ("organizationId", "region", "total_cpu_quota", "total_memory_quota", "total_disk_quota")
        SELECT 
          o."id" as "organizationId",
          r."name" as "region",
          o."total_cpu_quota",
          o."total_memory_quota",
          o."total_disk_quota"
        FROM "organization" o
        CROSS JOIN "region" r
        WHERE r."enforceQuotas" = true
      `)

    // Drop migrated region-specific quotas from organization table
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "total_cpu_quota"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "total_memory_quota"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "total_disk_quota"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" ADD "total_disk_quota" integer NOT NULL DEFAULT '30'`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "total_memory_quota" integer NOT NULL DEFAULT '10'`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "total_cpu_quota" integer NOT NULL DEFAULT '10'`)

    // For each organization, revert by taking the maximum values among all region quotas
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

    await queryRunner.query(`ALTER TABLE "region_quota" DROP CONSTRAINT "region_quota_organizationId_fk"`)
    await queryRunner.query(`DROP TABLE "region_quota"`)

    await queryRunner.query(`DROP TABLE "region"`)
  }
}
