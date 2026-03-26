/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Matches the structure of a JWT token: three base64url-encoded segments
 * separated by dots (header.payload.signature).
 *
 * Both header and payload must start with `eyJ` — the base64url encoding of `{"`,
 * which is guaranteed since both are JSON objects.
 *
 * The signature segment has no prefix constraint since it is raw bytes.
 */
export const JWT_REGEX = /^eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$/
