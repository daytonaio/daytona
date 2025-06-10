/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

interface Window {
  pylon?: {
    chat_settings: {
      app_id: string
      email: string
      name: string
      avatar_url?: string
      email_hash?: string
    }
  }
}
