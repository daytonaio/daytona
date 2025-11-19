/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1763372816235 implements MigrationInterface {
  name = 'Migration1763372816235'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Create region table
    await queryRunner.query(
      `CREATE TABLE "region" ("id" character varying NOT NULL, "name" character varying NOT NULL, "organizationId" uuid, "enforceQuotas" boolean NOT NULL DEFAULT true, "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "region_id_pk" PRIMARY KEY ("id"))`,
    )

    await queryRunner.query(
      `CREATE UNIQUE INDEX "region_organizationId_null_name_unique" ON "region" ("name") WHERE "organizationId" IS NULL`,
    )
    await queryRunner.query(
      `CREATE UNIQUE INDEX "region_organizationId_name_unique" ON "region" ("organizationId", "name") WHERE "organizationId" IS NOT NULL`,
    )

    // Expand organization table with defaultRegionId column
    await queryRunner.query(`ALTER TABLE "organization" ADD "defaultRegionId" character varying NULL`)
    await queryRunner.query(`UPDATE "organization" SET "defaultRegionId" = "defaultRegion"`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "defaultRegionId" SET NOT NULL`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Drop region table
    await queryRunner.query(`DROP TABLE "region"`)

    // Drop defaultRegionId column from organization table
    await queryRunner.dropColumn('organization', 'defaultRegionId')
  }
}
