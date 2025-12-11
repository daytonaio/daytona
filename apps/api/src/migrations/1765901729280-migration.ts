/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1765901729280 implements MigrationInterface {
  name = 'Migration1765901729280'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX "public"."job_resourceType_resourceId_index"`)
    await queryRunner.query(`ALTER TABLE "job" ALTER COLUMN "resourceType" SET NOT NULL`)
    await queryRunner.query(`ALTER TABLE "job" ALTER COLUMN "resourceId" SET NOT NULL`)
    await queryRunner.query(`ALTER TABLE "job" ALTER COLUMN "payload" TYPE character varying`)
    await queryRunner.query(`ALTER TABLE "job" ADD "resultMetadata" character varying`)
    await queryRunner.query(
      `CREATE UNIQUE INDEX "IDX_UNIQUE_INCOMPLETE_JOB" ON "job" ("resourceType", "resourceId") WHERE "completedAt" IS NULL`,
    )
    await queryRunner.query(`CREATE INDEX "job_resourceType_resourceId_index" ON "job" ("resourceType", "resourceId") `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX "public"."job_resourceType_resourceId_index"`)
    await queryRunner.query(`DROP INDEX "public"."IDX_UNIQUE_INCOMPLETE_JOB"`)
    await queryRunner.query(`ALTER TABLE "job" DROP COLUMN "resultMetadata"`)
    await queryRunner.query(`ALTER TABLE "job" DROP COLUMN "payload"`)
    await queryRunner.query(`ALTER TABLE "job" ADD "payload" jsonb`)
    await queryRunner.query(`ALTER TABLE "job" ALTER COLUMN "resourceId" DROP NOT NULL`)
    await queryRunner.query(`ALTER TABLE "job" ALTER COLUMN "resourceType" DROP NOT NULL`)
    await queryRunner.query(`CREATE INDEX "job_resourceType_resourceId_index" ON "job" ("resourceId", "resourceType") `)
  }
}
