/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1744114341077 implements MigrationInterface {
  name = 'Migration1744114341077'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment" DROP CONSTRAINT "organization_role_assignment_roleId_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment_invitation" DROP CONSTRAINT "organization_role_assignment_invitation_roleId_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "workspace" ADD "authToken" character varying NOT NULL DEFAULT MD5(random()::text)`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment" ADD CONSTRAINT "organization_role_assignment_roleId_fk" FOREIGN KEY ("roleId") REFERENCES "organization_role"("id") ON DELETE NO ACTION ON UPDATE NO ACTION`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment_invitation" ADD CONSTRAINT "organization_role_assignment_invitation_roleId_fk" FOREIGN KEY ("roleId") REFERENCES "organization_role"("id") ON DELETE NO ACTION ON UPDATE NO ACTION`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment_invitation" DROP CONSTRAINT "organization_role_assignment_invitation_roleId_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment" DROP CONSTRAINT "organization_role_assignment_roleId_fk"`,
    )
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "authToken"`)
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment_invitation" ADD CONSTRAINT "organization_role_assignment_invitation_roleId_fk" FOREIGN KEY ("roleId") REFERENCES "organization_role"("id") ON DELETE CASCADE ON UPDATE CASCADE`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment" ADD CONSTRAINT "organization_role_assignment_roleId_fk" FOREIGN KEY ("roleId") REFERENCES "organization_role"("id") ON DELETE CASCADE ON UPDATE CASCADE`,
    )
  }
}
