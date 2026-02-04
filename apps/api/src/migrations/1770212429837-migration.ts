/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1770212429837 implements MigrationInterface {
  name = 'Migration1770212429837'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" ADD "authenticated_rate_limit_ttl" integer`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "sandbox_create_rate_limit_ttl" integer`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "sandbox_lifecycle_rate_limit_ttl" integer`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "sandbox_lifecycle_rate_limit_ttl"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "sandbox_create_rate_limit_ttl"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "authenticated_rate_limit_ttl"`)
  }
}
