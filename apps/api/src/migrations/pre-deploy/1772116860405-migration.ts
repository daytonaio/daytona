/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

/**
 * Reconciliation migration that resolves drift between the database schema and the current entity definitions in the codebase.
 */
export class Migration1772116860405 implements MigrationInterface {
  name = 'Migration1772116860405'

  public async up(queryRunner: QueryRunner): Promise<void> {
    /**
     * Constraint renames due to conflict with custom naming strategy.
     */
    await queryRunner.query(
      `ALTER TABLE "snapshot_region" RENAME CONSTRAINT "FK_snapshot_region_snapshot" TO "snapshot_region_snapshotId_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "snapshot_region" RENAME CONSTRAINT "FK_snapshot_region_region" TO "snapshot_region_regionId_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "snapshot" RENAME CONSTRAINT "public.snapshot_buildInfoImageRef_fk" TO "snapshot_buildInfoSnapshotRef_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "sandbox" RENAME CONSTRAINT "public.sandbox_buildInfoSnapshotRef_fk" TO "sandbox_buildInfoSnapshotRef_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "snapshot" RENAME CONSTRAINT "image_organizationId_name_unique" TO "snapshot_organizationId_name_unique"`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" RENAME CONSTRAINT "public.sandbox_id_pk" TO "sandbox_id_pk"`)

    /**
     * Index reconciliation due to missing partial sandbox index in migrations.
     */
    await queryRunner.query(
      `CREATE INDEX IF NOT EXISTS "sandbox_active_only_idx" ON "sandbox" ("id") WHERE (state <> ALL (ARRAY['destroyed'::sandbox_state_enum, 'archived'::sandbox_state_enum]))`,
    )

    /**
     * The intial migration incorrectly added the column, it was never actually added to the entity definition.
     */
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "sshPass"`)

    /** Add missing column defaults for runner that the entity defines but the original migration omitted. */
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "cpu" SET DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "memoryGiB" SET DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "diskGiB" SET DEFAULT '0'`)

    /**
     * Recreate roleId FKs with NO ACTION instead of CASCADE to match the entity definition.
     * Deleting a role should fail if it is still assigned, not silently cascade.
     */
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment_invitation" DROP CONSTRAINT "organization_role_assignment_invitation_roleId_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment" DROP CONSTRAINT "organization_role_assignment_roleId_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment_invitation" ADD CONSTRAINT "organization_role_assignment_invitation_roleId_fk" FOREIGN KEY ("roleId") REFERENCES "organization_role"("id") ON DELETE NO ACTION ON UPDATE NO ACTION`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment" ADD CONSTRAINT "organization_role_assignment_roleId_fk" FOREIGN KEY ("roleId") REFERENCES "organization_role"("id") ON DELETE NO ACTION ON UPDATE NO ACTION`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment" DROP CONSTRAINT "organization_role_assignment_roleId_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment_invitation" DROP CONSTRAINT "organization_role_assignment_invitation_roleId_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment" ADD CONSTRAINT "organization_role_assignment_roleId_fk" FOREIGN KEY ("roleId") REFERENCES "organization_role"("id") ON DELETE CASCADE ON UPDATE CASCADE`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment_invitation" ADD CONSTRAINT "organization_role_assignment_invitation_roleId_fk" FOREIGN KEY ("roleId") REFERENCES "organization_role"("id") ON DELETE CASCADE ON UPDATE CASCADE`,
    )

    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "diskGiB" DROP DEFAULT`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "memoryGiB" DROP DEFAULT`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "cpu" DROP DEFAULT`)

    await queryRunner.query(
      `ALTER TABLE "sandbox" ADD "sshPass" character varying(32) NOT NULL DEFAULT replace((uuid_generate_v4()), '-', '')`,
    )

    await queryRunner.query(`DROP INDEX IF EXISTS "sandbox_active_only_idx"`)

    await queryRunner.query(`ALTER TABLE "sandbox" RENAME CONSTRAINT "sandbox_id_pk" TO "public.sandbox_id_pk"`)

    await queryRunner.query(
      `ALTER TABLE "snapshot" RENAME CONSTRAINT "snapshot_organizationId_name_unique" TO "image_organizationId_name_unique"`,
    )

    await queryRunner.query(
      `ALTER TABLE "sandbox" RENAME CONSTRAINT "sandbox_buildInfoSnapshotRef_fk" TO "public.sandbox_buildInfoSnapshotRef_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "snapshot" RENAME CONSTRAINT "snapshot_buildInfoSnapshotRef_fk" TO "public.snapshot_buildInfoImageRef_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "snapshot_region" RENAME CONSTRAINT "snapshot_region_regionId_fk" TO "FK_snapshot_region_region"`,
    )
    await queryRunner.query(
      `ALTER TABLE "snapshot_region" RENAME CONSTRAINT "snapshot_region_snapshotId_fk" TO "FK_snapshot_region_snapshot"`,
    )
  }
}
