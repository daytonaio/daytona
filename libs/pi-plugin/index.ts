/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * pi-daytona — run Pi's tools inside a remote, ephemeral Daytona sandbox.
 *
 * The agent runs locally; only tool execution
 * (bash + file I/O) is redirected into a Daytona container. Activation is
 * launch-scoped via the `--daytona` flag; the sandbox is torn down on exit.
 *
 * Blueprint: examples/extensions/ssh.ts from @earendil-works/pi-coding-agent.
 */

import { Daytona, type Sandbox } from '@daytona/sdk'
import type { ExtensionAPI, ExtensionContext } from '@earendil-works/pi-coding-agent'
import { SessionManager } from '@earendil-works/pi-coding-agent'
import { resolveApiKey } from './src/auth.ts'
import { registerTools } from './src/tools.ts'
import { execCommand } from './src/sandbox.ts'
import { joinPath, normalizeRepoUrl, repoName, shellQuote, shortId } from './src/util.ts'
import {
  type RepoSlug,
  branchUrl,
  compareUrl,
  deleteBranch,
  detectLocalRepo,
  ensureBranch,
  getBranchAhead,
  getBranchSha,
  getDefaultBranch,
  getGithubToken,
  mergeBranch,
  parseRepoSlug,
  prUrl,
} from './src/github.ts'
import { pushChanges } from './src/sync.ts'

/** Session custom-entry type recording the sandbox bound to this session. */
const SESSION_ENTRY = 'daytona-session'

/** GitHub sync target for a session (set only when pushing is enabled). */
interface GitTarget {
  slug: RepoSlug
  base: string
  branch: string
}

/** Persisted record so a session can reattach its sandbox on resume. */
interface SessionEntryData {
  sandboxId: string
  cwd: string
  git?: GitTarget
}

/** State for the sandbox bound to the current session. */
interface ActiveSandbox {
  sandbox: Sandbox
  /** Working directory inside the sandbox (repo root, or workspace when no --repo). */
  cwd: string
  /** GitHub sync target — set only when --repo is a github.com repo and gh has a token. */
  git?: GitTarget
}

