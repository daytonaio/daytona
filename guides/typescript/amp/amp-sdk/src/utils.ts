/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

const ESC = '\u001b'
const BOLD = ESC + '[1m'
const ITALIC = ESC + '[3m'
const DIM = ESC + '[2m'
const RESET = ESC + '[0m'

/** Basic markdown to ANSI (same as letta-code): **bold**, *italic*, `code` */
export function renderMarkdown(text: string): string {
  return text
    .replace(/\*\*(.+?)\*\*/g, `${BOLD}$1${RESET}`)
    .replace(/(?<!\*)\*([^*\n]+?)\*(?!\*)/g, `${ITALIC}$1${RESET}`)
    .replace(/`([^`]+?)`/g, `${DIM}$1${RESET}`)
}
