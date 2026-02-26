/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const SANDBOX_LOOKUP_CACHE_TTL_MS = 10_000
export const SANDBOX_BUILD_INFO_CACHE_TTL_MS = 60_000
export const SANDBOX_ORG_ID_CACHE_TTL_MS = 60_000
export const TOOLBOX_PROXY_URL_CACHE_TTL_S = 30 * 60 // 30 minutes

type SandboxLookupCacheKeyArgs = {
  organizationId?: string | null
  returnDestroyed?: boolean
}

export function sandboxLookupCacheKeyById(args: SandboxLookupCacheKeyArgs & { sandboxId: string }): string {
  const organizationId = args.organizationId ?? 'none'
  const returnDestroyed = args.returnDestroyed ? 1 : 0
  return `sandbox:lookup:by-id:org:${organizationId}:returnDestroyed:${returnDestroyed}:value:${args.sandboxId}`
}

export function sandboxLookupCacheKeyByName(args: SandboxLookupCacheKeyArgs & { sandboxName: string }): string {
  const organizationId = args.organizationId ?? 'none'
  const returnDestroyed = args.returnDestroyed ? 1 : 0
  return `sandbox:lookup:by-name:org:${organizationId}:returnDestroyed:${returnDestroyed}:value:${args.sandboxName}`
}

export function sandboxLookupCacheKeyByAuthToken(args: { authToken: string }): string {
  return `sandbox:lookup:by-authToken:${args.authToken}`
}

type SandboxOrgIdCacheKeyArgs = {
  organizationId?: string
}

export function sandboxOrgIdCacheKeyById(args: SandboxOrgIdCacheKeyArgs & { sandboxId: string }): string {
  const organizationId = args.organizationId ?? 'none'
  return `sandbox:orgId:by-id:org:${organizationId}:value:${args.sandboxId}`
}

export function sandboxOrgIdCacheKeyByName(args: SandboxOrgIdCacheKeyArgs & { sandboxName: string }): string {
  const organizationId = args.organizationId ?? 'none'
  return `sandbox:orgId:by-name:org:${organizationId}:value:${args.sandboxName}`
}

export function toolboxProxyUrlCacheKey(regionId: string): string {
  return `toolbox-proxy-url:region:${regionId}`
}
