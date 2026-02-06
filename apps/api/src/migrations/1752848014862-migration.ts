/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1752848014862 implements MigrationInterface {
  name = 'Migration1752848014862'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "warm_pool" ALTER COLUMN "createdAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "warm_pool" ALTER COLUMN "updatedAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "lastChecked" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "createdAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "updatedAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" ALTER COLUMN "startAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" ALTER COLUMN "endAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "snapshot_runner" ALTER COLUMN "createdAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "snapshot_runner" ALTER COLUMN "updatedAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "createdAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "updatedAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "lastActivityAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "lastBackupAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "build_info" ALTER COLUMN "lastUsedAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "build_info" ALTER COLUMN "createdAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "build_info" ALTER COLUMN "updatedAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "createdAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "updatedAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "createdAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "updatedAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "volume" ALTER COLUMN "createdAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "volume" ALTER COLUMN "updatedAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "organization_user" ALTER COLUMN "createdAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "organization_user" ALTER COLUMN "updatedAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(
      `ALTER TABLE "organization_invitation" ALTER COLUMN "expiresAt" TYPE TIMESTAMP WITH TIME ZONE`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_invitation" ALTER COLUMN "createdAt" TYPE TIMESTAMP WITH TIME ZONE`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_invitation" ALTER COLUMN "updatedAt" TYPE TIMESTAMP WITH TIME ZONE`,
    )
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "suspendedAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "suspendedUntil" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "createdAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "updatedAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "organization_role" ALTER COLUMN "createdAt" TYPE TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "organization_role" ALTER COLUMN "updatedAt" TYPE TIMESTAMP WITH TIME ZONE`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization_role" ALTER COLUMN "updatedAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "organization_role" ALTER COLUMN "createdAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "updatedAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "createdAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "suspendedUntil" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "suspendedAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "organization_invitation" ALTER COLUMN "updatedAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "organization_invitation" ALTER COLUMN "createdAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "organization_invitation" ALTER COLUMN "expiresAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "organization_user" ALTER COLUMN "updatedAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "organization_user" ALTER COLUMN "createdAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "volume" ALTER COLUMN "updatedAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "volume" ALTER COLUMN "createdAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "updatedAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "createdAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "updatedAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "createdAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "imageName" SET DEFAULT ''`)
    await queryRunner.query(`ALTER TABLE "build_info" ALTER COLUMN "updatedAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "build_info" ALTER COLUMN "createdAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "build_info" ALTER COLUMN "lastUsedAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "lastBackupAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "lastActivityAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "updatedAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "createdAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "snapshot_runner" ALTER COLUMN "updatedAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "snapshot_runner" ALTER COLUMN "createdAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" ALTER COLUMN "endAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" ALTER COLUMN "startAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "updatedAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "createdAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "lastChecked" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "warm_pool" ALTER COLUMN "updatedAt" TYPE TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "warm_pool" ALTER COLUMN "createdAt" TYPE TIMESTAMP`)
  }
}
