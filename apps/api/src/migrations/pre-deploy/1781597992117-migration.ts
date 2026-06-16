/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1781597992117 implements MigrationInterface {
  name = 'Migration1781597992117'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `ALTER TABLE "sandbox_usage_periods_archive" ADD "regionType" character varying NOT NULL DEFAULT 'shared'`,
    )
    await queryRunner.query(
      `ALTER TABLE "sandbox_usage_periods" ADD "regionType" character varying NOT NULL DEFAULT 'shared'`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" DROP COLUMN "regionType"`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods_archive" DROP COLUMN "regionType"`)
  }
}
