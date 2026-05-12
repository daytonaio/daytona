// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { vitePlugin as remix } from '@remix-run/dev'
import { defineConfig } from 'vite'
import tsconfigPaths from 'vite-tsconfig-paths'

export default defineConfig({
  plugins: [remix(), tsconfigPaths()],
})