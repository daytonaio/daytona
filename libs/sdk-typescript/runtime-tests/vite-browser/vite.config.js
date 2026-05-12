// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { defineConfig } from 'vite'
import { nodePolyfills } from 'vite-plugin-node-polyfills'

export default defineConfig({
  plugins: [nodePolyfills({ globals: { Buffer: true, process: true, global: true } })],
})