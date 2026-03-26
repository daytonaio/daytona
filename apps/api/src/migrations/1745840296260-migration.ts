/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'
import * as crypto from 'crypto'

export class Migration1745840296260 implements MigrationInterface {
  name = 'Migration1745840296260'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Add the new columns
    await queryRunner.query(`ALTER TABLE "api_key" ADD "keyHash" character varying NOT NULL DEFAULT ''`)
    await queryRunner.query(`ALTER TABLE "api_key" ADD "keyPrefix" character varying NOT NULL DEFAULT ''`)
    await queryRunner.query(`ALTER TABLE "api_key" ADD "keySuffix" character varying NOT NULL DEFAULT ''`)

    // Get all existing API keys
    const existingKeys = await queryRunner.query(`SELECT "value" FROM "api_key"`)

    // Update each key with its hash, prefix, and suffix
    for (const key of existingKeys) {
      const value = key.value
      const keyHash = crypto.createHash('sha256').update(value).digest('hex')
      const keyPrefix = value.substring(0, 3)
      const keySuffix = value.slice(-3)

      await queryRunner.query(
        `UPDATE "api_key" 
                SET "keyHash" = $1, 
                    "keyPrefix" = $2, 
                    "keySuffix" = $3 
                WHERE "value" = $4`,
        [keyHash, keyPrefix, keySuffix, value],
      )
    }

    // Drop value column and its unique constraint
    await queryRunner.query(`ALTER TABLE "api_key" DROP CONSTRAINT "UQ_4b0873b633484d5de20b2d8f852"`)
    await queryRunner.query(`ALTER TABLE "api_key" DROP COLUMN "value"`)

    // Add unique constraint
    await queryRunner.query(`ALTER TABLE "api_key" ADD CONSTRAINT "api_key_keyHash_unique" UNIQUE ("keyHash")`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Revert the schema changes
    await queryRunner.query(`ALTER TABLE "api_key" DROP CONSTRAINT "api_key_keyHash_unique"`)
    await queryRunner.query(`ALTER TABLE "api_key" DROP COLUMN "keySuffix"`)
    await queryRunner.query(`ALTER TABLE "api_key" DROP COLUMN "keyPrefix"`)
    await queryRunner.query(`ALTER TABLE "api_key" DROP COLUMN "keyHash"`)
    await queryRunner.query(`ALTER TABLE "api_key" ADD "value" character varying NOT NULL DEFAULT ''`)
  }
}
