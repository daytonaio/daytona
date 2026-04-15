/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1772734390158 implements MigrationInterface {
  name = 'Migration1772734390158'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "snapshot_runner" ADD "lastUsedAt" TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(
      // Note: not using CONCURRENTLY + skipping transactions because of reverting issue: https://github.com/typeorm/typeorm/issues/9981
      `CREATE INDEX "snapshot_runner_lastusedat_idx" ON "snapshot_runner" ("lastUsedAt")`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX "public"."snapshot_runner_lastusedat_idx"`)
    await queryRunner.query(`ALTER TABLE "snapshot_runner" DROP COLUMN "lastUsedAt"`)
  }
}
