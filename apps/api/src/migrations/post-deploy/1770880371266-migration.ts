/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1770880371266 implements MigrationInterface {
  name = 'Migration1770880371266'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Note: not using CONCURRENTLY + skipping transactions because of reverting issue: https://github.com/typeorm/typeorm/issues/9981
    await queryRunner.query(`DROP INDEX "public"."IDX_UNIQUE_INCOMPLETE_JOB"`)
    await queryRunner.query(
      `CREATE UNIQUE INDEX "IDX_UNIQUE_INCOMPLETE_JOB" ON "job" ("resourceType", "resourceId", "runnerId") WHERE "completedAt" IS NULL AND "type" != 'CREATE_BACKUP'`,
    )
    await queryRunner.query(
      `CREATE UNIQUE INDEX "IDX_UNIQUE_INCOMPLETE_BACKUP_JOB" ON "job" ("resourceType", "resourceId", "runnerId") WHERE "completedAt" IS NULL AND "type" = 'CREATE_BACKUP'`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX "public"."IDX_UNIQUE_INCOMPLETE_BACKUP_JOB"`)
    await queryRunner.query(`DROP INDEX "public"."IDX_UNIQUE_INCOMPLETE_JOB"`)
    await queryRunner.query(
      `CREATE UNIQUE INDEX "IDX_UNIQUE_INCOMPLETE_JOB" ON "job" ("resourceType", "resourceId", "runnerId") WHERE "completedAt" IS NULL`,
    )
  }
}
