/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1770900000001 implements MigrationInterface {
  name = 'Migration1770900000001'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" ADD "active_snapshot_quota" integer NOT NULL DEFAULT 100`)
    await queryRunner.query(`UPDATE "organization" SET "active_snapshot_quota" = "snapshot_quota"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "active_snapshot_quota"`)
  }
}
