/// <reference types='vitest' />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import checker from 'vite-plugin-checker'
import { nxViteTsPaths } from '@nx/vite/plugins/nx-tsconfig-paths.plugin'
import { nxCopyAssetsPlugin } from '@nx/vite/plugins/nx-copy-assets.plugin'
import path from 'path'

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
    nxViteTsPaths(),
    nxCopyAssetsPlugin(['*.md']),
    // enforce typechecking for build mode
    mode.command === 'build' &&
      checker({
        typescript: {
          tsconfigPath: './tsconfig.app.json',
        },
      }),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'), // Make sure this points to dashboard's src
    },
  },
  // Uncomment this if you are using workers.
  // worker: {
  //  plugins: [ nxViteTsPaths() ],
  // },
  build: {
    outDir: '../../dist/apps/dashboard',
    emptyOutDir: true,
    reportCompressedSize: true,
    commonjsOptions: {
      transformMixedEsModules: true,
    },
  },
}))
