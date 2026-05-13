/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1777170000000 implements MigrationInterface {
  name = 'Migration1777170000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Volume table: identify the backend for each row and (for layered)
    // record the disk + region. The per-attachment mount token lives on
    // `sandbox_volume`, not here.
    await queryRunner.query(`ALTER TABLE "volume" ADD "backend" character varying NOT NULL DEFAULT 's3fuse'`)
    await queryRunner.query(`ALTER TABLE "volume" ADD "layeredDiskId" character varying`)
    await queryRunner.query(`ALTER TABLE "volume" ADD "layeredRegion" character varying`)

    // Per-mount attachment table. Holds only layered mounts; legacy s3fuse
    // mounts continue to live on the `sandbox.volumes` JSONB column.
    await queryRunner.query(`
      CREATE TABLE "sandbox_volume" (
        "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
        "sandboxId" uuid NOT NULL,
        "volumeId" uuid NOT NULL,
        "mountPath" character varying NOT NULL,
        "subpath" character varying,
        "readOnly" boolean NOT NULL DEFAULT false,
        "mountKeyEnc" text,
        "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
        "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
        CONSTRAINT "PK_sandbox_volume_id" PRIMARY KEY ("id"),
        CONSTRAINT "UQ_sandbox_volume_sandbox_volume_mountpath" UNIQUE ("sandboxId", "volumeId", "mountPath")
      )
    `)
    await queryRunner.query(`CREATE INDEX "IDX_sandbox_volume_sandboxId" ON "sandbox_volume" ("sandboxId")`)
    await queryRunner.query(`CREATE INDEX "IDX_sandbox_volume_volumeId" ON "sandbox_volume" ("volumeId")`)
    await queryRunner.query(
      `ALTER TABLE "sandbox_volume" ADD CONSTRAINT "FK_sandbox_volume_sandbox" FOREIGN KEY ("sandboxId") REFERENCES "sandbox"("id") ON DELETE CASCADE ON UPDATE NO ACTION`,
    )
    await queryRunner.query(
      `ALTER TABLE "sandbox_volume" ADD CONSTRAINT "FK_sandbox_volume_volume" FOREIGN KEY ("volumeId") REFERENCES "volume"("id") ON DELETE RESTRICT ON UPDATE NO ACTION`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox_volume" DROP CONSTRAINT "FK_sandbox_volume_volume"`)
    await queryRunner.query(`ALTER TABLE "sandbox_volume" DROP CONSTRAINT "FK_sandbox_volume_sandbox"`)
    await queryRunner.query(`DROP INDEX "IDX_sandbox_volume_volumeId"`)
    await queryRunner.query(`DROP INDEX "IDX_sandbox_volume_sandboxId"`)
    await queryRunner.query(`DROP TABLE "sandbox_volume"`)
    await queryRunner.query(`ALTER TABLE "volume" DROP COLUMN "layeredRegion"`)
    await queryRunner.query(`ALTER TABLE "volume" DROP COLUMN "layeredDiskId"`)
    await queryRunner.query(`ALTER TABLE "volume" DROP COLUMN "backend"`)
  }
}
