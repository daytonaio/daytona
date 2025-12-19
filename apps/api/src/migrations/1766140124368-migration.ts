/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1766140124368 implements MigrationInterface {
  name = 'Migration1766140124368'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.renameColumn('runner', 'version', 'apiVersion')
    await queryRunner.query(`ALTER TABLE "runner" ADD "appVersion" character varying DEFAULT 'v0.0.0-dev'`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "appVersion"`)
    await queryRunner.renameColumn('runner', 'apiVersion', 'version')
  }
}
