// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

declare module '*.json' {
  const value: Record<string, unknown>
  export = value
}

declare module '../package.json' {
  const value: { name: string; version: string }
  export = value
}
