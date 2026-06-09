/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1780918142801 implements MigrationInterface {
  name = 'Migration1780918142801'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "runnerClass"`)
    await queryRunner.query(`DROP TYPE "public"."runner_runnerclass_enum"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`CREATE TYPE "public"."runner_runnerclass_enum" AS ENUM('container', 'vm')`)
    await queryRunner.query(
      `ALTER TABLE "runner" ADD "runnerClass" "public"."runner_runnerclass_enum" NOT NULL DEFAULT 'container'`,
    )
  }
}
