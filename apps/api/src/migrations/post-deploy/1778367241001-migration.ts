/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

/**
 * Seeds the `python-default` general session template (and its backing snapshot row).
 *
 * This is post-deploy so the new SessionModule routes are live before any seed data lands —
 * the API would otherwise see a template referencing a snapshot row whose `state=ACTIVE` was
 * set by this migration but whose `imageName` is still the placeholder. The dashboard /
 * runner-side validation lives outside the migration; the placeholder here gets overridden by
 * the real `daytonaio/session-runtime:python-default-<version>` image once it's built and
 * published. Until then this seed makes the tables non-empty so list-templates returns the
 * canonical entry without admins having to hand-curl a row in.
 */
export class Migration1778367241001 implements MigrationInterface {
  name = 'Migration1778367241001'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
      DO $$
      DECLARE snap_id uuid;
      BEGIN
        SELECT id INTO snap_id FROM "snapshot" WHERE "name" = 'python-default' AND "general" = true LIMIT 1;
        IF snap_id IS NULL THEN
          INSERT INTO "snapshot" ("name", "general", "imageName", "state", "cpu", "mem", "disk", "gpu", "entrypoint")
          VALUES (
            'python-default',
            true,
            'daytonaio/session-runtime:python-default-placeholder',
            'pending',
            2,
            2,
            5,
            0,
            ARRAY['/opt/daytona/session-daemon']::text[]
          )
          RETURNING id INTO snap_id;
        END IF;

        IF NOT EXISTS (SELECT 1 FROM "session_template" WHERE "name" = 'python-default' AND "general" = true) THEN
          INSERT INTO "session_template" ("name", "general", "languages", "snapshotId", "description", "packages")
          VALUES (
            'python-default',
            true,
            ARRAY['python', 'typescript', 'bash']::text[],
            snap_id,
            'Default general-purpose session template — Python 3.11 + Node 22 with curated package catalogs.',
            ARRAY['numpy', 'pandas', 'matplotlib', 'openai', 'requests', 'zod', 'lodash-es', '@anthropic-ai/sdk']::text[]
          );
        END IF;
      END
      $$;
    `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DELETE FROM "session_template" WHERE "name" = 'python-default' AND "general" = true`)
    // Leave the snapshot row in place — admins may have customized cpu/mem/imageName after the fact.
  }
}
