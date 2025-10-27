/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1761120409046 implements MigrationInterface {
  name = 'Migration1761120409046'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.renameColumn('docker_registry', 'isDefault', 'isActive')
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.renameColumn('docker_registry', 'isActive', 'isDefault')
  }
}
