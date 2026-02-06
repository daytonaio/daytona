/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'
import { GlobalOrganizationRolesIds } from '../organization/constants/global-organization-roles.constant'

export class Migration1753100751731 implements MigrationInterface {
  name = 'Migration1753100751731'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
      UPDATE "organization_role" 
      SET "name" = 'Snapshots Admin', "description" = 'Grants admin access to snapshots in the organization'
      WHERE "id" = '${GlobalOrganizationRolesIds.SNAPSHOTS_ADMIN}'
    `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
      UPDATE "organization_role" 
      SET "name" = 'Images Admin', "description" = 'Grants admin access to images in the organization'
      WHERE "id" = '${GlobalOrganizationRolesIds.SNAPSHOTS_ADMIN}'
    `)
  }
}
