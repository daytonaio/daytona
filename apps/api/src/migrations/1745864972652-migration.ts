/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1745864972652 implements MigrationInterface {
  name = 'Migration1745864972652'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
            ALTER TABLE "workspace"
              ALTER COLUMN "env" DROP DEFAULT,
              ALTER COLUMN "env" TYPE jsonb USING "env"::jsonb,
              ALTER COLUMN "env" SET DEFAULT '{}'::jsonb,
              ALTER COLUMN "env" SET NOT NULL;
          `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
            ALTER TABLE "workspace"
              ALTER COLUMN "env" DROP DEFAULT,
              ALTER COLUMN "env" TYPE text USING "env"::text,
              ALTER COLUMN "env" SET DEFAULT '{}'::text,
              ALTER COLUMN "env" SET NOT NULL;
          `)
  }
}
