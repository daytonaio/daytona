/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'
import { GlobalOrganizationRolesIds } from '../organization/constants/global-organization-roles.constant'
import { OrganizationResourcePermission } from '../organization/enums/organization-resource-permission.enum'

export class Migration1741088883002 implements MigrationInterface {
  name = 'Migration1741088883002'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
      INSERT INTO "organization_role" 
        ("id", "name", "description", "permissions", "isGlobal")
      VALUES 
        (
          '${GlobalOrganizationRolesIds.DEVELOPER}',
          'Developer', 
          'Grants the ability to create sandboxes and keys in the organization', 
          ARRAY[
            '${OrganizationResourcePermission.WRITE_SANDBOXES}'
          ]::organization_role_permissions_enum[],
          TRUE
        )
    `)

    await queryRunner.query(`
      INSERT INTO "organization_role" 
        ("id", "name", "description", "permissions", "isGlobal")
      VALUES 
        (
          '${GlobalOrganizationRolesIds.SANDBOXES_ADMIN}',
          'Sandboxes Admin', 
          'Grants admin access to sandboxes in the organization', 
          ARRAY[
            '${OrganizationResourcePermission.WRITE_SANDBOXES}',
            '${OrganizationResourcePermission.DELETE_SANDBOXES}'
          ]::organization_role_permissions_enum[],
          TRUE
        )
    `)

    await queryRunner.query(`
      INSERT INTO "organization_role" 
        ("id", "name", "description", "permissions", "isGlobal")
      VALUES 
        (
          '${GlobalOrganizationRolesIds.SNAPSHOTS_ADMIN}',
          'Images Admin', 
          'Grants admin access to images in the organization', 
          ARRAY[
            'write:images',
            'delete:images'
          ]::organization_role_permissions_enum[],
          TRUE
        )
    `)

    await queryRunner.query(`
      INSERT INTO "organization_role" 
        ("id", "name", "description", "permissions", "isGlobal")
      VALUES 
        (
          '${GlobalOrganizationRolesIds.REGISTRIES_ADMIN}',
          'Registries Admin', 
          'Grants admin access to registries in the organization', 
          ARRAY[
            '${OrganizationResourcePermission.WRITE_REGISTRIES}',
            '${OrganizationResourcePermission.DELETE_REGISTRIES}'
          ]::organization_role_permissions_enum[],
          TRUE
        )
    `)

    await queryRunner.query(`
      INSERT INTO "organization_role" 
        ("id", "name", "description", "permissions", "isGlobal")
      VALUES 
        (
          '${GlobalOrganizationRolesIds.SUPER_ADMIN}',
          'Super Admin', 
          'Grants full access to all resources in the organization', 
          ARRAY[
            '${OrganizationResourcePermission.WRITE_REGISTRIES}',
            '${OrganizationResourcePermission.DELETE_REGISTRIES}',
            'write:images',
            'delete:images',
            '${OrganizationResourcePermission.WRITE_SANDBOXES}',
            '${OrganizationResourcePermission.DELETE_SANDBOXES}'
          ]::organization_role_permissions_enum[],
          TRUE
        )
    `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DELETE FROM "organization_role" WHERE "isGlobal" = TRUE`)
  }
}