export default function (pi: ExtensionAPI) {
  pi.registerFlag('daytona', { description: 'Run tools inside a Daytona sandbox', type: 'boolean' })
  pi.registerFlag('repo', { description: 'Git repo to clone into the sandbox', type: 'string' })
  pi.registerFlag('branch', { description: 'Branch to clone (with --repo)', type: 'string' })
  pi.registerFlag('snapshot', { description: 'Daytona snapshot/base image to use', type: 'string' })
  pi.registerFlag('public', { description: 'Create a public sandbox (preview URLs need no token)', type: 'boolean' })
  pi.registerFlag('idle-stop', { description: 'Minutes idle before the sandbox pauses (default 15)', type: 'string' })

  // Resolved lazily on session_start (CLI flags are not available at load time).
  let active: ActiveSandbox | null = null
  // Daytona client for the session; reused for reaping at shutdown.
  let daytona: Daytona | null = null

  // Register all tools (each runs in the sandbox when one is active).
  registerTools(pi, () => active)

  // --- Informational commands (read-only; don't change the backend) ---

  // Status for the active sandbox: state, working dir, branch, sync status, PR link.
  pi.registerCommand('sandbox', {
    description: "Show the active Daytona sandbox's status",
    handler: async (_args, ctx) => {
      if (!active) {
        ctx.ui.notify('No Daytona sandbox is active. Launch Pi with --daytona.', 'warning')
        return
      }
      const { sandbox, cwd, git } = active
      try {
        await sandbox.refreshData()
      } catch {
        // Show last-known data if the refresh call fails.
      }
      const state = sandbox.state ?? 'unknown'
      const visibility = sandbox.public ? 'public' : 'private'
      const lines = [
        `☁ ${shortId(sandbox.id)} · ${state} · ${visibility}${sandbox.snapshot ? ` · ${sandbox.snapshot}` : ''}`,
        `cwd: ${cwd}`,
      ]
      if (git) {
        let sync = ''
        try {
          const st = await sandbox.git.status(cwd)
          const ahead = st.ahead ?? 0
          const dirty = st.fileStatus?.length ?? 0
          sync = ` · ${ahead} unpushed${dirty ? `, ${dirty} uncommitted` : ''}`
        } catch {
          // status is best-effort
        }
        lines.push(`branch: ${git.branch} → ${git.base}${sync}`)
        lines.push(`github: ${branchUrl(git.slug, git.branch)}`)
      } else {
        lines.push('github sync: off (launch with --repo and `gh auth login`)')
      }
      ctx.ui.notify(lines.join('\n'), 'info')
    },
  })

  // Merge this session's branch into its base on GitHub (direct API merge).
  pi.registerCommand('merge', {
    description: "Merge this session's branch into its base on GitHub",
    handler: async (_args, ctx) => {
      if (!active?.git) {
        ctx.ui.notify('Merge needs a GitHub repo. Launch Pi with --repo.', 'warning')
        return
      }
      const { slug, base, branch } = active.git
      const ok = await ctx.ui.confirm(
        'Merge branch',
        `Merge ${branch} into ${base}? This does a direct GitHub merge (merge commit).`,
      )
      if (!ok) return
      try {
        // Push the agent's latest commits first so the merge includes them.
        const token = await getGithubToken(pi)
        await pushChanges({ sandbox: active.sandbox, cwd: active.cwd, pushEnabled: true }, token)
        const res = await mergeBranch(pi, slug, base, branch)
        if (!res.ok) {
          ctx.ui.notify(`Merge failed: ${res.message}`, 'error')
          return
        }
        ctx.ui.notify(`Merged ${branch} into ${base} ✓`, 'info')
      } catch (err) {
        ctx.ui.notify(`Merge failed: ${errorMessage(err)}`, 'error')
      }
    },
  })

  // Open GitHub's pre-filled "Open a pull request" page for this session's branch.
  pi.registerCommand('pr', {
    description: "Open a pull request for this session's branch on GitHub",
    handler: async (_args, ctx) => {
      if (!active?.git) {
        ctx.ui.notify('Opening a PR needs a GitHub repo. Launch Pi with --repo.', 'warning')
        return
      }
      const { slug, base, branch } = active.git
      const url = prUrl(slug, base, branch)
      try {
        await openUrl(pi, url)
      } catch {
        // Couldn't launch a browser — the URL is still shown below.
      }
      ctx.ui.notify(`Open PR: ${url}`, 'info')
    },
  })

  // Open this session's branch compare view on GitHub in the browser.
  pi.registerCommand('compare', {
    description: "Open this session's branch compare view on GitHub",
    handler: async (_args, ctx) => {
      if (!active?.git) {
        ctx.ui.notify('No GitHub branch for this session. Launch Pi with --repo.', 'warning')
        return
      }
      const { slug, base, branch } = active.git
      const url = compareUrl(slug, base, branch)
      try {
        await openUrl(pi, url)
      } catch {
        // Couldn't launch a browser — the URL is still shown below.
      }
      ctx.ui.notify(`Compare: ${url}`, 'info')
    },
  })

  // Open this session's branch on GitHub in the browser.
  pi.registerCommand('github', {
    description: "Open this session's branch on GitHub",
    handler: async (_args, ctx) => {
      if (!active?.git) {
        ctx.ui.notify('No GitHub branch for this session. Launch Pi with --repo.', 'warning')
        return
      }
      const url = branchUrl(active.git.slug, active.git.branch)
      try {
        await openUrl(pi, url)
      } catch {
        // Couldn't launch a browser — the URL is still shown below.
      }
      ctx.ui.notify(`GitHub: ${url}`, 'info')
    },
  })

  // --- Lifecycle ---

  pi.on('session_start', async (event, ctx) => {
    if (pi.getFlag('daytona') !== true) return
    if (active) return // already running (e.g. after reload)

    const apiKey = await resolveApiKey(ctx)
    if (!apiKey) {
      ctx.ui.notify('Daytona: no API key found — staying local. Set DAYTONA_API_KEY.', 'error')
      return
    }

    const dt = new Daytona({ apiKey })
    daytona = dt
    const persisted = ctx.sessionManager.getSessionFile() !== undefined
    const sessionId = ctx.sessionManager.getSessionId()

    // Reap sandboxes whose session was deleted from the resume menu. Runs in
    // the background so it never slows startup.
    if (persisted) void reapOrphans(dt)

    setStatus(ctx, '☁ daytona · spinning up sandbox…')
    const startedAt = Date.now()

    // Tracked outside the try so the catch can clean up a sandbox that was
    // created but whose later setup (clone, git init, …) failed — otherwise the
    // sandbox would leak (no session entry points at it, so reapOrphans can't
    // attribute it either).
    let created: Sandbox | undefined

    try {
      // Reattach to this session's existing sandbox on resume/reload. A fork
      // always gets a fresh sandbox (branched off the parent below).
      if (persisted && event.reason !== 'fork') {
        const prev = latestSessionEntry(ctx)
        if (prev) {
          try {
            setStatus(ctx, '☁ daytona · resuming sandbox…')
            const sandbox = await dt.get(prev.sandboxId)
            await ensureStarted(sandbox)
            active = { sandbox, cwd: prev.cwd, git: prev.git }
            ctx.ui.notify(
              `Reattached sandbox · ${shortId(sandbox.id)}${prev.git ? ` · ${prev.git.branch}` : ''}`,
              'info',
            )
            setRunningStatus(ctx, sandbox.id, prev.cwd)
            return
          } catch {
            // Sandbox is gone (reaped/deleted) — fall through and create a fresh one.
          }
        }
      }

      const snapshot = stringFlag(pi.getFlag('snapshot'))
      const isPublic = pi.getFlag('public') === true

      const sandbox = await dt.create({
        // Full session id for a globally-unique sandbox name (branches stay short).
        name: `pi-${sessionId}`,
        snapshot,
        public: isPublic,
        // Idle PAUSES the sandbox (filesystem preserved); the next tool call
        // transparently restarts it (see withRecovery). Auto-delete is disabled —
        // the sandbox is reaped only when its session is deleted (see reapOrphans).
        // Default 15 min (matches Daytona's own default); overridable via --idle-stop.
        autoStopInterval: numberFlag(pi.getFlag('idle-stop')) ?? 15,
        autoDeleteInterval: -1, // never auto-delete
        labels: { 'created-by': 'pi-daytona', 'session-id': sessionId },
      })
      created = sandbox

      const home = (await sandbox.getUserHomeDir()) ?? '/home/daytona'
      // Temporary git identity so the agent's commits (and our init commit) just work.
      await execCommand(
        sandbox,
        `git config --global user.name "pi-agent" && git config --global user.email "agent@pi.daytona"`,
        home,
      )
      let cwd = home
      let git: GitTarget | undefined

      // Use --repo if given; otherwise fall back to the git repo you launched Pi in.
      let repo = stringFlag(pi.getFlag('repo'))
      let detectedBranch: string | undefined
      if (!repo) {
        const local = await detectLocalRepo(pi, ctx.sessionManager.getCwd())
        if (local) {
          repo = local.url
          detectedBranch = local.branch || undefined
        }
      }

      if (repo) {
        cwd = joinPath(home, repoName(repo))
        const slug = parseRepoSlug(normalizeRepoUrl(repo))
        const token = slug ? await getGithubToken(pi) : undefined

        if (slug && token) {
          // Each session gets its own GitHub branch pi/<short-session-id>. We create
          // the ref on GitHub first (off the base), then clone that branch so
          // the sandbox has an upstream to push back to (see sync.ts).
          const branch = `pi/${shortId(sessionId)}`
          let base = stringFlag(pi.getFlag('branch'))
          // A fork branches off the parent session's branch.
          if (event.reason === 'fork') {
            const parent = latestSessionEntry(ctx)
            if (parent?.git) base = parent.git.branch
          }
          if (!base) base = detectedBranch // the branch you're on locally
          if (!base) base = await getDefaultBranch(pi, slug)
          if (!base) throw new Error('Could not resolve a base branch on GitHub.')

          const sha = await getBranchSha(pi, slug, base)
          if (!sha) throw new Error(`Base branch '${base}' not found on GitHub.`)
          await ensureBranch(pi, slug, branch, sha)
          // Clone over HTTPS with the token regardless of the origin's format
          // (a detected origin may be SSH, which the token can't authenticate).
          const cloneUrl = `https://github.com/${slug.owner}/${slug.repo}.git`
          setStatus(ctx, `☁ daytona · cloning ${slug.owner}/${slug.repo}…`)
          await sandbox.git.clone(cloneUrl, cwd, branch, undefined, 'x-access-token', token)
          git = { slug, base, branch }
        } else {
          // Not a github.com repo, or no gh token: clone read-only, no push.
          setStatus(ctx, `☁ daytona · cloning ${repoName(repo)}…`)
          await sandbox.git.clone(normalizeRepoUrl(repo), cwd, stringFlag(pi.getFlag('branch')) ?? detectedBranch)
          ctx.ui.notify('Daytona: GitHub sync disabled (needs `gh auth login` and a github.com repo).', 'warning')
        }
      } else {
        // Not in a git repo: throwaway local repo so the agent can still commit
        // (never pushed). The initial empty commit gives HEAD a valid ref.
        cwd = joinPath(home, 'workspace')
        await execCommand(
          sandbox,
          `mkdir -p ${shellQuote(cwd)} && cd ${shellQuote(cwd)} && git init -q -b pi && ` +
            `git commit -q --allow-empty -m "pi: init"`,
          home,
        )
      }

      active = { sandbox, cwd, git }
      // Record the sandbox so this session can reattach it after a restart.
      if (persisted) {
        const data: SessionEntryData = { sandboxId: sandbox.id, cwd, git }
        pi.appendEntry(SESSION_ENTRY, data)
      }

      const secs = ((Date.now() - startedAt) / 1000).toFixed(1)
      const branchInfo = git ? ` · ${git.branch}` : ''
      ctx.ui.notify(`Sandbox ready · ${shortId(sandbox.id)}${branchInfo} · ${secs}s`, 'info')
      setRunningStatus(ctx, sandbox.id, cwd)
    } catch (err) {
      active = null
      // Delete the half-created sandbox so it doesn't leak (best-effort).
      if (created) await created.delete().catch(() => undefined)
      setStatus(ctx, undefined)
      ctx.ui.notify(`Daytona: failed to start sandbox — ${errorMessage(err)}`, 'error')
    }
  })

  // Rewrite the agent's "current working directory" to the sandbox path.
  // Match the whole line (not a literal host path) so this works regardless of
  // what Pi used as the prompt cwd — avoids a silent no-op if they diverge.
  // Point the agent's working-directory line at the sandbox and add the
  // commit-not-push guideline. Project context (AGENTS.md/CLAUDE.md) is left to
  // Pi's default loading from the local files.
  pi.on('before_agent_start', (event) => {
    if (!active) return
    const cwdLine = `Current working directory: ${active.cwd} (Daytona sandbox ${shortId(active.sandbox.id)})`
    let systemPrompt = event.systemPrompt.replace(/Current working directory: .*/g, cwdLine)
    systemPrompt +=
      '\n\nThis project is a git repository inside a Daytona sandbox. After you finish a unit of work, ' +
      'commit it with git (e.g. `git add -A && git commit -m "..."`). Do not push — pushing is handled automatically.'
    return { systemPrompt }
  })

  // After each agent loop ends, push any commits the agent made to the session's
  // GitHub branch. We don't commit here — the agent commits its own work. The
  // push is serialized and skips a branch with nothing ahead of its remote.
  pi.on('agent_end', async (_event, ctx) => {
    if (!active?.git) return
    try {
      const token = await getGithubToken(pi)
      const res = await pushChanges({ sandbox: active.sandbox, cwd: active.cwd, pushEnabled: true }, token)
      if (res.pushed) {
        ctx.ui.notify(
          `Pushed ${active.git.branch} → ${compareUrl(active.git.slug, active.git.base, active.git.branch)}`,
          'info',
        )
      }
    } catch (err) {
      ctx.ui.notify(`Daytona: push failed — ${errorMessage(err)}`, 'warning')
    }
  })

  // On exit, flush a final sync, then KEEP the sandbox (autoStop pauses it) so
  // the session can be resumed later. The sandbox is only deleted once its
  // session is deleted from the resume menu — handled by reapOrphans, which we
  // also run here to catch sessions deleted during this run.
  pi.on('session_shutdown', async (event, ctx) => {
    if (!active) return
    // Skip only in-process session handoffs (new/resume/fork) — there the session
    // continues and its sandbox is deliberately kept/reattached. Everything else
    // (quit, reload, or an unlabeled shutdown from Pi builds that emit no reason)
    // is a real teardown and must run cleanup. Matching an allow-list of
    // {quit,reload} instead silently skipped cleanup whenever reason was absent.
    if (event.reason === 'new' || event.reason === 'resume' || event.reason === 'fork') return
    const current = active
    active = null
    setStatus(ctx, undefined)

    const persisted = ctx.sessionManager.getSessionFile() !== undefined
    if (persisted) {
      // Reap sandboxes whose session no longer exists; this session's own
      // sandbox stays (its session file still exists) and is paused by autoStop.
      if (daytona) await reapOrphans(daytona)
    } else {
      // In-memory session: nothing to resume, so tidy up GitHub and delete the
      // sandbox now. Push any commits made after the last agent_end (e.g. a
      // manual `!git commit`) so work isn't silently lost; then, if the branch
      // contributed nothing, delete the throwaway ref we created at startup so
      // it doesn't leak.
      if (current.git) {
        try {
          const token = await getGithubToken(pi)
          await pushChanges({ sandbox: current.sandbox, cwd: current.cwd, pushEnabled: true }, token)
          // Delete the branch only if it contributed nothing (HEAD == base on
          // GitHub). Compare on the remote — local ahead-of-remote is 0 right
          // after the push and would wrongly flag branches with real work.
          const ahead = await getBranchAhead(pi, current.git.slug, current.git.base, current.git.branch)
          if (ahead === 0) await deleteBranch(pi, current.git.slug, current.git.branch)
        } catch {
          // Best-effort cleanup; a leaked branch is preferable to lost work.
        }
      }
      try {
        await current.sandbox.delete()
      } catch {
        // Best-effort: autoStop + autoDelete reap it later if this didn't run.
      }
    }
  })
}

