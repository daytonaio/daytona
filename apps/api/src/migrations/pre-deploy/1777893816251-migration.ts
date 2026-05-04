/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1777893816251 implements MigrationInterface {
  name = 'Migration1777893816251'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Add new typed otelConfig column
    await queryRunner.query(`ALTER TABLE "organization" ADD "otelConfig" jsonb`)

    // Backfill from the existing experimentalConfig.otel jsonb path
    await queryRunner.query(`
      UPDATE "organization"
      SET "otelConfig" = "experimentalConfig"->'otel'
      WHERE "experimentalConfig" IS NOT NULL
        AND "experimentalConfig" ? 'otel'
        AND "experimentalConfig"->'otel'->>'endpoint' IS NOT NULL
        AND "experimentalConfig"->'otel'->>'endpoint' != ''
    `)

    // Bidirectional sync trigger so the rolling deploy is safe:
    //   Old API writes to experimentalConfig.otel -> mirror to otelConfig
    //   New API writes to otelConfig             -> mirror to experimentalConfig.otel
    // INSERT and UPDATE need separate paths because OLD is undefined on INSERT.
    await queryRunner.query(`
      CREATE OR REPLACE FUNCTION sync_organization_otel_config()
      RETURNS TRIGGER AS $$
      BEGIN
        IF TG_OP = 'INSERT' THEN
          IF NEW."otelConfig" IS NOT NULL THEN
            NEW."experimentalConfig" := jsonb_set(
              COALESCE(NEW."experimentalConfig", '{}'::jsonb), '{otel}', NEW."otelConfig"
            );
          ELSIF NEW."experimentalConfig" IS NOT NULL AND NEW."experimentalConfig" ? 'otel' THEN
            NEW."otelConfig" := NEW."experimentalConfig" -> 'otel';
          END IF;
          RETURN NEW;
        END IF;

        -- TG_OP = 'UPDATE'
        IF NEW."otelConfig" IS DISTINCT FROM OLD."otelConfig" THEN
          IF NEW."otelConfig" IS NULL THEN
            NEW."experimentalConfig" := COALESCE(NEW."experimentalConfig", '{}'::jsonb) - 'otel';
          ELSE
            NEW."experimentalConfig" := jsonb_set(
              COALESCE(NEW."experimentalConfig", '{}'::jsonb), '{otel}', NEW."otelConfig"
            );
          END IF;
        ELSIF NEW."experimentalConfig" IS DISTINCT FROM OLD."experimentalConfig" THEN
          NEW."otelConfig" := NEW."experimentalConfig" -> 'otel';
        END IF;
        RETURN NEW;
      END;
      $$ LANGUAGE plpgsql;
    `)

    await queryRunner.query(`
      CREATE TRIGGER organization_otel_config_sync
      BEFORE INSERT OR UPDATE ON "organization"
      FOR EACH ROW EXECUTE FUNCTION sync_organization_otel_config();
    `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP TRIGGER IF EXISTS organization_otel_config_sync ON "organization"`)
    await queryRunner.query(`DROP FUNCTION IF EXISTS sync_organization_otel_config()`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "otelConfig"`)
  }
}
