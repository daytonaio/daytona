/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1781022372014 implements MigrationInterface {
  name = 'Migration1781022372014'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TYPE "public"."sandbox_state_enum" ADD VALUE IF NOT EXISTS 'pausing'`)
    await queryRunner.query(`ALTER TYPE "public"."sandbox_state_enum" ADD VALUE IF NOT EXISTS 'paused'`)
    await queryRunner.query(`ALTER TYPE "public"."sandbox_state_enum" ADD VALUE IF NOT EXISTS 'resuming'`)
    await queryRunner.query(`ALTER TYPE "public"."sandbox_desiredstate_enum" ADD VALUE IF NOT EXISTS 'paused'`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // PostgreSQL does not support removing individual enum values.
    // A full revert would require the rename-recreate-drop pattern
    // with all dependent indexes dropped and recreated.
  }
}
