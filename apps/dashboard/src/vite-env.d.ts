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
  readonly VITE_OIDC_DOMAIN: string
  readonly VITE_OIDC_CLIENT_ID: string
  readonly VITE_OIDC_AUDIENCE: string
  readonly VITE_API_URL: string
  readonly VITE_BILLING_API_URL: string | undefined
  readonly VITE_POSTHOG_KEY: string | undefined
  readonly VITE_POSTHOG_HOST: string | undefined
  readonly VITE_LINKED_ACCOUNTS_ENABLED: string | undefined
  readonly VITE_ANNOUNCEMENT_BANNER_TEXT: string | undefined
  readonly VITE_ANNOUNCEMENT_BANNER_LEARN_MORE_URL: string | undefined
  readonly VITE_PYLON_APP_ID: string | undefined
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
