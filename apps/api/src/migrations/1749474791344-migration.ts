/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1749474791344 implements MigrationInterface {
  name = 'Migration1749474791344'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Snapshot to backup rename
    await queryRunner.renameColumn('workspace', 'snapshotRegistryId', 'backupRegistryId')
    await queryRunner.renameColumn('workspace', 'snapshotImage', 'backupImage')
    await queryRunner.renameColumn('workspace', 'lastSnapshotAt', 'lastBackupAt')
    await queryRunner.renameColumn('workspace', 'snapshotState', 'backupState')
    await queryRunner.renameColumn('workspace', 'existingSnapshotImages', 'existingBackupImages')

    // Node to runner rename
    await queryRunner.renameColumn('image', 'buildNodeId', 'buildRunnerId')
    await queryRunner.renameTable('image_node', 'image_runner')
    await queryRunner.renameColumn('image_runner', 'nodeId', 'runnerId')
    await queryRunner.renameTable('node', 'runner')
    await queryRunner.renameColumn('workspace', 'nodeId', 'runnerId')
    await queryRunner.renameColumn('workspace', 'prevNodeId', 'prevRunnerId')

    // Image to snapshot rename
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

    // Workspace to sandbox rename
    await queryRunner.renameTable('workspace', 'sandbox')
    await queryRunner.renameTable('workspace_usage_periods', 'sandbox_usage_periods')
    await queryRunner.renameColumn('sandbox_usage_periods', 'workspaceId', 'sandboxId')
    await queryRunner.renameColumn('organization', 'max_cpu_per_workspace', 'max_cpu_per_sandbox')
    await queryRunner.renameColumn('organization', 'max_memory_per_workspace', 'max_memory_per_sandbox')
    await queryRunner.renameColumn('organization', 'max_disk_per_workspace', 'max_disk_per_sandbox')

    // Snapshot fields
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "imageName" character varying NOT NULL DEFAULT ''`)
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "cpu" integer NOT NULL DEFAULT '1'`)
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "gpu" integer NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "mem" integer NOT NULL DEFAULT '1'`)
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "disk" integer NOT NULL DEFAULT '3'`)
    await queryRunner.query(`UPDATE "snapshot" SET "imageName" = "name"`)

    // Add hideFromUsers column
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "hideFromUsers" boolean NOT NULL DEFAULT false`)

    // Set hideFromUsers to true for general snapshots with names starting with "daytonaio/"
    await queryRunner.query(
      `UPDATE "snapshot" SET "hideFromUsers" = true WHERE "general" = true AND "name" LIKE 'daytonaio/%'`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Remove hideFromUsers column
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "hideFromUsers"`)

    // Snapshot fields
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "disk"`)
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "mem"`)
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "gpu"`)
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "cpu"`)
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "imageName"`)

    // Revert workspace to sandbox rename
    await queryRunner.renameColumn('organization', 'max_disk_per_sandbox', 'max_disk_per_workspace')
    await queryRunner.renameColumn('organization', 'max_memory_per_sandbox', 'max_memory_per_workspace')
    await queryRunner.renameColumn('organization', 'max_cpu_per_sandbox', 'max_cpu_per_workspace')
    await queryRunner.renameColumn('sandbox_usage_periods', 'sandboxId', 'workspaceId')
    await queryRunner.renameTable('sandbox_usage_periods', 'workspace_usage_periods')
    await queryRunner.renameTable('sandbox', 'workspace')

    // Revert image to snapshot rename
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

    // Revert node to runner rename
    await queryRunner.renameColumn('workspace', 'prevRunnerId', 'prevNodeId')
    await queryRunner.renameColumn('workspace', 'runnerId', 'nodeId')
    await queryRunner.renameTable('runner', 'node')
    await queryRunner.renameColumn('image_runner', 'runnerId', 'nodeId')
    await queryRunner.renameTable('image_runner', 'image_node')
    await queryRunner.renameColumn('image', 'buildRunnerId', 'buildNodeId')

    // Revert snapshot to backup rename
    await queryRunner.renameColumn('workspace', 'existingBackupImages', 'existingSnapshotImages')
    await queryRunner.renameColumn('workspace', 'backupState', 'snapshotState')
    await queryRunner.renameColumn('workspace', 'lastBackupAt', 'lastSnapshotAt')
    await queryRunner.renameColumn('workspace', 'backupImage', 'snapshotImage')
    await queryRunner.renameColumn('workspace', 'backupRegistryId', 'snapshotRegistryId')
  }
}
