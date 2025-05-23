/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * @module Errors
 */

/**
 * Base error for Daytona SDK.
 */
export class DaytonaError extends Error {}

export class DaytonaNotFoundError extends DaytonaError {}
