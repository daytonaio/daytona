/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1763561822000 implements MigrationInterface {
  name = 'Migration1763561822000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" ADD "authenticated_rate_limit" integer`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "sandbox_create_rate_limit" integer`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "sandbox_lifecycle_rate_limit" integer`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "sandbox_lifecycle_rate_limit"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "sandbox_create_rate_limit"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "authenticated_rate_limit"`)
  }
}
