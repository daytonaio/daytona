/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1770000000004 implements MigrationInterface {
  name = 'Migration1770000000004'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Add unassigned column to sandbox_usage_periods table
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" ADD "unassigned" boolean NOT NULL DEFAULT false`)

    // Add unassigned column to sandbox_usage_periods_archive table
    await queryRunner.query(
      `ALTER TABLE "sandbox_usage_periods_archive" ADD "unassigned" boolean NOT NULL DEFAULT false`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods_archive" DROP COLUMN "unassigned"`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" DROP COLUMN "unassigned"`)
  }
}
