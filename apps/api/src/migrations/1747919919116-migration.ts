/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1747919919116 implements MigrationInterface {
  name = 'Migration1747919919116'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.renameColumn('workspace', 'snapshotRegistryId', 'backupRegistryId')
    await queryRunner.renameColumn('workspace', 'snapshotImage', 'backupImage')
    await queryRunner.renameColumn('workspace', 'lastSnapshotAt', 'lastBackupAt')
    await queryRunner.renameColumn('workspace', 'snapshotState', 'backupState')
    await queryRunner.renameColumn('workspace', 'existingSnapshotImages', 'existingBackupImages')
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.renameColumn('workspace', 'existingBackupImages', 'existingSnapshotImages')
    await queryRunner.renameColumn('workspace', 'backupState', 'snapshotState')
    await queryRunner.renameColumn('workspace', 'lastBackupAt', 'lastSnapshotAt')
    await queryRunner.renameColumn('workspace', 'backupImage', 'snapshotImage')
    await queryRunner.renameColumn('workspace', 'backupRegistryId', 'snapshotRegistryId')
  }
}
