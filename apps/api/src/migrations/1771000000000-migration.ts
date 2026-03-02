/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1771000000000 implements MigrationInterface {
  name = 'Migration1771000000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
      CREATE TABLE "sandbox_last_activity" (
        "sandboxId" uuid NOT NULL,
        "lastActivityAt" TIMESTAMP WITH TIME ZONE NOT NULL,
        CONSTRAINT "PK_sandbox_last_activity" PRIMARY KEY ("sandboxId"),
        CONSTRAINT "FK_sandbox_last_activity_sandbox" FOREIGN KEY ("sandboxId")
          REFERENCES "sandbox"("id") ON DELETE CASCADE
      )
    `)

    // Populate from existing sandbox.lastActivityAt
    await queryRunner.query(`
      INSERT INTO "sandbox_last_activity" ("sandboxId", "lastActivityAt")
      SELECT "id", COALESCE("lastActivityAt", "createdAt", NOW())
      FROM "sandbox"
      WHERE "state" NOT IN ('destroyed')
    `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP TABLE "sandbox_last_activity"`)
  }
}
