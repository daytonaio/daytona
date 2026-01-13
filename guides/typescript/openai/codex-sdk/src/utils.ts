/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

export const ESC = '\u001b'
export const BOLD = ESC + '[1m'
export const ITALIC = ESC + '[3m'
export const DIM = ESC + '[2m'
export const RESET = ESC + '[0m'

export function renderMarkdown(text: string): string {
  return text
    .replace(/\*\*(.+?)\*\*/g, `${BOLD}$1${RESET}`) // **bold**
    .replace(/(?<!\*)\*([^*\n]+?)\*(?!\*)/g, `${ITALIC}$1${RESET}`) // *italic*
    .replace(/`([^`]+?)`/g, `${DIM}$1${RESET}`) // `code`
}
