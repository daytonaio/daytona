import nx from '@nx/eslint-plugin'
import react from 'eslint-plugin-react'
import reactHooks from 'eslint-plugin-react-hooks'

/** @type {import('eslint').Linter.Config[]} */
export default [
  ...nx.configs['flat/base'],
  ...nx.configs['flat/typescript'],
  ...nx.configs['flat/javascript'],
  {
    ignores: [
      '**/dist',
      '**/vite.config.*.timestamp*',
      '**/vitest.config.*.timestamp*',
      'apps/docs/**',
      'libs/*api-client*/**',
    ],
  },
  {
    files: ['**/*.ts', '**/*.tsx', '**/*.js', '**/*.jsx'],
    plugins: {
      react,
      'react-hooks': reactHooks,
    },
    rules: {
      '@nx/enforce-module-boundaries': [
        'error',
        {
          enforceBuildableLibDependency: true,
          allow: ['^.*/eslint(\\.base)?\\.config\\.[cm]?js$'],
          depConstraints: [
            {
              sourceTag: '*',
              onlyDependOnLibsWithTags: ['*'],
            },
          ],
        },
      ],
    },
  },
  {
    files: ['**/*.ts', '**/*.tsx', '**/*.cts', '**/*.mts', '**/*.js', '**/*.jsx', '**/*.cjs', '**/*.mjs'],
    // Override or add rules here
    rules: {
      '@typescript-eslint/interface-name-prefix': 'off',
      '@typescript-eslint/explicit-function-return-type': 'off',
      '@typescript-eslint/explicit-module-boundary-types': 'off',
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-useless-escape': 'off',
    },
  },
  {
    files: ['src/migrations/**/*.ts'],
    rules: {
      quotes: 'off',
    },
  },
  {
    // The SDK runtime-test fixtures intentionally import from '@daytona/sdk'
    // (the packed published package) instead of the workspace source — that's
    // the whole point of the tests. Disable the enforce-module-boundaries
    // auto-fix that rewrites those imports to relative source paths.
    files: ['libs/sdk-typescript/runtime-tests/**/*.{ts,tsx,js,jsx,mjs,cjs}'],
    rules: {
      '@nx/enforce-module-boundaries': 'off',
    },
  },
  {
    // pi-extension consumes '@daytona/sdk' as a published package — statically in the
    // agent code and via dynamic import() in the helper scripts — but nx, seeing it
    // as a workspace project, would rewrite those imports to source. Off tree-wide
    // on purpose: it's a published leaf, so narrowing to SDK-only files would add
    // maintenance for no real boundary value.
    files: ['libs/pi-extension/**/*.{ts,tsx,js,jsx,mjs,cjs}'],
    rules: {
      '@nx/enforce-module-boundaries': 'off',
    },
  },
]
