/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1769600000000 implements MigrationInterface {
  name = 'Migration1769600000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" ADD "deletedAt" TIMESTAMP WITH TIME ZONE`)

    // TODO: create concurrently
    await queryRunner.query(
      `CREATE INDEX "idx_organization_deleted_at" ON "organization" ("deletedAt") WHERE "deletedAt" IS NOT NULL`,
    )

    await queryRunner.query(
      `CREATE INDEX "idx_organization_created_at_not_deleted" ON "organization" ("createdAt") WHERE "deletedAt" IS NULL`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "deletedAt"`)
  }
}
