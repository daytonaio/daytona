/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1781622481905 implements MigrationInterface {
  name = 'Migration1781622481905'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP TRIGGER IF EXISTS sandbox_lifecycle_sync ON "sandbox"`)
    await queryRunner.query(`DROP FUNCTION IF EXISTS sync_sandbox_lifecycle()`)

    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "state"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "desiredState"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "pending"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "errorReason"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "recoverable"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "daemonVersion"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "runnerId"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "prevRunnerId"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "backupState"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "lastBackupAt"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "backupSnapshot"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "backupRegistryId"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "backupErrorReason"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "existingBackupSnapshots"`)
  }

  public async down(): Promise<void> {
    throw new Error(
      'Reverting this migration requires a manual procedure for re-adding the 14 columns and re-seeding them from the sandbox_lifecycle table.',
    )
  }
}
