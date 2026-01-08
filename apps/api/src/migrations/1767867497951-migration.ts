/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1767867497951 implements MigrationInterface {
  name = 'Migration1767867497951'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "job" DROP CONSTRAINT "VALIDATE_JOB_TYPE"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `ALTER TABLE "job" ADD CONSTRAINT "VALIDATE_JOB_TYPE" CHECK (((type)::text = ANY ((ARRAY['CREATE_SANDBOX'::character varying, 'START_SANDBOX'::character varying, 'STOP_SANDBOX'::character varying, 'DESTROY_SANDBOX'::character varying, 'CREATE_BACKUP'::character varying, 'BUILD_SNAPSHOT'::character varying, 'PULL_SNAPSHOT'::character varying, 'REMOVE_SNAPSHOT'::character varying, 'UPDATE_SANDBOX_NETWORK_SETTINGS'::character varying])::text[])))`,
    )
  }
}
