/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1748167133281 implements MigrationInterface {
  name = 'Migration1748167133281'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.renameTable('workspace', 'sandbox')

    await queryRunner.renameTable('workspace_usage_periods', 'sandbox_usage_periods')
    await queryRunner.renameColumn('sandbox_usage_periods', 'workspaceId', 'sandboxId')

    await queryRunner.renameColumn('organization', 'max_cpu_per_workspace', 'max_cpu_per_sandbox')
    await queryRunner.renameColumn('organization', 'max_memory_per_workspace', 'max_memory_per_sandbox')
    await queryRunner.renameColumn('organization', 'max_disk_per_workspace', 'max_disk_per_sandbox')
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.renameColumn('organization', 'max_disk_per_sandbox', 'max_disk_per_workspace')
    await queryRunner.renameColumn('organization', 'max_memory_per_sandbox', 'max_memory_per_workspace')
    await queryRunner.renameColumn('organization', 'max_cpu_per_sandbox', 'max_cpu_per_workspace')

    await queryRunner.renameColumn('sandbox_usage_periods', 'sandboxId', 'workspaceId')
    await queryRunner.renameTable('sandbox_usage_periods', 'workspace_usage_periods')

    await queryRunner.renameTable('sandbox', 'workspace')
  }
}
