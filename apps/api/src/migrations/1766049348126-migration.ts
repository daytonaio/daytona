/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1766049348126 implements MigrationInterface {
  name = 'Migration1766049348126'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // basic sandbox list
    await queryRunner.query(`
      CREATE INDEX CONCURRENTLY idx_sandbox_org_created
      ON sandbox ("organizationId", "createdAt" DESC)
    `)

    // sandbox list with state (and desired state) filter
    await queryRunner.query(`
      CREATE INDEX CONCURRENTLY idx_sandbox_org_state_desired_created
      ON sandbox ("organizationId", state, "desiredState", "createdAt" DESC)
    `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX IF EXISTS idx_sandbox_org_name_prefix_created`)
    await queryRunner.query(`DROP INDEX IF EXISTS idx_sandbox_org_state_desired_created`)
    await queryRunner.query(`DROP INDEX IF EXISTS idx_sandbox_org_created`)
  }
}
