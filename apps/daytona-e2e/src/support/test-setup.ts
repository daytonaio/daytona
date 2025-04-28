/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import axios from 'axios'

module.exports = async function () {
  // Configure axios for tests to use.
  const host = process.env.HOST ?? 'localhost'
  const port = process.env.PORT ?? '3000'
  axios.defaults.baseURL = `http://${host}:${port}`
}
