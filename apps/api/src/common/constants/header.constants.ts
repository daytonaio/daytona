/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const CustomHeaders = {
  ORGANIZATION_ID: {
    name: 'X-Daytona-Organization-ID',
    description: 'Use with JWT to specify the organization ID',
    required: false,
    schema: {
      type: 'string',
    },
  },
  SOURCE: {
    name: 'X-Daytona-Source',
    description: 'Use to specify the source of the request',
    required: false,
    schema: {
      type: 'string',
    },
  },
  SDK_VERSION: {
    name: 'X-Daytona-SDK-Version',
    description: 'Use to specify the version of the SDK',
    required: false,
    schema: {
      type: 'string',
    },
  },
} as const
