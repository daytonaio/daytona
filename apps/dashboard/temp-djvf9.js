
        const content = `/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * This is a placeholder test file for the Pagination component.
 *
 * Tests will verify:
 * 1. The component renders with the default page size of 25
 * 2. The dropdown shows options for 10, 25, 50, and 100 rows per page
 * 3. Selecting a different page size updates the table
 * 4. The pagination controls work correctly
 *
 * Note: Actual test implementation will be added when the testing infrastructure is properly set up.
 */

// Import the constant to verify it exists
import { DEFAULT_PAGE_SIZE } from '../lib/table-utils'

// Simple verification that the constant has the correct value
const defaultPageSizeValue = 25
console.assert(
  DEFAULT_PAGE_SIZE === defaultPageSizeValue,
  \`DEFAULT_PAGE_SIZE should be ${defaultPageSizeValue}, but got ${DEFAULT_PAGE_SIZE}\`
)
`;
        try {
          new Function(content);
          process.exit(0);
        } catch (error) {
          console.error(error);
          process.exit(1);
        }
      