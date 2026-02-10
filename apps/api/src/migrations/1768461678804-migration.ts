/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1768461678804 implements MigrationInterface {
  name = 'Migration1768461678804'

  // TODO: Add migrationsTransactionMode: 'each', to data-source.ts
  // TypeORM currently does not support non-transactional reverts
  // Needed because CREATE/DROP INDEX CONCURRENTLY cannot run inside a transaction
  // public readonly transaction = false

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
      CREATE INDEX IF NOT EXISTS "idx_sandbox_volumes_gin"
      ON "sandbox"
      USING GIN ("volumes" jsonb_path_ops)
      WHERE "desiredState" <> 'destroyed';
    `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
      DROP INDEX IF EXISTS "idx_sandbox_volumes_gin";
    `)
  }
}
