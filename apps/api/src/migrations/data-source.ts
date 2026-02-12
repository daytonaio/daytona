/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DataSource, DataSourceOptions } from 'typeorm'
import { CustomNamingStrategy } from '../common/utils/naming-strategy.util'
import { join } from 'path'
import { config } from 'dotenv'

config({ path: [join(__dirname, '../../.env'), join(__dirname, '../../.env.local')] })

export const baseDataSourceOptions: DataSourceOptions = {
  type: 'postgres' as const,
  host: process.env.DB_HOST,
  port: parseInt(process.env.DB_PORT!, 10),
  username: process.env.DB_USERNAME,
  password: process.env.DB_PASSWORD,
  database: process.env.DB_DATABASE,
  synchronize: false,
  migrationsRun: false,
  logging: process.env.DB_LOGGING === 'true',
  namingStrategy: new CustomNamingStrategy(),
  entities: [join(__dirname, '../**/*.entity.ts')],
  entitySkipConstructor: true,
}

const AppDataSource = new DataSource({
  ...baseDataSourceOptions,
  migrations: [join(__dirname, '**/*-migration.{ts,js}')],
})

export default AppDataSource
