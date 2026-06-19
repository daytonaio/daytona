/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1781308800000 implements MigrationInterface {
  name = 'Migration1781308800000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TABLE "sandbox_metadata" ("sandboxId" character varying NOT NULL, "signingKey" character varying NOT NULL, CONSTRAINT "PK_sandbox_metadata" PRIMARY KEY ("sandboxId"))`,
    )
    await queryRunner.query(
      `ALTER TABLE "sandbox_metadata" ADD CONSTRAINT "FK_sandbox_metadata_sandbox" FOREIGN KEY ("sandboxId") REFERENCES "sandbox"("id") ON DELETE CASCADE ON UPDATE NO ACTION`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox_metadata" DROP CONSTRAINT "FK_sandbox_metadata_sandbox"`)
    await queryRunner.query(`DROP TABLE "sandbox_metadata"`)
  }
}
