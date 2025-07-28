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
}
