/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

/**
 * Migration for updating the registry type enum
 */
export class Migration1759323052374 implements MigrationInterface {
  name = 'Migration1759323052374'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Drop default
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "registryType" DROP DEFAULT`)

    // Rename the old enum type
    await queryRunner.query(
      `ALTER TYPE "docker_registry_registrytype_enum" RENAME TO "docker_registry_registrytype_enum_old"`,
    )

    // Create the new enum type with updated values
    await queryRunner.query(
      `CREATE TYPE "docker_registry_registrytype_enum" AS ENUM('snapshot', 'source', 'backup', 'transient')`,
    )

    // Update the column to use the new enum type, mapping old values to new ones
    await queryRunner.query(`
      ALTER TABLE "docker_registry" 
      ALTER COLUMN "registryType" TYPE "docker_registry_registrytype_enum" 
      USING CASE 
        WHEN "registryType"::text = 'internal' THEN 'snapshot'::"docker_registry_registrytype_enum"
        WHEN "registryType"::text = 'organization' THEN 'source'::"docker_registry_registrytype_enum"
        WHEN "registryType"::text = 'backup' THEN 'backup'::"docker_registry_registrytype_enum"
        WHEN "registryType"::text = 'transient' THEN 'transient'::"docker_registry_registrytype_enum"
      END
    `)

    // Drop the old enum type
    await queryRunner.query(`DROP TYPE "docker_registry_registrytype_enum_old"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Rename the current enum type
    await queryRunner.query(
      `ALTER TYPE "docker_registry_registrytype_enum" RENAME TO "docker_registry_registrytype_enum_old"`,
    )

    // Recreate the old enum type
    await queryRunner.query(
      `CREATE TYPE "docker_registry_registrytype_enum" AS ENUM('internal', 'organization', 'transient', 'backup')`,
    )

    // Revert the column to use the old enum type, mapping new values back to old ones
    await queryRunner.query(`
      ALTER TABLE "docker_registry" 
      ALTER COLUMN "registryType" TYPE "docker_registry_registrytype_enum" 
      USING CASE 
        WHEN "registryType"::text = 'snapshot' THEN 'internal'::"docker_registry_registrytype_enum"
        WHEN "registryType"::text = 'source' THEN 'organization'::"docker_registry_registrytype_enum"
        WHEN "registryType"::text = 'backup' THEN 'backup'::"docker_registry_registrytype_enum"
        WHEN "registryType"::text = 'transient' THEN 'transient'::"docker_registry_registrytype_enum"
      END
    `)

    // Drop the new enum type
    await queryRunner.query(`DROP TYPE "docker_registry_registrytype_enum_old"`)

    // Revert registry type default
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "registryType" SET DEFAULT 'internal'`)
  }
}
