/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1768475454675 implements MigrationInterface {
  name = 'Migration1768475454675'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE INDEX "idx_region_custom" ON "region" ("organizationId") WHERE "regionType" = 'custom'`,
    )
    await queryRunner.query(
      `CREATE UNIQUE INDEX "region_sshGatewayApiKeyHash_unique" ON "region" ("sshGatewayApiKeyHash") WHERE "sshGatewayApiKeyHash" IS NOT NULL`,
    )
    await queryRunner.query(
      `CREATE UNIQUE INDEX "region_proxyApiKeyHash_unique" ON "region" ("proxyApiKeyHash") WHERE "proxyApiKeyHash" IS NOT NULL`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX "public"."region_proxyApiKeyHash_unique"`)
    await queryRunner.query(`DROP INDEX "public"."region_sshGatewayApiKeyHash_unique"`)
    await queryRunner.query(`DROP INDEX "public"."idx_region_custom"`)
  }
}
