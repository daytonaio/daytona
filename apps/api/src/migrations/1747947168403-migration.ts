/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1747947168403 implements MigrationInterface {
  name = 'Migration1747947168403'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.renameColumn('image', 'buildNodeId', 'buildRunnerId')
    await queryRunner.renameTable('image_node', 'image_runner')
    await queryRunner.renameColumn('image_runner', 'nodeId', 'runnerId')
    await queryRunner.renameTable('node', 'runner')
    await queryRunner.renameColumn('workspace', 'nodeId', 'runnerId')
    await queryRunner.renameColumn('workspace', 'prevNodeId', 'prevRunnerId')
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.renameColumn('workspace', 'prevRunnerId', 'prevNodeId')
    await queryRunner.renameColumn('workspace', 'runnerId', 'nodeId')
    await queryRunner.renameTable('runner', 'node')
    await queryRunner.renameColumn('image_runner', 'runnerId', 'nodeId')
    await queryRunner.renameTable('image_runner', 'image_node')
    await queryRunner.renameColumn('image', 'buildRunnerId', 'buildNodeId')
  }
}
