/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AuthStrategyType } from '../enums/auth-strategy-type.enum'

declare global {
  namespace Express {
    interface Request {
      authStrategyType?: AuthStrategyType
    }
  }
}
