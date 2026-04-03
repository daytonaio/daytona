/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export class EmailUtils {
  static normalize(email: string): string {
    return email.toLowerCase().trim()
  }

  static areEqual(email1: string, email2: string): boolean {
    return this.normalize(email1) === this.normalize(email2)
  }

  // Inserts zero-width spaces (U+200B) to break URL patterns and prevent email client auto-linking.
  static sanitizeForDisplay(text: string): string {
    return text.replace(/:\/\//g, ':\u200B/\u200B/').replace(/(\w)\.(\w)/g, '$1\u200B.\u200B$2')
  }
}
