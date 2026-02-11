/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1770000000000 implements MigrationInterface {
  name = 'Migration1770000000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Add unassigned column to sandbox table
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "unassigned" boolean NOT NULL DEFAULT false`)

    // Set unassigned = true for all existing warm pool sandboxes (organizationId = '00000000-0000-0000-0000-000000000000')
    await queryRunner.query(
      `UPDATE "sandbox" SET "unassigned" = true WHERE "organizationId" = '00000000-0000-0000-0000-000000000000'`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "unassigned"`)
  }
}
