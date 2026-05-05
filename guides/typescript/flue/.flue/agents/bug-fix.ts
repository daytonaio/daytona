/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */
import type { FlueContext } from '@flue/sdk/client'
import { Daytona } from '@daytona/sdk'
import { daytona } from '../connectors/daytona'
import * as v from 'valibot'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

export const triggers = { webhook: true }

const projectRoot = process.cwd()

const REPO_SLUG = /^[a-zA-Z0-9._-]+\/[a-zA-Z0-9._-]+$/
function assertRepoSlug(value: string, label: string): void {
  if (!REPO_SLUG.test(value)) {
    throw new Error(
      `Invalid ${label}: ${JSON.stringify(value)}. Must be in '<owner>/<repo>' form ` +
        `(alphanumerics, dots, hyphens, underscores only).`,
    )
  }
}

const PayloadSchema = v.object({
  repo: v.optional(v.string()),
  issueNumber: v.optional(v.number()),
  issueRepo: v.optional(v.string()),
})

const ResultSchema = v.object({
  branch: v.string(),
  prUrl: v.string(),
  testFile: v.string(),
  filesChanged: v.array(v.string()),
  summary: v.string(),
})

function requireEnv(env: Record<string, string | undefined>, key: string): string {
  const value = env[key]
  if (!value) {
    throw new Error(`Missing required environment variable: ${key}`)
  }
  return value
}

