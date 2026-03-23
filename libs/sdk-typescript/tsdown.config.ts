import { defineConfig } from 'tsdown'

export default defineConfig({
  entry: ['src/index.ts'],
  tsconfig: 'tsconfig.tsdown.json',
  platform: 'node',
  format: 'esm',
  noExternal: [/.*/],
  dts: {
    build: true,
  },
  fixedExtension: false,
})
