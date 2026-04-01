/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/** @type {import('ts-jest').JestConfigWithTsJest} */
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  transform: {
    '^.+\\.tsx?$': [
      'ts-jest',
      {
        tsconfig: {
          resolveJsonModule: true,
          esModuleInterop: true,
          module: 'commonjs',
          moduleResolution: 'node10',
          experimentalDecorators: true,
          emitDecoratorMetadata: true,
          target: 'ES2022',
          lib: ['es2022', 'dom'],
          skipLibCheck: true,
        },
      },
    ],
  },
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json', 'node'],
  testMatch: ['**/__tests__/**/*.test.ts'],
  moduleNameMapper: {
    '^@daytonaio/sdk$': '<rootDir>/src/index.ts',
    '^@daytonaio/api-client$': '<rootDir>/../api-client/src/index.ts',
    '^@daytonaio/api-client/(.*)$': '<rootDir>/../api-client/src/$1',
    '^@daytonaio/toolbox-api-client$': '<rootDir>/../toolbox-api-client/src/index.ts',
    '^@daytonaio/toolbox-api-client/(.*)$': '<rootDir>/../toolbox-api-client/src/$1',
  },
}
