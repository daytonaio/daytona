/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1774438866002 implements MigrationInterface {
  name = 'Migration1774438866002'

  public async up(queryRunner: QueryRunner): Promise<void> {
    /**
     * Drop DB-level default for sandbox name. Now set exclusively in the entity constructor.
     */
    await queryRunner.query(`ALTER TABLE "sandbox" DROP CONSTRAINT "sandbox_organizationId_name_unique"`)
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "name" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "sandbox" ADD CONSTRAINT "sandbox_organizationId_name_unique" UNIQUE ("organizationId", "name")`,
    )

    /**
     * Drop DB-level default for sandbox authToken. Now set exclusively via class field initializer.
     */
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "authToken" DROP DEFAULT`)

    /**
     * The initial migration incorrectly added the column, it was never actually added to the entity definition.
     */
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "sshPass"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `ALTER TABLE "sandbox" ADD "sshPass" character varying(32) NOT NULL DEFAULT REPLACE(uuid_generate_v4()::text, '-', '')`,
    )

    await queryRunner.query(`ALTER TABLE "sandbox" DROP CONSTRAINT "sandbox_organizationId_name_unique"`)
    await queryRunner.query(
      `ALTER TABLE "sandbox" ALTER COLUMN "name" SET DEFAULT 'sandbox-' || substring(gen_random_uuid()::text, 1, 10)`,
    )
    await queryRunner.query(
      `ALTER TABLE "sandbox" ADD CONSTRAINT "sandbox_organizationId_name_unique" UNIQUE ("organizationId", "name")`,
    )

    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "authToken" SET DEFAULT MD5(random()::text)`)
  }
}
