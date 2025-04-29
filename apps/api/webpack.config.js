/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

const { composePlugins, withNx } = require('@nx/webpack')
const path = require('path')
const glob = require('glob')

const migrationFiles = glob.sync('apps/api/src/migrations/*')
const migrationEntries = migrationFiles.reduce((acc, migrationFile) => {
  const entryName = migrationFile.substring(migrationFile.lastIndexOf('/') + 1, migrationFile.lastIndexOf('.'))
  acc[entryName] = migrationFile
  return acc
}, {})

module.exports = composePlugins(
  // Default Nx composable plugin
  withNx(),
  // Custom composable plugin
  (config, { options, context }) => {
    // `config` is the Webpack configuration object
    // `options` is the options passed to the `@nx/webpack:webpack` executor
    // `context` is the context passed to the `@nx/webpack:webpack` executor
    // customize configuration here
    config.output.devtoolModuleFilenameTemplate = function (info) {
      const rel = path.relative(process.cwd(), info.absoluteResourcePath)
      return `webpack:///./${rel}`
    }
    // add typeorm migrations as entry points
    for (const key in migrationEntries) {
      config.entry[`migrations/${key}`] = migrationEntries[key]
    }
    config.mode = process.env.NODE_ENV
    return config
  },
)
