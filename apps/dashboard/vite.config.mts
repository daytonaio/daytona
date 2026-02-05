/// <reference types='vitest' />
import { nxCopyAssetsPlugin } from '@nx/vite/plugins/nx-copy-assets.plugin'
import { nxViteTsPaths } from '@nx/vite/plugins/nx-tsconfig-paths.plugin'
import react from '@vitejs/plugin-react'
import fs from 'fs'
import path from 'path'
import { defineConfig } from 'vite'
import { analyzer } from 'vite-bundle-analyzer'
import checker from 'vite-plugin-checker'
import { nodePolyfills } from 'vite-plugin-node-polyfills'

const outDir = '../../dist/apps/dashboard'

export default defineConfig((mode) => ({
  root: __dirname,
  cacheDir: '../../node_modules/.vite/apps/dashboard',
  server: {
    port: 3000,
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: 'http://localhost:3001',
        ws: true,
        changeOrigin: true,
        rewriteWsOrigin: true,
      },
    },
  },
  plugins: [
    react(),
    // Required for @daytonaio/sdk
    nodePolyfills({
      globals: { global: true, process: true, Buffer: true },
      overrides: {
        path: 'path-browserify-win32',
      },
      protocolImports: true,
    }),
    nxViteTsPaths(),
    nxCopyAssetsPlugin(['*.md']),
    // enforce typechecking for build mode
    mode.command === 'build' &&
      checker({
        typescript: {
          tsconfigPath: './tsconfig.app.json',
        },
      }),

    {
      name: 'exclude-msw',
      apply: 'build',
      writeBundle() {
        if (mode.mode === 'production') {
          const mswPath = path.resolve(__dirname, outDir, 'mockServiceWorker.js')

          if (fs.existsSync(mswPath)) {
            fs.rmSync(mswPath)
            console.log('Removed mockServiceWorker.js from production build.')
          }
        }
      },
    },
    analyzer({
      openAnalyzer: false,
      analyzerPort: 4000,
      enabled: mode.mode === 'analyze',
    }),
  ],
  resolve: {
    alias: [
      // Resolve @daytonaio/sdk to the built distribution
      {
        find: '@daytonaio/sdk',
        replacement: path.resolve(__dirname, '../../libs/sdk-typescript/src'),
      },
      // Target @ but not @daytonaio,
      {
        // find: /^@(?!daytonaio)/,
        find: '@',
        replacement: path.resolve(__dirname, './src'), // Make sure this points to dashboard's src
      },
    ],
  },
  // Uncomment this if you are using workers.
  // worker: {
  //  plugins: [ nxViteTsPaths() ],
  // },
  optimizeDeps: {
    exclude: ['tar'],
  },
  build: {
    outDir,
    emptyOutDir: true,
    reportCompressedSize: true,
    commonjsOptions: {
      transformMixedEsModules: true,
    },
    // we'd ideally polyfill it but until https://github.com/davidmyersdev/vite-plugin-node-polyfills/issues/118 gets resolved we can just exclude it
    rollupOptions: {
      external: ['tar'],
    },
  },
}))
