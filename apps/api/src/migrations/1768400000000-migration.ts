/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1768400000000 implements MigrationInterface {
  name = 'Migration1768400000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Note: not using CONCURRENTLY + skipping transactions because of reverting issue: https://github.com/typeorm/typeorm/issues/9981

    // api_key indexes
    await queryRunner.query(`CREATE INDEX "api_key_org_user_idx" ON "api_key" ("organizationId", "userId")`)

    // sandbox indexes
    await queryRunner.query(
      `CREATE INDEX "sandbox_active_only_idx" ON "sandbox" ("id") WHERE "state" <> ALL (ARRAY['destroyed'::sandbox_state_enum, 'archived'::sandbox_state_enum])`,
    )
    await queryRunner.query(
      `CREATE INDEX "sandbox_labels_gin_full_idx" ON "sandbox" USING gin ("labels" jsonb_path_ops)`,
    )
    await queryRunner.query(`CREATE INDEX "sandbox_pending_idx" ON "sandbox" ("id") WHERE "pending" = true`)
    await queryRunner.query(
      `CREATE INDEX "sandbox_runner_state_desired_idx" ON "sandbox" ("runnerId", "state", "desiredState") WHERE "pending" = false`,
    )
    await queryRunner.query(`CREATE INDEX "sandbox_state_idx" ON "sandbox" ("state")`)
    await queryRunner.query(`CREATE INDEX "sandbox_desiredstate_idx" ON "sandbox" ("desiredState")`)
    await queryRunner.query(`CREATE INDEX "sandbox_snapshot_idx" ON "sandbox" ("snapshot")`)
    await queryRunner.query(`CREATE INDEX "sandbox_runnerid_idx" ON "sandbox" ("runnerId")`)
    await queryRunner.query(`CREATE INDEX "sandbox_runner_state_idx" ON "sandbox" ("runnerId", "state")`)
    await queryRunner.query(`CREATE INDEX "sandbox_organizationid_idx" ON "sandbox" ("organizationId")`)
    await queryRunner.query(`CREATE INDEX "sandbox_region_idx" ON "sandbox" ("region")`)
    await queryRunner.query(`CREATE INDEX "sandbox_resources_idx" ON "sandbox" ("cpu", "mem", "disk", "gpu")`)
    await queryRunner.query(`CREATE INDEX "sandbox_backupstate_idx" ON "sandbox" ("backupState")`)

    // snapshot indexes
    await queryRunner.query(`CREATE INDEX "snapshot_name_idx" ON "snapshot" ("name")`)
    await queryRunner.query(`CREATE INDEX "snapshot_state_idx" ON "snapshot" ("state")`)

    // snapshot_runner indexes
    await queryRunner.query(`CREATE INDEX "snapshot_runner_snapshotref_idx" ON "snapshot_runner" ("snapshotRef")`)
    await queryRunner.query(
      `CREATE INDEX "snapshot_runner_runnerid_snapshotref_idx" ON "snapshot_runner" ("runnerId", "snapshotRef")`,
    )
    await queryRunner.query(`CREATE INDEX "snapshot_runner_runnerid_idx" ON "snapshot_runner" ("runnerId")`)
    await queryRunner.query(`CREATE INDEX "snapshot_runner_state_idx" ON "snapshot_runner" ("state")`)

    // warm_pool indexes
    await queryRunner.query(
      `CREATE INDEX "warm_pool_find_idx" ON "warm_pool" ("snapshot", "target", "class", "cpu", "mem", "disk", "gpu", "osUser", "env")`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // warm_pool indexes
    await queryRunner.query(`DROP INDEX "public"."warm_pool_find_idx"`)

    // snapshot_runner indexes
    await queryRunner.query(`DROP INDEX "public"."snapshot_runner_state_idx"`)
    await queryRunner.query(`DROP INDEX "public"."snapshot_runner_runnerid_idx"`)
    await queryRunner.query(`DROP INDEX "public"."snapshot_runner_runnerid_snapshotref_idx"`)
    await queryRunner.query(`DROP INDEX "public"."snapshot_runner_snapshotref_idx"`)

    // snapshot indexes
    await queryRunner.query(`DROP INDEX "public"."snapshot_state_idx"`)
    await queryRunner.query(`DROP INDEX "public"."snapshot_name_idx"`)

    // sandbox indexes
    await queryRunner.query(`DROP INDEX "public"."sandbox_backupstate_idx"`)
    await queryRunner.query(`DROP INDEX "public"."sandbox_resources_idx"`)
    await queryRunner.query(`DROP INDEX "public"."sandbox_region_idx"`)
    await queryRunner.query(`DROP INDEX "public"."sandbox_organizationid_idx"`)
    await queryRunner.query(`DROP INDEX "public"."sandbox_runner_state_idx"`)
    await queryRunner.query(`DROP INDEX "public"."sandbox_runnerid_idx"`)
    await queryRunner.query(`DROP INDEX "public"."sandbox_snapshot_idx"`)
    await queryRunner.query(`DROP INDEX "public"."sandbox_desiredstate_idx"`)
    await queryRunner.query(`DROP INDEX "public"."sandbox_state_idx"`)
    await queryRunner.query(`DROP INDEX "public"."sandbox_runner_state_desired_idx"`)
    await queryRunner.query(`DROP INDEX "public"."sandbox_pending_idx"`)
    await queryRunner.query(`DROP INDEX "public"."sandbox_labels_gin_full_idx"`)
    await queryRunner.query(`DROP INDEX "public"."sandbox_active_only_idx"`)

    // api_key indexes
    await queryRunner.query(`DROP INDEX "public"."api_key_org_user_idx"`)
  }
}
