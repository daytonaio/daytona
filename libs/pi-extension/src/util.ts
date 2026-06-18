/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/** Small helpers shared across the extension. */

/**
 * Quote a string for safe use as a single shell argument.
 * Uses single quotes so `$`, backticks, and globs are not expanded.
 */
export function shellQuote(arg: string): string {
  return `'${arg.replace(/'/g, `'\\''`)}'`
}

/** Normalize a repo argument into a clonable URL. Accepts `github.com/a/b`, full URLs, or `git@` SSH. */
export function normalizeRepoUrl(url: string): string {
  const trimmed = url.trim()
  if (/^[a-z][a-z0-9+.-]*:\/\//i.test(trimmed) || trimmed.startsWith('git@')) {
    return trimmed
  }
  return `https://${trimmed}`
}

/** Derive a directory name from a repo URL, e.g. `github.com/acme/api` -> `api`. */
export function repoName(url: string): string {
  const cleaned = url
    .replace(/\.git$/i, '')
    .replace(/[/]+$/, '')
    .replace(/^git@[^:]+:/, '')
  const parts = cleaned.split(/[/]/)
  return parts[parts.length - 1] || 'repo'
}

/** Short, human-friendly form of a sandbox id for status display. */
export function shortId(id: string): string {
  return id.slice(0, 8)
}

/** Join a base dir and a child segment with a single slash. */
export function joinPath(base: string, child: string): string {
  return `${base.replace(/[/]+$/, '')}/${child.replace(/^[/]+/, '')}`
}
