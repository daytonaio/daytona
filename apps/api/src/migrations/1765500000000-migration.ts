/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1765500000000 implements MigrationInterface {
  name = 'Migration1765500000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Create the device_auth_status enum type
    await queryRunner.query(`
      CREATE TYPE "device_auth_status_enum" AS ENUM ('pending', 'approved', 'denied', 'expired')
    `)

    // Create the device_authorization_request table
    await queryRunner.query(`
      CREATE TABLE "device_authorization_request" (
        "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
        "deviceCode" character varying NOT NULL,
        "userCode" character varying NOT NULL,
        "clientId" character varying NOT NULL,
        "scope" character varying,
        "status" "device_auth_status_enum" NOT NULL DEFAULT 'pending',
        "userId" character varying,
        "organizationId" character varying,
        "accessToken" character varying,
        "createdAt" TIMESTAMP NOT NULL DEFAULT now(),
        "expiresAt" TIMESTAMP NOT NULL,
        "approvedAt" TIMESTAMP,
        "lastPolledAt" TIMESTAMP,
        CONSTRAINT "UQ_device_authorization_request_deviceCode" UNIQUE ("deviceCode"),
        CONSTRAINT "UQ_device_authorization_request_userCode" UNIQUE ("userCode"),
        CONSTRAINT "PK_device_authorization_request" PRIMARY KEY ("id")
      )
    `)

    // Create indexes for better query performance
    await queryRunner.query(`
      CREATE INDEX "IDX_device_authorization_request_deviceCode" ON "device_authorization_request" ("deviceCode")
    `)
    await queryRunner.query(`
      CREATE INDEX "IDX_device_authorization_request_userCode" ON "device_authorization_request" ("userCode")
    `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX "IDX_device_authorization_request_userCode"`)
    await queryRunner.query(`DROP INDEX "IDX_device_authorization_request_deviceCode"`)
    await queryRunner.query(`DROP TABLE "device_authorization_request"`)
    await queryRunner.query(`DROP TYPE "device_auth_status_enum"`)
  }
}
