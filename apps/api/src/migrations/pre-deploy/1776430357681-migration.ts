/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1776430357681 implements MigrationInterface {
  name = 'Migration1776430357681'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Note: not using CONCURRENTLY + skipping transactions because of reverting issue: https://github.com/typeorm/typeorm/issues/9981
    await queryRunner.query(`CREATE INDEX idx_sandbox_last_activity_at ON sandbox_last_activity ("lastActivityAt")`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX idx_sandbox_last_activity_at`)
  }
}
