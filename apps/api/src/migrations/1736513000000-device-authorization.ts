/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner, Table, TableIndex } from 'typeorm'

export class DeviceAuthorization1736513000000 implements MigrationInterface {
  name = 'DeviceAuthorization1736513000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.createTable(
      new Table({
        name: 'device_authorization',
        columns: [
          {
            name: 'id',
            type: 'uuid',
            isPrimary: true,
            default: 'uuid_generate_v4()',
          },
          {
            name: 'device_code',
            type: 'varchar',
            length: '128',
            isUnique: true,
            isNullable: false,
          },
          {
            name: 'user_code',
            type: 'varchar',
            length: '16',
            isUnique: true,
            isNullable: false,
          },
          {
            name: 'client_id',
            type: 'varchar',
            length: '128',
            isNullable: false,
          },
          {
            name: 'scope',
            type: 'text',
            isNullable: true,
          },
          {
            name: 'status',
            type: 'varchar',
            length: '32',
            isNullable: false,
            default: "'pending'",
          },
          {
            name: 'user_id',
            type: 'uuid',
            isNullable: true,
          },
          {
            name: 'organization_id',
            type: 'uuid',
            isNullable: true,
          },
          {
            name: 'created_at',
            type: 'timestamp',
            default: 'now()',
            isNullable: false,
          },
          {
            name: 'expires_at',
            type: 'timestamp',
            isNullable: false,
          },
          {
            name: 'approved_at',
            type: 'timestamp',
            isNullable: true,
          },
          {
            name: 'last_poll_at',
            type: 'timestamp',
            isNullable: true,
          },
        ],
      }),
      true,
    )

    // Create indexes for better query performance
    await queryRunner.createIndex(
      'device_authorization',
      new TableIndex({
        name: 'idx_device_authorization_device_code',
        columnNames: ['device_code'],
      }),
    )

    await queryRunner.createIndex(
      'device_authorization',
      new TableIndex({
        name: 'idx_device_authorization_user_code',
        columnNames: ['user_code'],
      }),
    )

    await queryRunner.createIndex(
      'device_authorization',
      new TableIndex({
        name: 'idx_device_authorization_status',
        columnNames: ['status'],
      }),
    )

    await queryRunner.createIndex(
      'device_authorization',
      new TableIndex({
        name: 'idx_device_authorization_expires_at',
        columnNames: ['expires_at'],
      }),
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.dropTable('device_authorization')
  }
}
