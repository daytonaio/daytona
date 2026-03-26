/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1744028841133 implements MigrationInterface {
  name = 'Migration1744028841133'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "workspace_usage_periods" DROP COLUMN "storage"`)
    await queryRunner.query(`ALTER TABLE "workspace_usage_periods" ADD "organizationId" character varying NOT NULL`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "workspace_usage_periods" DROP COLUMN "organizationId"`)
    await queryRunner.query(`ALTER TABLE "workspace_usage_periods" ADD "storage" double precision NOT NULL`)
  }
}
