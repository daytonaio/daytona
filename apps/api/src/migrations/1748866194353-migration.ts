/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1748866194353 implements MigrationInterface {
  name = 'Migration1748866194353'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "workspace" ADD "autoArchiveInterval" integer NOT NULL DEFAULT '10080'`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "autoArchiveInterval"`)
  }
}
