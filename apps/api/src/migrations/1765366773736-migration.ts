/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1765366773736 implements MigrationInterface {
  name = 'Migration1765366773736'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "recoverable" boolean NOT NULL DEFAULT false`)

    // Update existing sandboxes with recoverable error reasons to set recoverable = true
    await queryRunner.query(`
            UPDATE "sandbox" 
            SET "recoverable" = true 
            WHERE "state" = 'error' 
            AND (
                LOWER("errorReason") LIKE '%no space left on device%'
                OR LOWER("errorReason") LIKE '%storage limit%'
                OR LOWER("errorReason") LIKE '%disk quota exceeded%'
            )
        `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "recoverable"`)
  }
}
