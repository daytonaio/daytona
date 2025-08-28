/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1755464957487 implements MigrationInterface {
  name = 'Migration1755464957487'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TABLE "ssh_access" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "sandboxId" character varying NOT NULL, "token" text NOT NULL, "expiresAt" TIMESTAMP NOT NULL, "createdAt" TIMESTAMP NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP NOT NULL DEFAULT now(), CONSTRAINT "ssh_access_id_pk" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(
      `ALTER TABLE "ssh_access" ADD CONSTRAINT "ssh_access_sandboxId_fk" FOREIGN KEY ("sandboxId") REFERENCES "sandbox"("id") ON DELETE CASCADE ON UPDATE NO ACTION`,
    )
    await queryRunner.query(
      `ALTER TABLE "sandbox" ADD "sshPass" character varying(32) NOT NULL DEFAULT REPLACE(uuid_generate_v4()::text, '-', '')`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "sshPass"`)
    await queryRunner.query(`ALTER TABLE "ssh_access" DROP CONSTRAINT "ssh_access_sandboxId_fk"`)
    await queryRunner.query(`DROP TABLE "ssh_access"`)
  }
}
