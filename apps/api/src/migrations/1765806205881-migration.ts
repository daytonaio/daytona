/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1765806205881 implements MigrationInterface {
  name = 'Migration1765806205881'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Create snapshot_region table
    await queryRunner.query(`
        CREATE TABLE "snapshot_region" (
          "snapshotId" uuid NOT NULL,
          "regionId" character varying NOT NULL,
          "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
          "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
          CONSTRAINT "PK_snapshot_region" PRIMARY KEY ("snapshotId", "regionId")
        )
      `)

    // Add foreign key constraints
    await queryRunner.query(`
        ALTER TABLE "snapshot_region"
        ADD CONSTRAINT "FK_snapshot_region_snapshot"
        FOREIGN KEY ("snapshotId") REFERENCES "snapshot"("id") ON DELETE CASCADE ON UPDATE NO ACTION
      `)

    await queryRunner.query(`
        ALTER TABLE "snapshot_region"
        ADD CONSTRAINT "FK_snapshot_region_region"
        FOREIGN KEY ("regionId") REFERENCES "region"("id") ON DELETE CASCADE ON UPDATE NO ACTION
      `)

    // Migrate existing snapshots: add snapshot_region entries based on organization's default region
    await queryRunner.query(`
        INSERT INTO "snapshot_region" ("snapshotId", "regionId")
        SELECT s.id, o."defaultRegionId"
        FROM "snapshot" s
        INNER JOIN "organization" o ON s."organizationId" = o.id
        WHERE o."defaultRegionId" IS NOT NULL
      `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Drop foreign key constraints
    await queryRunner.query(`ALTER TABLE "snapshot_region" DROP CONSTRAINT "FK_snapshot_region_region"`)
    await queryRunner.query(`ALTER TABLE "snapshot_region" DROP CONSTRAINT "FK_snapshot_region_snapshot"`)

    // Drop snapshot_region table
    await queryRunner.query(`DROP TABLE "snapshot_region"`)
  }
}
