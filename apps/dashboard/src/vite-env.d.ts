/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/// <reference types="vite/client" />

declare module '*.png' {
  const content: string
  export default content
}

interface ImportMetaEnv {
  readonly VITE_API_URL: string
  readonly VITE_CLIENT_SIDE_SANDBOX_PAGINATION?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
