/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { StorybookConfig } from '@storybook/react-vite'
import { mergeConfig } from 'vite'
import path from 'path'

const config: StorybookConfig = {
  stories: ['../src/**/*.stories.@(ts|tsx)'],
  addons: ['@storybook/addon-essentials'],
  framework: {
    name: '@storybook/react-vite',
    options: {},
  },
  typescript: {
    reactDocgen: false,
  },
  viteFinal: async (config) => {
    return mergeConfig(config, {
      resolve: {
        alias: [
          {
            find: '@daytonaio/sdk',
            replacement: path.resolve(__dirname, '../../../libs/sdk-typescript/src'),
          },
          {
            find: '@',
            replacement: path.resolve(__dirname, '../src'),
          },
        ],
      },
    })
  },
}

export default config
