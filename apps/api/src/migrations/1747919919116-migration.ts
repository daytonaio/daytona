import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1747919919116 implements MigrationInterface {
  name = 'Migration1747919919116'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.renameColumn('workspace', 'snapshotRegistryId', 'backupRegistryId')
    await queryRunner.renameColumn('workspace', 'snapshotImage', 'backupImage')
    await queryRunner.renameColumn('workspace', 'lastSnapshotAt', 'lastBackupAt')
    await queryRunner.renameColumn('workspace', 'snapshotState', 'backupState')
    await queryRunner.query(`ALTER TYPE "public"."workspace_snapshotstate_enum" RENAME TO "workspace_backupstate_enum"`)
    await queryRunner.renameColumn('workspace', 'existingSnapshotImages', 'existingBackupImages')
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.renameColumn('workspace', 'existingBackupImages', 'existingSnapshotImages')
    await queryRunner.query(`ALTER TYPE "public"."workspace_backupstate_enum" RENAME TO "workspace_snapshotstate_enum"`)
    await queryRunner.renameColumn('workspace', 'backupState', 'snapshotState')
    await queryRunner.renameColumn('workspace', 'lastBackupAt', 'lastSnapshotAt')
    await queryRunner.renameColumn('workspace', 'backupImage', 'snapshotImage')
    await queryRunner.renameColumn('workspace', 'backupRegistryId', 'snapshotRegistryId')
  }
}
