/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1754658130421 implements MigrationInterface {
  name = 'Migration1754658130421'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
            CREATE TYPE "sandbox_state_enum_new" AS ENUM (
                'creating',
                'restoring',
                'destroyed',
                'destroying',
                'started',
                'stopped',
                'starting',
                'stopping',
                'error',
                'build_failed',
                'pending_build',
                'building_snapshot',
                'unknown',
                'pulling_snapshot',
                'pending_archive',
                'archiving',
                'archived'
            )
        `)

    await queryRunner.query(`
            ALTER TABLE "sandbox"
            ALTER COLUMN "state" DROP DEFAULT
        `)

    await queryRunner.query(`
            ALTER TABLE "sandbox"
            ALTER COLUMN "state"
            TYPE "sandbox_state_enum_new"
            USING "state"::text::"sandbox_state_enum_new"
        `)

    await queryRunner.query(`DROP TYPE "sandbox_state_enum"`)

    await queryRunner.query(`
            ALTER TYPE "sandbox_state_enum_new"
            RENAME TO "sandbox_state_enum"
        `)

    await queryRunner.query(`
            ALTER TABLE "sandbox"
            ALTER COLUMN "state" SET DEFAULT 'unknown'
        `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
            UPDATE "sandbox"
            SET "state" = 'stopped'
            WHERE "state" = 'pending_archive'
        `)

    await queryRunner.query(`
            ALTER TABLE "sandbox"
            ALTER COLUMN "state" DROP DEFAULT
        `)

    await queryRunner.query(`
            CREATE TYPE "sandbox_state_enum_old" AS ENUM (
                'archived',
                'archiving',
                'build_failed',
                'building_snapshot',
                'creating',
                'destroyed',
                'destroying',
                'error',
                'pending_build',
                'pulling_snapshot',
                'restoring',
                'started',
                'starting',
                'stopped',
                'stopping',
                'unknown'
            )
        `)

    await queryRunner.query(`
            ALTER TABLE "sandbox"
            ALTER COLUMN "state"
            TYPE "sandbox_state_enum_old"
            USING "state"::text::"sandbox_state_enum_old"
        `)

    await queryRunner.query(`DROP TYPE "sandbox_state_enum"`)

    await queryRunner.query(`
            ALTER TYPE "sandbox_state_enum_old"
            RENAME TO "sandbox_state_enum"
        `)

    await queryRunner.query(`
            ALTER TABLE "sandbox"
            ALTER COLUMN "state" SET DEFAULT 'unknown'
        `)
  }
}
