/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1768300000000 implements MigrationInterface {
  name = 'Migration1768300000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Add parentSandboxId column to sandbox table for tracking fork lineage
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "parentSandboxId" uuid`)

    // Update the VALIDATE_JOB_TYPE constraint to include FORK_SANDBOX
    await queryRunner.query(`ALTER TABLE "job" DROP CONSTRAINT "VALIDATE_JOB_TYPE"`)
    await queryRunner.query(
      `ALTER TABLE "job" ADD CONSTRAINT "VALIDATE_JOB_TYPE" CHECK ("type" IN ('CREATE_SANDBOX', 'START_SANDBOX', 'STOP_SANDBOX', 'DESTROY_SANDBOX', 'CREATE_BACKUP', 'BUILD_SNAPSHOT', 'PULL_SNAPSHOT', 'REMOVE_SNAPSHOT', 'UPDATE_SANDBOX_NETWORK_SETTINGS', 'CREATE_SANDBOX_SNAPSHOT', 'FORK_SANDBOX'))`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Revert the VALIDATE_JOB_TYPE constraint to exclude FORK_SANDBOX
    await queryRunner.query(`ALTER TABLE "job" DROP CONSTRAINT "VALIDATE_JOB_TYPE"`)
    await queryRunner.query(
      `ALTER TABLE "job" ADD CONSTRAINT "VALIDATE_JOB_TYPE" CHECK ("type" IN ('CREATE_SANDBOX', 'START_SANDBOX', 'STOP_SANDBOX', 'DESTROY_SANDBOX', 'CREATE_BACKUP', 'BUILD_SNAPSHOT', 'PULL_SNAPSHOT', 'REMOVE_SNAPSHOT', 'UPDATE_SANDBOX_NETWORK_SETTINGS', 'CREATE_SANDBOX_SNAPSHOT'))`,
    )

    // Remove parentSandboxId column from sandbox table
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "parentSandboxId"`)
  }
}
