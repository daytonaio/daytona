/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1776946267385 implements MigrationInterface {
  name = 'Migration1776946267385'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Note: not using CONCURRENTLY + skipping transactions because of reverting issue: https://github.com/typeorm/typeorm/issues/9981
    await queryRunner.query(
      `CREATE INDEX "sandbox_buildinfosnapshotref_idx" ON "sandbox" ("buildInfoSnapshotRef") WHERE "buildInfoSnapshotRef" IS NOT NULL`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX "sandbox_buildinfosnapshotref_idx"`)
  }
}
