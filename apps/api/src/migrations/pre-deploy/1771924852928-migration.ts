/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1771924852928 implements MigrationInterface {
  name = 'Migration1771924852928'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" ADD "deletedAt" TIMESTAMP WITH TIME ZONE`)

    // Note: not using CONCURRENTLY + skipping transactions because of reverting issue: https://github.com/typeorm/typeorm/issues/9981
    await queryRunner.query(
      `CREATE INDEX "idx_organization_deleted_at" ON "organization" ("deletedAt") WHERE "deletedAt" IS NOT NULL`,
    )
    await queryRunner.query(
      `CREATE INDEX "idx_organization_created_at_not_deleted" ON "organization" ("createdAt") WHERE "deletedAt" IS NULL`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX IF EXISTS "idx_organization_deleted_at"`)
    await queryRunner.query(`DROP INDEX IF EXISTS "idx_organization_created_at_not_deleted"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "deletedAt"`)
  }
}