/** Most recent Daytona sandbox record in this session (for reattach / fork base). */
function latestSessionEntry(ctx: ExtensionContext): SessionEntryData | undefined {
  const entries = ctx.sessionManager.getEntries()
  for (let i = entries.length - 1; i >= 0; i--) {
    const e = entries[i] as { type?: string; customType?: string; data?: unknown }
    if (e.type === 'custom' && e.customType === SESSION_ENTRY) {
      return e.data as SessionEntryData
    }
  }
  return undefined
}

/** Ensure a sandbox is running, starting it if it was paused or archived. */
async function ensureStarted(sandbox: Sandbox): Promise<void> {
  try {
    await sandbox.refreshData()
  } catch {
    // If we can't read state, let start() surface the real error.
  }
  if (sandbox.state !== 'started') {
    await sandbox.start()
  }
}

/**
 * Delete pi-daytona sandboxes whose session no longer exists. This is how a
 * sandbox gets cleaned up when its session is deleted from the resume menu —
 * Pi has no session-deleted hook, so we reconcile against SessionManager.listAll().
 * Best-effort: never throws.
 */
async function reapOrphans(daytona: Daytona): Promise<void> {
  try {
    const live = new Set((await SessionManager.listAll()).map((s) => s.id))
    const orphans: Sandbox[] = []
    for await (const sandbox of daytona.list({ labels: { 'created-by': 'pi-daytona' } })) {
      let labels = sandbox.labels
      if (!labels || Object.keys(labels).length === 0) {
        try {
          await sandbox.refreshData()
          labels = sandbox.labels
        } catch {
          continue
        }
      }
      const sid = labels?.['session-id']
      // Only reap sandboxes we can attribute to a session that no longer exists.
      if (sid && !live.has(sid)) orphans.push(sandbox)
    }
    await Promise.allSettled(orphans.map((s) => s.delete()))
  } catch {
    // best-effort reconciliation
  }
}

// --- helpers ---

function setStatus(ctx: ExtensionContext, text: string | undefined): void {
  ctx.ui.setStatus('daytona', text === undefined ? undefined : ctx.ui.theme.fg('accent', text))
}

function setRunningStatus(ctx: ExtensionContext, id: string, cwd: string): void {
  setStatus(ctx, `☁ daytona · ${shortId(id)} · running · ${cwd}`)
}

function stringFlag(value: boolean | string | undefined): string | undefined {
  return typeof value === 'string' && value.length > 0 ? value : undefined
}

/** Parse a flag into a positive integer (minutes), or undefined if unset/invalid. */
function numberFlag(value: boolean | string | undefined): number | undefined {
  const n = typeof value === 'string' ? Number(value) : NaN
  return Number.isFinite(n) && n > 0 ? Math.floor(n) : undefined
}

function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}

/** Open a URL in the host's default browser (best-effort, cross-platform). */
async function openUrl(pi: ExtensionAPI, url: string): Promise<void> {
  if (process.platform === 'darwin') {
    await pi.exec('open', [url])
  } else if (process.platform === 'win32') {
    await pi.exec('cmd', ['/c', 'start', '', url])
  } else {
    await pi.exec('xdg-open', [url])
  }
}
