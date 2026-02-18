/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1770902744887 implements MigrationInterface {
  name = 'Migration1770902744887'

  // Note: not using CONCURRENTLY + skipping transactions because of reverting issue: https://github.com/typeorm/typeorm/issues/9981
  public async up(queryRunner: QueryRunner): Promise<void> {
    // Basic sandbox list
    await queryRunner.query(`
      CREATE INDEX idx_sandbox_org_created
      ON sandbox ("organizationId", "createdAt" DESC, id DESC)
      WHERE "desiredState" != 'destroyed'
    `)

    // Sandbox list with state (and desired state) filter
    await queryRunner.query(`
      CREATE INDEX idx_sandbox_org_state_created
      ON sandbox ("organizationId", state, "createdAt" DESC, id DESC)
      WHERE "desiredState" != 'destroyed'
    `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX IF EXISTS idx_sandbox_org_state_created`)
    await queryRunner.query(`DROP INDEX IF EXISTS idx_sandbox_org_created`)
  }
}
