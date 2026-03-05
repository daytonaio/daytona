/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class PostDeployMigration1770900000001 implements MigrationInterface {
  name = 'PostDeployMigration1770900000001'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "snapshot_quota"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" ADD "snapshot_quota" integer NOT NULL DEFAULT 100`)
    await queryRunner.query(`UPDATE "organization" SET "snapshot_quota" = "active_snapshot_quota"`)
  }
}
