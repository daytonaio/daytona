/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { fromJson, type Message } from '@bufbuild/protobuf'
import type { GenMessage } from '@bufbuild/protobuf/codegenv2'
import * as protoExports from '@daytona/runner-proto'

type AnySchema = GenMessage<Message>

const registry = new Map<string, AnySchema>()

for (const value of Object.values(protoExports)) {
  if (value && typeof value === 'object' && 'typeName' in value) {
    registry.set((value as AnySchema).typeName, value as AnySchema)
  }
}

function toPlainObject(value: Record<string, any>): Record<string, any> {
  const result: Record<string, any> = {}
  for (const [k, v] of Object.entries(value)) {
    if (k === '$typeName' || k === '$unknown') continue
    result[k] = convertValue(v)
  }
  return result
}

function convertValue(value: any): any {
  if (typeof value === 'bigint') return Number(value)
  if (Array.isArray(value)) return value.map(convertValue)
  if (value && typeof value === 'object') return toPlainObject(value)
  return value
}

export function parseProto(typeName: string, data: Record<string, any>): Record<string, any> | null {
  const schema = registry.get(typeName)
  if (!schema) return null
  return toPlainObject(fromJson(schema, data))
}
