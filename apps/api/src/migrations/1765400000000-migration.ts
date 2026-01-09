/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1765400000000 implements MigrationInterface {
  name = 'Migration1765400000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Normalize Docker Hub URLs to 'docker.io' for consistency
    // The runner will convert to 'index.docker.io/v1/' for builds where needed
    await queryRunner.query(`
      UPDATE "docker_registry"
      SET "url" = 'docker.io'
      WHERE LOWER("url") LIKE '%docker.io%'
    `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Cannot reliably reverse this migration as we don't know the original URLs
    // This is a one-way normalization
  }
}
