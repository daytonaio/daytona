// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { defineConfig } from 'vite'
import { nodePolyfills } from 'vite-plugin-node-polyfills'

export default defineConfig({
  build: {
    // es2022 target is required for top-level await in the browser bundle.
    // Playwright installs a recent Chromium (Chrome 94+) that supports ES2022.
    target: 'es2022',
  },
  plugins: [nodePolyfills({ globals: { Buffer: true, process: true, global: true } })],
})
