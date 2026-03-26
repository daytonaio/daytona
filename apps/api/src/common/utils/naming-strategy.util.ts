/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DefaultNamingStrategy, NamingStrategyInterface, Table } from 'typeorm'

export class CustomNamingStrategy extends DefaultNamingStrategy implements NamingStrategyInterface {
  primaryKeyName(tableOrName: Table | string, columnNames: string[]) {
    const table = tableOrName instanceof Table ? tableOrName.name : tableOrName
    const columnsSnakeCase = columnNames.join('_')
    return `${table}_${columnsSnakeCase}_pk`
  }

  foreignKeyName(tableOrName: Table | string, columnNames: string[]): string {
    const table = tableOrName instanceof Table ? tableOrName.name : tableOrName
    const columnsSnakeCase = columnNames.join('_')
    return `${table}_${columnsSnakeCase}_fk`
  }

  uniqueConstraintName(tableOrName: Table | string, columnNames: string[]): string {
    const table = tableOrName instanceof Table ? tableOrName.name : tableOrName
    const columnsSnakeCase = columnNames.join('_')
    return `${table}_${columnsSnakeCase}_unique`
  }

  indexName(tableOrName: Table | string, columnNames: string[]): string {
    const table = tableOrName instanceof Table ? tableOrName.name : tableOrName
    const columnsSnakeCase = columnNames.join('_')
    return `${table}_${columnsSnakeCase}_index`
  }
}
