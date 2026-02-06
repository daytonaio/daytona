/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1745393243334 implements MigrationInterface {
  name = 'Migration1745393243334'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // First, get all images with their current entrypoint values
    const images = await queryRunner.query(`SELECT id, entrypoint FROM "image" WHERE entrypoint IS NOT NULL`)

    // Rename the column to avoid data loss
    await queryRunner.query(`ALTER TABLE "image" RENAME COLUMN "entrypoint" TO "entrypoint_old"`)

    // Add the new jsonb column
    await queryRunner.query(`ALTER TABLE "image" ADD "entrypoint" text[]`)

    // Update each image to convert the string entrypoint to a JSON array
    for (const image of images) {
      const entrypointValue = image.entrypoint
      if (entrypointValue) {
        // Convert the string to a JSON array with a single element
        await queryRunner.query(`UPDATE "image" SET "entrypoint" = $1 WHERE id = $2`, [
          entrypointValue.split(' '),
          image.id,
        ])
      }
    }

    // Drop the old column
    await queryRunner.query(`ALTER TABLE "image" DROP COLUMN "entrypoint_old"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // First, get all images with their current entrypoint values
    const images = await queryRunner.query(`SELECT id, entrypoint FROM "image" WHERE entrypoint IS NOT NULL`)

    // Rename the column to avoid data loss
    await queryRunner.query(`ALTER TABLE "image" RENAME COLUMN "entrypoint" TO "entrypoint_old"`)

    // Add the new character varying column
    await queryRunner.query(`ALTER TABLE "image" ADD "entrypoint" character varying`)

    // Update each image to convert the JSON array to a string
    for (const image of images) {
      const entrypointArray = image.entrypoint_old
      if (entrypointArray && Array.isArray(entrypointArray) && entrypointArray.length > 0) {
        // Take the first element of the array as the string value
        await queryRunner.query(`UPDATE "image" SET "entrypoint" = $1 WHERE id = $2`, [entrypointArray[0], image.id])
      }
    }

    // Drop the old column
    await queryRunner.query(`ALTER TABLE "image" DROP COLUMN "entrypoint_old"`)
  }
}
