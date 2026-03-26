/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AccountProvider } from '../enums/account-provider.enum'

export const ACCOUNT_PROVIDER_DISPLAY_NAME: Record<AccountProvider, string> = {
  [AccountProvider.GOOGLE]: 'Google',
  [AccountProvider.GITHUB]: 'GitHub',
}
