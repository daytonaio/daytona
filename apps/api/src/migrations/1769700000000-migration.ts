/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1769700000000 implements MigrationInterface {
  name = 'Migration1769700000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Update the VALIDATE_JOB_TYPE constraint to include CLONE_SANDBOX
    await queryRunner.query(`ALTER TABLE "job" DROP CONSTRAINT "VALIDATE_JOB_TYPE"`)
    await queryRunner.query(
      `ALTER TABLE "job" ADD CONSTRAINT "VALIDATE_JOB_TYPE" CHECK ("type" IN ('CREATE_SANDBOX', 'START_SANDBOX', 'STOP_SANDBOX', 'DESTROY_SANDBOX', 'CREATE_BACKUP', 'BUILD_SNAPSHOT', 'PULL_SNAPSHOT', 'REMOVE_SNAPSHOT', 'UPDATE_SANDBOX_NETWORK_SETTINGS', 'CREATE_SANDBOX_SNAPSHOT', 'FORK_SANDBOX', 'CLONE_SANDBOX'))`,
    )

    // Add sourceSandboxId column for clone tracking
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "sourceSandboxId" uuid`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Remove sourceSandboxId column
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "sourceSandboxId"`)

    // Revert the VALIDATE_JOB_TYPE constraint to exclude CLONE_SANDBOX
    await queryRunner.query(`ALTER TABLE "job" DROP CONSTRAINT "VALIDATE_JOB_TYPE"`)
    await queryRunner.query(
      `ALTER TABLE "job" ADD CONSTRAINT "VALIDATE_JOB_TYPE" CHECK ("type" IN ('CREATE_SANDBOX', 'START_SANDBOX', 'STOP_SANDBOX', 'DESTROY_SANDBOX', 'CREATE_BACKUP', 'BUILD_SNAPSHOT', 'PULL_SNAPSHOT', 'REMOVE_SNAPSHOT', 'UPDATE_SANDBOX_NETWORK_SETTINGS', 'CREATE_SANDBOX_SNAPSHOT', 'FORK_SANDBOX'))`,
    )
  }
}
