/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1777893816252 implements MigrationInterface {
  name = 'Migration1777893816252'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // New API has fully cut over — drop the bidirectional sync trigger.
    await queryRunner.query(`DROP TRIGGER IF EXISTS organization_otel_config_sync ON "organization"`)
    await queryRunner.query(`DROP FUNCTION IF EXISTS sync_organization_otel_config()`)

    // Strip the otel key from experimentalConfig now that it lives in its own column.
    // The experimentalConfig column itself stays — it's still used for future experimental features.
    await queryRunner.query(`
      UPDATE "organization"
      SET "experimentalConfig" = "experimentalConfig" - 'otel'
      WHERE "experimentalConfig" IS NOT NULL
        AND "experimentalConfig" ? 'otel'
    `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Restore otel key in experimentalConfig from otelConfig column
    await queryRunner.query(`
      UPDATE "organization"
      SET "experimentalConfig" = jsonb_set(
        COALESCE("experimentalConfig", '{}'::jsonb), '{otel}', "otelConfig"
      )
      WHERE "otelConfig" IS NOT NULL
    `)

    // Reinstall the sync trigger (must mirror pre-deploy: handle INSERT and UPDATE separately
    // because OLD is undefined on INSERT).
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
}
