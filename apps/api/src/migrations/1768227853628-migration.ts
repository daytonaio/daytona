/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1768227853628 implements MigrationInterface {
  name = 'Migration1768227853628'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Drop the existing constraint
    await queryRunner.query(`ALTER TABLE "job" DROP CONSTRAINT "VALIDATE_JOB_TYPE"`)
    // Add the new constraint with CREATE_SANDBOX_SNAPSHOT included
    await queryRunner.query(
      `ALTER TABLE "job" ADD CONSTRAINT "VALIDATE_JOB_TYPE" CHECK ("type" IN ('CREATE_SANDBOX', 'START_SANDBOX', 'STOP_SANDBOX', 'DESTROY_SANDBOX', 'CREATE_BACKUP', 'BUILD_SNAPSHOT', 'PULL_SNAPSHOT', 'REMOVE_SNAPSHOT', 'UPDATE_SANDBOX_NETWORK_SETTINGS', 'CREATE_SANDBOX_SNAPSHOT'))`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Drop the constraint with CREATE_SANDBOX_SNAPSHOT
    await queryRunner.query(`ALTER TABLE "job" DROP CONSTRAINT "VALIDATE_JOB_TYPE"`)
    // Re-add the constraint without CREATE_SANDBOX_SNAPSHOT
    await queryRunner.query(
      `ALTER TABLE "job" ADD CONSTRAINT "VALIDATE_JOB_TYPE" CHECK ("type" IN ('CREATE_SANDBOX', 'START_SANDBOX', 'STOP_SANDBOX', 'DESTROY_SANDBOX', 'CREATE_BACKUP', 'BUILD_SNAPSHOT', 'PULL_SNAPSHOT', 'REMOVE_SNAPSHOT', 'UPDATE_SANDBOX_NETWORK_SETTINGS'))`,
    )
  }
}
