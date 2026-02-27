// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1772112977674 implements MigrationInterface {
  name = 'Migration1772112977674'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ADD "serviceHealth" jsonb`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "serviceHealth"`)
  }
}
