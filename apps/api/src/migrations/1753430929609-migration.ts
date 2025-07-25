/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1753430929609 implements MigrationInterface {
  name = 'Migration1753430929609'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "user" ADD "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`)

    // For existing users, set createdAt to match their personal organization's createdAt
    await queryRunner.query(`
        UPDATE "user" u 
        SET "createdAt" = o."createdAt" 
        FROM "organization" o 
        WHERE o."createdBy" = u.id 
          AND o.personal = true;
    `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "user" DROP COLUMN "createdAt"`)
  }
}
