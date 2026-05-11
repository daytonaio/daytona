/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger'

/**
 * Signed direct-to-sandbox access bundle. Lets the SDK open a WebSocket to the
 * in-sandbox session-daemon (via the proxy chain) without going through the
 * API on every exec — same shape `sandbox.process.code_run` uses against the
 * classic daytona-daemon. The URL is self-authenticating (signed token lives
 * in the proxy subdomain), so the SDK does not pass `token` as a header. The
 * `token` field is returned for revocation / observability only.
 */
export class SessionAccessDto {
  @ApiProperty({ description: 'Signed http(s) base URL into the daemon (no trailing slash).' })
  httpUrl: string

  @ApiProperty({ description: 'Signed ws(s) URL for /sessions/:id/execute on the daemon.' })
  wsUrl: string

  @ApiProperty({
    description: 'Signed-URL token; embedded in the URL subdomain — informational, do not send as a header.',
  })
  token: string

  @ApiProperty({
    description:
      'When this signed URL stops being valid. The SDK refreshes via GET /sessions/:id/access before this point.',
  })
  tokenExpiresAt: string
}

export class SessionDto {
  @ApiProperty()
  id: string

  @ApiProperty()
  language: string

  @ApiPropertyOptional()
  cwd?: string

  @ApiProperty()
  createdAt: string

  @ApiPropertyOptional()
  lastUsedAt?: string

  @ApiProperty({
    description:
      'Computed expiry: min(lastUsedAt + idleTtl, createdAt + absoluteTtl). After this point the context will surface as SessionExpiredError.',
  })
  expiresAt: string

  @ApiPropertyOptional({
    type: SessionAccessDto,
    description:
      'Direct-to-sandbox access bundle. Present on createSession / POST /sessions/transients responses and refreshable via GET /sessions/:id/access. Omitted on listSessions to keep payloads small.',
  })
  access?: SessionAccessDto
}
