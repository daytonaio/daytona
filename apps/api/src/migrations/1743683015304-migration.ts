/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1743683015304 implements MigrationInterface {
  name = 'Migration1743683015304'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "name"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP CONSTRAINT "PK_ca86b6f9b3be5fe26d307d09b49"`)
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "id" SET DEFAULT uuid_generate_v4()`)
    await queryRunner.query(`ALTER TABLE "workspace" ADD CONSTRAINT "workspace_id_pk" PRIMARY KEY ("id")`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "workspace" DROP CONSTRAINT "workspace_id_pk"`)
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "id" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "workspace" ADD CONSTRAINT "PK_ca86b6f9b3be5fe26d307d09b49" PRIMARY KEY ("id")`,
    )
    await queryRunner.query(`ALTER TABLE "workspace" ADD "name" character varying NOT NULL DEFAULT ''`)
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "name" DROP DEFAULT`)
  }
}
