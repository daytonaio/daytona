
        const content = `/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Default page size for tables
 */
export const DEFAULT_PAGE_SIZE = 25
`;
        try {
          new Function(content);
          process.exit(0);
        } catch (error) {
          console.error(error);
          process.exit(1);
        }
      