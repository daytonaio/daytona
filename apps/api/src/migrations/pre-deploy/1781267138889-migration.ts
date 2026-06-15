/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'
import { GlobalOrganizationRolesIds } from '../../organization/constants/global-organization-roles.constant'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'

export class Migration1781267138889 implements MigrationInterface {
  name = 'Migration1781267138889'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
      UPDATE "organization_role"
      SET "permissions" = ARRAY[
        '${OrganizationResourcePermission.WRITE_REGISTRIES}',
        '${OrganizationResourcePermission.DELETE_REGISTRIES}',
        '${OrganizationResourcePermission.WRITE_SNAPSHOTS}',
        '${OrganizationResourcePermission.DELETE_SNAPSHOTS}',
        '${OrganizationResourcePermission.WRITE_SANDBOXES}',
        '${OrganizationResourcePermission.DELETE_SANDBOXES}',
        '${OrganizationResourcePermission.READ_VOLUMES}',
        '${OrganizationResourcePermission.WRITE_VOLUMES}',
        '${OrganizationResourcePermission.DELETE_VOLUMES}',
        '${OrganizationResourcePermission.WRITE_REGIONS}',
        '${OrganizationResourcePermission.DELETE_REGIONS}',
        '${OrganizationResourcePermission.READ_RUNNERS}',
        '${OrganizationResourcePermission.WRITE_RUNNERS}',
        '${OrganizationResourcePermission.DELETE_RUNNERS}',
        '${OrganizationResourcePermission.READ_AUDIT_LOGS}'
      ]::organization_role_permissions_enum[]
      WHERE "id" = '${GlobalOrganizationRolesIds.SUPER_ADMIN}'
    `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
      UPDATE "organization_role"
      SET "permissions" = ARRAY[
        '${OrganizationResourcePermission.WRITE_REGISTRIES}',
        '${OrganizationResourcePermission.DELETE_REGISTRIES}',
        '${OrganizationResourcePermission.WRITE_SNAPSHOTS}',
        '${OrganizationResourcePermission.DELETE_SNAPSHOTS}',
        '${OrganizationResourcePermission.WRITE_SANDBOXES}',
        '${OrganizationResourcePermission.DELETE_SANDBOXES}'
      ]::organization_role_permissions_enum[]
      WHERE "id" = '${GlobalOrganizationRolesIds.SUPER_ADMIN}'
    `)
  }
}
