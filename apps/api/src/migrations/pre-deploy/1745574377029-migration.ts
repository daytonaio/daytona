/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1745574377029 implements MigrationInterface {
  name = 'Migration1745574377029'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TABLE "build_info" ("imageRef" character varying NOT NULL, "dockerfileContent" text, "contextHashes" text, "lastUsedAt" TIMESTAMP NOT NULL DEFAULT now(), "createdAt" TIMESTAMP NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP NOT NULL DEFAULT now(), CONSTRAINT "build_info_imageRef_pk" PRIMARY KEY ("imageRef"))`,
    )
    await queryRunner.renameColumn('image_node', 'internalImageName', 'imageRef')
    await queryRunner.query(`ALTER TABLE "image_node" DROP COLUMN "image"`)
    await queryRunner.query(`ALTER TABLE "image" ADD "buildInfoImageRef" character varying`)
    await queryRunner.query(`ALTER TABLE "workspace" ADD "buildInfoImageRef" character varying`)
    await queryRunner.query(`ALTER TYPE "public"."image_node_state_enum" RENAME TO "image_node_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."image_node_state_enum" AS ENUM('pulling_image', 'building_image', 'ready', 'error', 'removing')`,
    )
    await queryRunner.query(`ALTER TABLE "image_node" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "image_node" ALTER COLUMN "state" TYPE "public"."image_node_state_enum" USING "state"::"text"::"public"."image_node_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "image_node" ALTER COLUMN "state" SET DEFAULT 'pulling_image'`)
    await queryRunner.query(`DROP TYPE "public"."image_node_state_enum_old"`)
    await queryRunner.query(`ALTER TYPE "public"."image_state_enum" RENAME TO "image_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."image_state_enum" AS ENUM('build_pending', 'building', 'pending', 'pulling_image', 'pending_validation', 'validating', 'active', 'error', 'removing')`,
    )
    await queryRunner.query(`ALTER TABLE "image" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "image" ALTER COLUMN "state" TYPE "public"."image_state_enum" USING "state"::"text"::"public"."image_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "image" ALTER COLUMN "state" SET DEFAULT 'pending'`)
    await queryRunner.query(`DROP TYPE "public"."image_state_enum_old"`)
    await queryRunner.query(`ALTER TYPE "public"."workspace_state_enum" RENAME TO "workspace_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."workspace_state_enum" AS ENUM('creating', 'restoring', 'destroyed', 'destroying', 'started', 'stopped', 'starting', 'stopping', 'resizing', 'error', 'pending_build', 'building_image', 'unknown', 'pulling_image', 'archiving', 'archived')`,
    )
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "workspace" ALTER COLUMN "state" TYPE "public"."workspace_state_enum" USING "state"::"text"::"public"."workspace_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "state" SET DEFAULT 'unknown'`)
    await queryRunner.query(`DROP TYPE "public"."workspace_state_enum_old"`)
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "image" DROP NOT NULL`)
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "authToken" SET DEFAULT MD5(random()::text)`)
    await queryRunner.query(
      `ALTER TABLE "image" ADD CONSTRAINT "image_buildInfoImageRef_fk" FOREIGN KEY ("buildInfoImageRef") REFERENCES "build_info"("imageRef") ON DELETE NO ACTION ON UPDATE NO ACTION`,
    )
    await queryRunner.query(
      `ALTER TABLE "workspace" ADD CONSTRAINT "workspace_buildInfoImageRef_fk" FOREIGN KEY ("buildInfoImageRef") REFERENCES "build_info"("imageRef") ON DELETE NO ACTION ON UPDATE NO ACTION`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "workspace" DROP CONSTRAINT "workspace_buildInfoImageRef_fk"`)
    await queryRunner.query(`ALTER TABLE "image" DROP CONSTRAINT "image_buildInfoImageRef_fk"`)
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "authToken" SET DEFAULT md5((random()))`)
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "image" SET NOT NULL`)
    await queryRunner.query(
      `CREATE TYPE "public"."workspace_state_enum_old" AS ENUM('archived', 'archiving', 'creating', 'destroyed', 'destroying', 'error', 'pulling_image', 'resizing', 'restoring', 'started', 'starting', 'stopped', 'stopping', 'unknown')`,
    )
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "workspace" ALTER COLUMN "state" TYPE "public"."workspace_state_enum_old" USING "state"::"text"::"public"."workspace_state_enum_old"`,
    )
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "state" SET DEFAULT 'unknown'`)
    await queryRunner.query(`DROP TYPE "public"."workspace_state_enum"`)
    await queryRunner.query(`ALTER TYPE "public"."workspace_state_enum_old" RENAME TO "workspace_state_enum"`)
    await queryRunner.query(
      `CREATE TYPE "public"."image_state_enum_old" AS ENUM('active', 'error', 'pending', 'pending_validation', 'pulling_image', 'removing', 'validating')`,
    )
    await queryRunner.query(`ALTER TABLE "image" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "image" ALTER COLUMN "state" TYPE "public"."image_state_enum_old" USING "state"::"text"::"public"."image_state_enum_old"`,
    )
    await queryRunner.query(`ALTER TABLE "image" ALTER COLUMN "state" SET DEFAULT 'pending'`)
    await queryRunner.query(`DROP TYPE "public"."image_state_enum"`)
    await queryRunner.query(`ALTER TYPE "public"."image_state_enum_old" RENAME TO "image_state_enum"`)
    await queryRunner.query(
      `CREATE TYPE "public"."image_node_state_enum_old" AS ENUM('error', 'pulling_image', 'ready', 'removing')`,
    )
    await queryRunner.query(`ALTER TABLE "image_node" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "image_node" ALTER COLUMN "state" TYPE "public"."image_node_state_enum_old" USING "state"::"text"::"public"."image_node_state_enum_old"`,
    )
    await queryRunner.query(`ALTER TABLE "image_node" ALTER COLUMN "state" SET DEFAULT 'pulling_image'`)
    await queryRunner.query(`DROP TYPE "public"."image_node_state_enum"`)
    await queryRunner.query(`ALTER TYPE "public"."image_node_state_enum_old" RENAME TO "image_node_state_enum"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "buildInfoImageRef"`)
    await queryRunner.query(`ALTER TABLE "image" DROP COLUMN "buildInfoImageRef"`)
    await queryRunner.renameColumn('image_node', 'imageRef', 'internalImageName')
    await queryRunner.query(`ALTER TABLE "image_node" ADD "image" character varying NOT NULL`)
    await queryRunner.query(`DROP TABLE "build_info"`)
  }
}
