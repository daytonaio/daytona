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
  Pylon?: {
    (command: 'show' | 'hide'): void
    (command: 'onShow' | 'onHide', callback: (() => void) | null): void
    (command: 'onChangeUnreadMessagesCount', callback: ((count: number) => void) | null): void
  }
}
