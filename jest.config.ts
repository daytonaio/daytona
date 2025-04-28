/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { getJestProjectsAsync } from '@nx/jest'

export default async () => ({
  projects: await getJestProjectsAsync(),
})
