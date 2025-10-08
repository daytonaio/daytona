/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1759241690774 implements MigrationInterface {
  name = 'Migration1759241690774'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "backupStartedAt" TIMESTAMP WITH TIME ZONE`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "backupStartedAt"`)
  }
}
