/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1775729705068 implements MigrationInterface {
  name = 'Migration1775729705068'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "audit_log" ADD "actorApiKeyPrefix" character varying`)
    await queryRunner.query(`ALTER TABLE "audit_log" ADD "actorApiKeySuffix" character varying`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "audit_log" DROP COLUMN "actorApiKeySuffix"`)
    await queryRunner.query(`ALTER TABLE "audit_log" DROP COLUMN "actorApiKeyPrefix"`)
  }
}