export default async function ({ init, payload, env }: FlueContext) {
  const parsed = v.parse(PayloadSchema, payload ?? {})
  const repo = parsed.repo ?? requireEnv(env, 'DEMO_REPO')
  assertRepoSlug(repo, 'repo')
  const issueNumber = parsed.issueNumber ?? Number.parseInt(requireEnv(env, 'DEMO_ISSUE'), 10)
  if (!Number.isInteger(issueNumber) || issueNumber <= 0) {
    throw new Error(`Invalid issueNumber: ${issueNumber} (must be a positive integer)`)
  }

  const daytonaApiKey = requireEnv(env, 'DAYTONA_API_KEY')
  const githubToken = requireEnv(env, 'GITHUB_TOKEN')
  const model = env.MODEL ?? 'anthropic/claude-sonnet-4-6'
  const projectDir = '/home/daytona/project'

  console.log(`[bug-fix] target: ${repo}#${issueNumber} (model: ${model})`)

  const client = new Daytona({ apiKey: daytonaApiKey })
  const sandbox = await client.create({
    envVars: { GH_TOKEN: githubToken },
  })
  console.log(`[bug-fix] sandbox ready (id: ${sandbox.id})`)

  type AgentHandle = Awaited<ReturnType<typeof init>>
  let setupAgent: AgentHandle | undefined
  let projectAgent: AgentHandle | undefined

  try {
    setupAgent = await init({
      sandbox: daytona(sandbox, { cleanup: true }),
      model,
    })
    const setup = await setupAgent.session()

    const ghCheck = await setup.shell('command -v gh >/dev/null && echo ok || echo missing')
    if (ghCheck.stdout.trim() !== 'ok') {
      console.log('[bug-fix] installing gh CLI in sandbox...')
      const ghInstall = await setup.shell('sudo apt-get update -qq && sudo apt-get install -y -qq gh')
      if (ghInstall.exitCode !== 0) {
        throw new Error(`Failed to install gh CLI: ${ghInstall.stderr || ghInstall.stdout}`)
      }
    }

    const ghSetup = await setup.shell('gh auth setup-git')
    if (ghSetup.exitCode !== 0) {
      throw new Error(`gh auth setup-git failed: ${ghSetup.stderr || ghSetup.stdout}`)
    }

    const userInfo = await setup.shell(`gh api user --jq '{login: .login, id: .id, name: .name}'`)
    if (userInfo.exitCode !== 0) {
      throw new Error(`Failed to query GitHub user: ${userInfo.stderr || userInfo.stdout}`)
    }
    const { login, id, name }: { login: string; id: number; name: string | null } = JSON.parse(userInfo.stdout)
    const gitUserName = name ?? login
    const gitUserEmail = `${id}+${login}@users.noreply.github.com`
    console.log(`[bug-fix] commits will be authored as ${gitUserName} <${gitUserEmail}>`)
    await setup.shell('git config --global user.email "$EMAIL"', {
      env: { EMAIL: gitUserEmail },
    })
    await setup.shell('git config --global user.name "$NAME"', {
      env: { NAME: gitUserName },
    })

    console.log(`[bug-fix] cloning ${repo} into sandbox...`)
    const cloneResult = await setup.shell(`gh repo clone ${repo} ${projectDir}`)
    if (cloneResult.exitCode !== 0) {
      throw new Error(`Failed to clone ${repo}: ${cloneResult.stderr || cloneResult.stdout}`)
    }

    const pmDetect = await setup.shell(
      'if [ -f pnpm-lock.yaml ]; then echo pnpm; ' +
        'elif [ -f yarn.lock ]; then echo yarn; ' +
        'elif [ -f bun.lockb ] || [ -f bun.lock ]; then echo bun; ' +
        'else echo npm; fi',
      { cwd: projectDir },
    )
    const packageManager = pmDetect.stdout.trim() || 'npm'
    console.log(`[bug-fix] detected package manager: ${packageManager}`)

    if (packageManager !== 'npm') {
      const pmCheck = await setup.shell(`command -v ${packageManager} >/dev/null && echo ok || echo missing`)
      if (pmCheck.stdout.trim() !== 'ok') {
        console.log(`[bug-fix] installing ${packageManager}...`)
        const pmInstall = await setup.shell(`npm install -g ${packageManager}`)
        if (pmInstall.exitCode !== 0) {
          const sudoInstall = await setup.shell(`sudo env PATH="$PATH" npm install -g ${packageManager}`)
          if (sudoInstall.exitCode !== 0) {
            throw new Error(`Failed to install ${packageManager}: ${sudoInstall.stderr || sudoInstall.stdout}`)
          }
        }
      }
    }

    console.log('[bug-fix] installing project dependencies...')
    const install = await setup.shell(`${packageManager} install`, { cwd: projectDir })
    if (install.exitCode !== 0) {
      throw new Error(`Dependency install failed: ${install.stderr || install.stdout}`)
    }

    let issueRepo = parsed.issueRepo ?? env.ISSUE_REPO
    if (!issueRepo) {
      const parentLookup = await setup.shell(
        `gh repo view ${repo} --json parent --jq 'if .parent then .parent.owner.login + "/" + .parent.name else "" end'`,
      )
      const detectedParent = parentLookup.stdout.trim()
      issueRepo = detectedParent || repo
    }
    assertRepoSlug(issueRepo, 'issueRepo')
    console.log(`[bug-fix] resolving issue source: ${issueRepo}`)

    console.log(`[bug-fix] fetching issue #${issueNumber} from ${issueRepo}...`)
    const issueResult = await setup.shell(
      `gh issue view ${issueNumber} --repo ${issueRepo} --json title,body,number,url,labels`,
      { cwd: projectDir },
    )
    if (issueResult.exitCode !== 0) {
      throw new Error(
        `Failed to fetch issue #${issueNumber} from ${issueRepo}: ` +
          `${issueResult.stderr || issueResult.stdout}\n` +
          `Hint: forks have issues disabled by default. The agent auto-detects the upstream parent ` +
          `via \`gh repo view\`; if your fork has no parent (or you want to override), set ` +
          `ISSUE_REPO in .env or pass "issueRepo" in the payload.`,
      )
    }

    console.log('[bug-fix] uploading skill into sandbox + excluding it from git...')
    const skillContent = readFileSync(resolve(projectRoot, '.agents/skills/bug-fix/SKILL.md'), 'utf-8')
    await setup.shell(`mkdir -p ${projectDir}/.agents/skills/bug-fix`)
    await sandbox.fs.uploadFile(Buffer.from(skillContent, 'utf-8'), `${projectDir}/.agents/skills/bug-fix/SKILL.md`)
    await setup.shell(`echo '.agents/' >> ${projectDir}/.git/info/exclude`)

    projectAgent = await init({
      id: `bug-fix-${issueNumber}`,
      sandbox: daytona(sandbox),
      cwd: projectDir,
      model,
    })
    const session = await projectAgent.session()

    console.log('[bug-fix] running TDD workflow (reproduce → fix → PR)...')
    const result = await session.skill('bug-fix', {
      args: {
        issueNumber,
        issueData: issueResult.stdout,
        repo,
        issueRepo,
        packageManager,
      },
      role: 'test-driven-developer',
      result: ResultSchema,
    })

    console.log(`[bug-fix] PR opened: ${result.prUrl}`)
    console.log(`[bug-fix] branch: ${result.branch}`)
    console.log(`[bug-fix] files changed: ${result.filesChanged.join(', ')}`)

    return result
  } finally {
    console.log('[bug-fix] tearing down agents + sandbox...')
    if (projectAgent) {
      try {
        await projectAgent.destroy()
      } catch (err) {
        console.error('[bug-fix] projectAgent.destroy() failed:', err)
      }
    }
    if (setupAgent) {
      try {
        await setupAgent.destroy()
      } catch (err) {
        console.error('[bug-fix] setupAgent.destroy() failed:', err)
      }
    } else {
      try {
        await sandbox.delete()
      } catch (err) {
        console.error('[bug-fix] sandbox.delete() fallback failed:', err)
      }
    }
  }
}
