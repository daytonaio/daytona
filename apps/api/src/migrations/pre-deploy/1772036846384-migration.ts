/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1772036846384 implements MigrationInterface {
  name = 'Migration1772036846384'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
          CREATE OR REPLACE FUNCTION prevent_finalized_audit_log_update()
          RETURNS TRIGGER AS $$
          BEGIN
            IF OLD."statusCode" IS NOT NULL THEN
              RAISE EXCEPTION 'Finalized audit logs are immutable.';
            END IF;
            RETURN NEW;
          END;
          $$ LANGUAGE plpgsql;
        `)
    await queryRunner.query(`
          CREATE TRIGGER audit_log_immutability
            BEFORE UPDATE ON "audit_log"
            FOR EACH ROW
            EXECUTE FUNCTION prevent_finalized_audit_log_update();
        `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP TRIGGER IF EXISTS audit_log_immutability ON "audit_log"`)
    await queryRunner.query(`DROP FUNCTION IF EXISTS prevent_finalized_audit_log_update`)
  }
}
