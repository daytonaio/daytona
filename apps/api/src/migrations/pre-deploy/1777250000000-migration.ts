/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1777250000000 implements MigrationInterface {
  name = 'Migration1777250000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
      CREATE TABLE "volume_disk" (
        "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
        "volumeId" uuid NOT NULL,
        "subpath" character varying,
        "archilDiskId" character varying NOT NULL,
        "archilRegion" character varying NOT NULL,
        "archilMountTokenEnc" character varying NOT NULL,
        "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
        CONSTRAINT "PK_volume_disk" PRIMARY KEY ("id"),
        CONSTRAINT "UQ_volume_disk_volume_subpath" UNIQUE ("volumeId", "subpath")
      )
    `)

    await queryRunner.query(`CREATE INDEX "idx_volume_disk_volume_id" ON "volume_disk" ("volumeId")`)

    // Migrate existing volume-level Archil disks into the new table so they
    // are accessible through the unified lookup path. These represent "root"
    // disks (no subpath / full bucket access).
    await queryRunner.query(`
      INSERT INTO "volume_disk" ("volumeId", "subpath", "archilDiskId", "archilRegion", "archilMountTokenEnc")
      SELECT "id", NULL, "archilDiskId", "archilRegion", "archilMountTokenEnc"
      FROM "volume"
      WHERE "archilDiskId" IS NOT NULL
        AND "archilMountTokenEnc" IS NOT NULL
    `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX "idx_volume_disk_volume_id"`)
    await queryRunner.query(`DROP TABLE "volume_disk"`)
  }
}
