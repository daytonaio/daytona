/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1770000000002 implements MigrationInterface {
  name = 'Migration1770000000002'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Add organizationId column to warm_pool table with default value
    await queryRunner.query(
      `ALTER TABLE "warm_pool" ADD "organizationId" uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000'`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "warm_pool" DROP COLUMN "organizationId"`)
  }
}
