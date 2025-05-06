
        const content = `/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * This is a placeholder test file for the table-utils module.
 *
 * Tests will verify:
 * 1. The DEFAULT_PAGE_SIZE constant is exported
 * 2. The DEFAULT_PAGE_SIZE constant has the correct value (25)
 *
 * Note: Actual test implementation will be added when the testing infrastructure is properly set up.
 */

import { DEFAULT_PAGE_SIZE } from './table-utils'

// Simple verification that the constant has the correct value
const expectedValue = 25
console.assert(
  DEFAULT_PAGE_SIZE === expectedValue,
  \`DEFAULT_PAGE_SIZE should be ${expectedValue}, but got ${DEFAULT_PAGE_SIZE}\`
)
`;
        try {
          new Function(content);
          process.exit(0);
        } catch (error) {
          console.error(error);
          process.exit(1);
        }
      