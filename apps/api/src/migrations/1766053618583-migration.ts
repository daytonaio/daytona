/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1766053618583 implements MigrationInterface {
  name = 'Migration1766053618583'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "job" ALTER COLUMN "type" TYPE character varying USING "type"::text`)
    await queryRunner.query(`ALTER TABLE "job" ALTER COLUMN "type" SET NOT NULL`)
    await queryRunner.query(`DROP TYPE "public"."job_type_enum"`)
    await queryRunner.query(
      `ALTER TABLE "job" ADD CONSTRAINT "VALIDATE_JOB_TYPE" CHECK ("type" IN ('CREATE_SANDBOX', 'START_SANDBOX', 'STOP_SANDBOX', 'DESTROY_SANDBOX', 'CREATE_BACKUP', 'BUILD_SNAPSHOT', 'PULL_SNAPSHOT', 'REMOVE_SNAPSHOT', 'UPDATE_SANDBOX_NETWORK_SETTINGS'))`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "job" DROP CONSTRAINT "VALIDATE_JOB_TYPE"`)
    await queryRunner.query(
      `ALTER TABLE "job" ALTER COLUMN "type" TYPE "public"."job_type_enum" USING "type"::"public"."job_type_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "job" ALTER COLUMN "type" SET NOT NULL`)
  }
}
