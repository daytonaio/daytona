/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1737539891636 implements MigrationInterface {
  name = 'Migration1737539891636'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TYPE "public"."sandbox_wakeonrequest_enum" AS ENUM('none', 'http', 'ssh', 'http_and_ssh')`,
    )
    await queryRunner.query(
      `ALTER TABLE "sandbox" ADD "wakeOnRequest" "public"."sandbox_wakeonrequest_enum" NOT NULL DEFAULT 'none'`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "wakeOnRequest"`)
    await queryRunner.query(`DROP TYPE "public"."sandbox_wakeonrequest_enum"`)
  }
}
