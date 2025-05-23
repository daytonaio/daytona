import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1748121897392 implements MigrationInterface {
  name = 'Migration1748121897392'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.renameColumn('warm_pool', 'image', 'snapshot')

    await queryRunner.renameColumn('organization', 'image_quota', 'snapshot_quota')
    await queryRunner.renameColumn('organization', 'max_image_size', 'max_snapshot_size')
    await queryRunner.query(
      `ALTER TYPE "public"."organization_role_permissions_enum" RENAME VALUE 'write:images' TO 'write:snapshots'`,
    )
    await queryRunner.query(
      `ALTER TYPE "public"."organization_role_permissions_enum" RENAME VALUE 'delete:images' TO 'delete:snapshots'`,
    )
    await queryRunner.query(
      `ALTER TYPE "public"."api_key_permissions_enum" RENAME VALUE 'write:images' TO 'write:snapshots'`,
    )
    await queryRunner.query(
      `ALTER TYPE "public"."api_key_permissions_enum" RENAME VALUE 'delete:images' TO 'delete:snapshots'`,
    )

    await queryRunner.renameTable('image_runner', 'snapshot_runner')
    await queryRunner.renameColumn('snapshot_runner', 'imageRef', 'snapshotRef')
    await queryRunner.query(
      `ALTER TYPE "public"."snapshot_runner_state_enum" RENAME VALUE 'pulling_image' TO 'pulling_snapshot'`,
    )
    await queryRunner.query(
      `ALTER TYPE "public"."snapshot_runner_state_enum" RENAME VALUE 'building_image' TO 'building_snapshot'`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot_runner" ALTER COLUMN "state" SET DEFAULT 'pulling_snapshot'`)

    await queryRunner.renameColumn('build_info', 'imageRef', 'snapshotRef')

    await queryRunner.renameTable('image', 'snapshot')
    await queryRunner.renameColumn('snapshot', 'buildInfoImageRef', 'buildInfoSnapshotRef')
    await queryRunner.query(`ALTER TYPE "public"."snapshot_state_enum" RENAME VALUE 'pulling_image' TO 'pulling'`)
    await queryRunner.query(`ALTER TYPE "public"."snapshot_state_enum" RENAME VALUE 'building_image' TO 'building'`)

    await queryRunner.renameColumn('workspace', 'image', 'snapshot')
    await queryRunner.renameColumn('workspace', 'buildInfoImageRef', 'buildInfoSnapshotRef')
    await queryRunner.renameColumn('workspace', 'backupImage', 'backupSnapshot')
    await queryRunner.renameColumn('workspace', 'existingBackupImages', 'existingBackupSnapshots')

    await queryRunner.query(
      `ALTER TYPE "public"."workspace_state_enum" RENAME VALUE 'pulling_image' TO 'pulling_snapshot'`,
    )
    await queryRunner.query(
      `ALTER TYPE "public"."workspace_state_enum" RENAME VALUE 'building_image' TO 'building_snapshot'`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `ALTER TYPE "public"."workspace_state_enum" RENAME VALUE 'pulling_snapshot' TO 'pulling_image'`,
    )
    await queryRunner.query(
      `ALTER TYPE "public"."workspace_state_enum" RENAME VALUE 'building_snapshot' TO 'building_image'`,
    )
    await queryRunner.renameColumn('workspace', 'existingBackupSnapshots', 'existingBackupImages')
    await queryRunner.renameColumn('workspace', 'backupSnapshot', 'backupImage')
    await queryRunner.renameColumn('workspace', 'buildInfoSnapshotRef', 'buildInfoImageRef')
    await queryRunner.renameColumn('workspace', 'snapshot', 'image')

    await queryRunner.query(`ALTER TYPE "public"."snapshot_state_enum" RENAME VALUE 'pulling' TO 'pulling_image'`)
    await queryRunner.query(`ALTER TYPE "public"."snapshot_state_enum" RENAME VALUE 'building' TO 'building_image'`)
    await queryRunner.renameColumn('snapshot', 'buildInfoSnapshotRef', 'buildInfoImageRef')
    await queryRunner.renameTable('snapshot', 'image')

    await queryRunner.renameColumn('build_info', 'snapshotRef', 'imageRef')

    await queryRunner.query(
      `ALTER TYPE "public"."snapshot_runner_state_enum" RENAME VALUE 'pulling_snapshot' TO 'pulling_image'`,
    )
    await queryRunner.query(
      `ALTER TYPE "public"."snapshot_runner_state_enum" RENAME VALUE 'building_snapshot' TO 'building_image'`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot_runner" ALTER COLUMN "state" SET DEFAULT 'pulling_image'`)
    await queryRunner.renameColumn('snapshot_runner', 'snapshotRef', 'imageRef')
    await queryRunner.renameTable('snapshot_runner', 'image_runner')

    await queryRunner.query(
      `ALTER TYPE "public"."api_key_permissions_enum" RENAME VALUE 'write:snapshots' TO 'write:images'`,
    )
    await queryRunner.query(
      `ALTER TYPE "public"."api_key_permissions_enum" RENAME VALUE 'delete:snapshots' TO 'delete:images'`,
    )
    await queryRunner.query(
      `ALTER TYPE "public"."organization_role_permissions_enum" RENAME VALUE 'write:snapshots' TO 'write:images'`,
    )
    await queryRunner.query(
      `ALTER TYPE "public"."organization_role_permissions_enum" RENAME VALUE 'delete:snapshots' TO 'delete:images'`,
    )
    await queryRunner.renameColumn('organization', 'max_snapshot_size', 'max_image_size')
    await queryRunner.renameColumn('organization', 'snapshot_quota', 'image_quota')

    await queryRunner.renameColumn('warm_pool', 'snapshot', 'image')
  }
}
