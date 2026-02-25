import { docsSchema, i18nSchema } from '@astrojs/starlight/schema'
import { defineCollection, z } from 'astro:content'
import { generateI18nSchema } from 'src/i18n/generateI18nSchema'
import { localizePath } from 'src/i18n/utils'

import type { NavigationGroup } from '../utils/navigation'

export const collections = {
  docs: defineCollection({
    schema: docsSchema({
      extend: z.object({
        licence: z.string().optional(),
        distribution: z.string().optional(),
        hideTitleOnPage: z.boolean().optional(),
      }),
    }),
  }),
  i18n: defineCollection({
    type: 'data',
    schema: i18nSchema({
      extend: generateI18nSchema(),
    }),
  }),
}

export enum NavigationCategory {
  MAIN,
  GENERAL,
  TYPESCRIPT_SDK,
  PYTHON_SDK,
  GUIDES,
  RUBY_SDK,
  GO_SDK,
}

/**
 * relatedGroupCategory - Applicable only to main navigation links. All links with that category will be shown in the sidebar when the link is active.
 * category - Applicable to groups. All links with that category will be shown in the sidebar when the link with that category or the main link that is related to the category is active.
 * homePageHref - Applicable to groups. The href of the link that will be used as previous link for the pagination component (if the current link is the first in the list).
 * disablePagination - Applicable to all links. If true, the pagination component will not be shown for the link.
 * autopopulateFromDir - Applicable to groups. If set, the group will be populated with all the files (except index file) in the directory.
 */
export const getSidebarConfig = (
  locale: string,
  labels?: ((key: string) => string) | Record<string, string>
): NavigationGroup[] => {
  const t =
    typeof labels === 'function'
      ? labels
      : (key: string) => labels?.[key] ?? key

  if (!t || typeof t !== 'function') return []
  return [
    {
      type: 'group',
      category: NavigationCategory.MAIN,
      entries: [
        {
          type: 'link',
          href: localizePath('/docs', locale),
          label: t('sidebarconfig.documentation'),
          attrs: {
            icon: 'home.svg',
          },
          relatedGroupCategory: NavigationCategory.GENERAL,
        },
        {
          type: 'link',
          href: localizePath('/docs/typescript-sdk', locale),
          label: t('sidebarconfig.tsSdkReference'),
          attrs: {
            icon: 'package.svg',
          },
          relatedGroupCategory: NavigationCategory.TYPESCRIPT_SDK,
        },
        {
          type: 'link',
          href: localizePath('/docs/python-sdk', locale),
          label: t('sidebarconfig.pythonSdkReference'),
          attrs: {
            icon: 'package.svg',
          },
          relatedGroupCategory: NavigationCategory.PYTHON_SDK,
        },
        {
          type: 'link',
          href: localizePath('/docs/ruby-sdk', locale),
          label: t('sidebarconfig.rubySdkReference'),
          attrs: {
            icon: 'package.svg',
          },
          relatedGroupCategory: NavigationCategory.RUBY_SDK,
        },
        {
          type: 'link',
          href: localizePath('/docs/go-sdk', locale),
          label: t('sidebarconfig.goSdkReference'),
          attrs: {
            icon: 'package.svg',
          },
          relatedGroupCategory: NavigationCategory.GO_SDK,
        },
        {
          type: 'link',
          href: localizePath('/docs/tools/api', locale),
          label: t('sidebarconfig.apiReference'),
          disablePagination: true,
          attrs: {
            icon: 'server.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/tools/cli', locale),
          label: t('sidebarconfig.cliReference'),
          disablePagination: true,
          attrs: {
            icon: 'terminal.svg',
          },
        },
        // {
        //   type: 'link',
        //   href: 'https://www.daytona.io/dotfiles/guides',
        //   label: t('sidebarconfig.guides'),
        //   disablePagination: true,
        //   external: true,
        //   attrs: {
        //     icon: 'book.svg',
        //   },
        // },
      ],
    },
    {
      type: 'group',
      label: t('sidebarconfig.introduction'),
      homePageHref: localizePath('/docs', locale),
      category: NavigationCategory.GENERAL,
      entries: [
        {
          type: 'link',
          href: localizePath('/docs', locale),
          label: t('sidebarconfig.quickStart'),
          attrs: {
            icon: 'rocket.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/getting-started', locale),
          label: t('sidebarconfig.gettingStarted'),
          description: t('sidebarconfig.gettingStartedDescription'),
          attrs: {
            icon: 'bookmark.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/sandboxes', locale),
          label: t('sidebarconfig.sandboxes'),
          description: t('sidebarconfig.sandboxesDescription'),
          attrs: {
            icon: 'rectangle.svg',
          },
        },
      ],
    },
    {
      type: 'group',
      label: t('sidebarconfig.sandbox'),
      homePageHref: localizePath('/docs', locale),
      category: NavigationCategory.GENERAL,
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/configuration', locale),
          label: t('sidebarconfig.environment'),
          description: t('sidebarconfig.configurationDescription'),
          attrs: {
            icon: 'git-commit.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/snapshots', locale),
          label: t('sidebarconfig.snapshots'),
          description: t('sidebarconfig.snapshotsDescription'),
          attrs: {
            icon: 'layers.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/declarative-builder', locale),
          label: t('sidebarconfig.declarativeBuilder'),
          description: t('sidebarconfig.declarativeBuilderDescription'),
          attrs: {
            icon: 'prebuilds.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/volumes', locale),
          label: t('sidebarconfig.volumes'),
          description: t('sidebarconfig.volumesDescription'),
          attrs: {
            icon: 'container-registries.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/regions', locale),
          label: t('sidebarconfig.regions'),
          description: t('sidebarconfig.regionsDescription'),
          attrs: {
            icon: 'globe.svg',
          },
        },
      ],
    },
    {
      type: 'group',
      label: t('sidebarconfig.agentTools'),
      homePageHref: localizePath('/docs', locale),
      category: NavigationCategory.GENERAL,
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/file-system-operations', locale),
          label: t('sidebarconfig.fileSystem'),
          description: t('sidebarconfig.fileSystemDescription'),
          attrs: {
            icon: 'folder.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/git-operations', locale),
          label: t('sidebarconfig.gitOperations'),
          description: t('sidebarconfig.gitOperationsDescription'),
          attrs: {
            icon: 'git-branch.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/language-server-protocol', locale),
          label: t('sidebarconfig.languageServerProtocol'),
          description: t('sidebarconfig.languageServerProtocolDescription'),
          attrs: {
            icon: 'pulse.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/process-code-execution', locale),
          label: t('sidebarconfig.processCodeExecution'),
          description: t('sidebarconfig.processCodeExecutionDescription'),
          attrs: {
            icon: 'computer.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/pty', locale),
          label: t('sidebarconfig.pty'),
          description: t('sidebarconfig.ptyDescription'),
          attrs: {
            icon: 'terminal.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/log-streaming', locale),
          label: t('sidebarconfig.logStreaming'),
          description: t('sidebarconfig.logStreamingDescription'),
          attrs: {
            icon: 'log2.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/mcp', locale),
          label: t('sidebarconfig.mcpServer'),
          disablePagination: true,
          attrs: {
            icon: 'server.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/computer-use', locale),
          label: t('sidebarconfig.computerUse'),
          disablePagination: true,
          attrs: {
            icon: 'computer.svg',
          },
        },
      ],
    },
    {
      type: 'group',
      label: t('sidebarconfig.humanTools'),
      homePageHref: localizePath('/docs', locale),
      category: NavigationCategory.GENERAL,
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/web-terminal', locale),
          label: t('sidebarconfig.webTerminal'),
          description: t('sidebarconfig.webTerminalDescription'),
          attrs: {
            icon: 'terminal.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/ssh-access', locale),
          label: t('sidebarconfig.sshAccess'),
          description: t('sidebarconfig.sshAccessDescription'),
          attrs: {
            icon: 'terminal.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/vnc-access', locale),
          label: t('sidebarconfig.vncAccess'),
          description: t('sidebarconfig.vncAccessDescription'),
          attrs: {
            icon: 'vnc.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/vpn-connections', locale),
          label: t('sidebarconfig.vpnConnection'),
          description: t('sidebarconfig.vpnConnectionDescription'),
          attrs: {
            icon: 'shield.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/preview', locale),
          label: t('sidebarconfig.preview'),
          description: t('sidebarconfig.previewDescription'),
          attrs: {
            icon: 'globe.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/playground', locale),
          label: t('sidebarconfig.playground'),
          description: t('sidebarconfig.playgroundDescription'),
          attrs: {
            icon: 'playground.svg',
          },
        },
      ],
    },
    {
      type: 'group',
      label: t('sidebarconfig.systemTools'),
      homePageHref: localizePath('/docs', locale),
      category: NavigationCategory.GENERAL,
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/webhooks', locale),
          label: t('sidebarconfig.webhooks'),
          description: t('sidebarconfig.webhooksDescription'),
          attrs: {
            icon: 'webhook.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/network-limits', locale),
          label: t('sidebarconfig.networkLimits'),
          description: t('sidebarconfig.networkLimitsDescription'),
          attrs: {
            icon: 'network-limits.svg',
          },
        },
      ],
    },
    {
      type: 'group',
      label: t('sidebarconfig.integrations'),
      homePageHref: localizePath('/docs/guides', locale),
      category: NavigationCategory.GUIDES,
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/guides', locale),
          label: t('sidebarconfig.guides'),
          hideInSidebar: true,
        },
        {
          type: 'link',
          href: localizePath('/docs/guides/data-analysis-with-ai', locale),
          label: t('sidebarconfig.dataAnalysis'),
          disablePagination: true,
          attrs: {
            icon: 'chart.svg',
          },
        },
        {
          type: 'link',
          href: localizePath(
            '/docs/guides/agentkit/inngest-agentkit-coding-agent',
            locale
          ),
          label: t('sidebarconfig.inngestAgentKit'),
          disablePagination: true,
          attrs: {
            icon: 'inngest-agentkit.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/guides/opencode', locale),
          label: t('sidebarconfig.opencode'),
          disablePagination: true,
          attrs: {
            icon: 'opencode.svg',
          },
        },
        {
          type: 'link',
          href: localizePath(
            '/docs/guides/opencode/opencode-plugin',
            locale
          ),
          label: t('sidebarconfig.opencode'),
          disablePagination: true,
          hideInSidebar: true,
        },
        {
          type: 'link',
          href: localizePath(
            '/docs/guides/opencode/opencode-web-agent',
            locale
          ),
          label: t('sidebarconfig.opencode'),
          disablePagination: true,
          hideInSidebar: true,
        },
        {
          type: 'link',
          href: localizePath(
            '/docs/guides/langchain/langchain-data-analysis',
            locale
          ),
          label: t('sidebarconfig.langchain'),
          disablePagination: true,
          attrs: {
            icon: 'langchain.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/guides/mastra/mastra-coding-agent', locale),
          label: t('sidebarconfig.mastra'),
          disablePagination: true,
          attrs: {
            icon: 'mastra.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/guides/claude', locale),
          label: t('sidebarconfig.claude'),
          disablePagination: true,
          attrs: {
            icon: 'claude.svg',
          },
        },
        {
          type: 'link',
          href: localizePath(
            '/docs/guides/claude/claude-code-run-cli-sandbox',
            locale
          ),
          label: t('sidebarconfig.claude'),
          disablePagination: true,
          hideInSidebar: true,
        },
        {
          type: 'link',
          href: localizePath(
            '/docs/guides/claude/claude-agent-sdk-interactive-terminal-sandbox',
            locale
          ),
          label: t('sidebarconfig.claude'),
          disablePagination: true,
          hideInSidebar: true,
        },
        {
          type: 'link',
          href: localizePath(
            '/docs/guides/claude/claude-agent-sdk-connect-service-sandbox',
            locale
          ),
          label: t('sidebarconfig.claude'),
          disablePagination: true,
          hideInSidebar: true,
        },
        {
          type: 'link',
          href: localizePath(
            '/docs/guides/claude/claude-code-run-tasks-stream-logs-sandbox',
            locale
          ),
          label: t('sidebarconfig.claude'),
          disablePagination: true,
          hideInSidebar: true,
        },
        {
          type: 'link',
          href: localizePath('/docs/guides/openclaw', locale),
          label: t('sidebarconfig.openclaw'),
          disablePagination: true,
          attrs: {
            icon: 'openclaw.svg',
          },
        },
        {
          type: 'link',
          href: localizePath(
            '/docs/guides/openclaw/openclaw-run-secure-sandbox',
            locale
          ),
          label: t('sidebarconfig.openclaw'),
          disablePagination: true,
          hideInSidebar: true,
        },
        {
          type: 'link',
          href: localizePath(
            '/docs/guides/openclaw/openclaw-sdk-sandbox',
            locale
          ),
          label: t('sidebarconfig.openclaw'),
          disablePagination: true,
          hideInSidebar: true,
        },
        {
          type: 'link',
          href: localizePath(
            '/docs/guides/codex/codex-sdk-interactive-terminal-sandbox',
            locale
          ),
          label: t('sidebarconfig.codex'),
          disablePagination: true,
          attrs: {
            icon: 'openai.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/guides/google-adk-code-generator', locale),
          label: t('sidebarconfig.googleAdk'),
          disablePagination: true,
          attrs: {
            icon: 'google-adk.svg',
          },
        },
        {
          type: 'link',
          href: localizePath(
            '/docs/guides/letta-code/letta-code-agent',
            locale
          ),
          label: t('sidebarconfig.lettacode'),
          disablePagination: true,
          attrs: {
            icon: 'letta-code.svg',
          },
        },
        {
          type: 'link',
          href: localizePath(
            '/docs/guides/reinforcement-learning/trl-grpo-training',
            locale
          ),
          label: t('sidebarconfig.trlGrpo'),
          disablePagination: true,
          attrs: {
            icon: 'trl-logo.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/guides/rlm', locale),
          label: t('sidebarconfig.rlm'),
          disablePagination: true,
          attrs: {
            icon: 'recursive-lm.svg',
          },
        },
        {
          type: 'link',
          href: localizePath(
            '/docs/guides/rlm/recursive-language-models',
            locale
          ),
          label: t('sidebarconfig.rlm'),
          disablePagination: true,
          hideInSidebar: true,
        },
        {
          type: 'link',
          href: localizePath('/docs/guides/rlm/dspy-rlms', locale),
          label: t('sidebarconfig.rlm'),
          disablePagination: true,
          hideInSidebar: true,
        },
      ],
    },
    {
      type: 'group',
      label: t('sidebarconfig.security'),
      homePageHref: localizePath('/docs', locale),
      category: NavigationCategory.GENERAL,
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/audit-logs', locale),
          label: t('sidebarconfig.auditLogs'),
          description: t('sidebarconfig.auditLogsDescription'),
          attrs: {
            icon: 'log.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/security-exhibit', locale),
          label: t('sidebarconfig.securityExhibit'),
          description: t('sidebarconfig.securityExhibitDescription'),
          attrs: {
            icon: 'security.svg',
          },
        },
      ],
    },
    {
      type: 'group',
      label: t('sidebarconfig.deployments'),
      homePageHref: localizePath('/docs', locale),
      category: NavigationCategory.GENERAL,
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/oss-deployment', locale),
          label: t('sidebarconfig.ossDeployment'),
          disablePagination: true,
          attrs: {
            icon: 'computer.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/runners', locale),
          label: t('sidebarconfig.customerManagedCompute'),
          description: t('sidebarconfig.customerManagedComputeDescription'),
          disablePagination: true,
          attrs: {
            icon: 'computer.svg',
          },
        },
      ],
    },
    {
      type: 'group',
      label: t('sidebarconfig.experimental'),
      homePageHref: localizePath('/docs', locale),
      category: NavigationCategory.GENERAL,
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/experimental/otel-collection', locale),
          label: t('sidebarconfig.otelCollection'),
          disablePagination: true,
          attrs: {
            icon: 'telemetry.svg',
          },
        },
      ],
    },
    {
      type: 'group',
      label: t('sidebarconfig.accountManagement'),
      homePageHref: localizePath('/docs', locale),
      category: NavigationCategory.GENERAL,
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/api-keys', locale),
          label: t('sidebarconfig.apiKeys'),
          description: t('sidebarconfig.apiKeysDescription'),
          attrs: {
            icon: 'tag.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/organizations', locale),
          label: t('sidebarconfig.organizations'),
          description: t('sidebarconfig.organizationsDescription'),
          attrs: {
            icon: 'building.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/limits', locale),
          label: t('sidebarconfig.limits'),
          description: t('sidebarconfig.limitsDescription'),
          attrs: {
            icon: 'log.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/billing', locale),
          label: t('sidebarconfig.billing'),
          description: t('sidebarconfig.billingDescription'),
          attrs: {
            icon: 'credit-card.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/linked-accounts', locale),
          label: t('sidebarconfig.linkedAccounts'),
          description: t('sidebarconfig.linkedAccountsDescription'),
          attrs: {
            icon: 'link.svg',
          },
        },
      ],
    },
    {
      type: 'group',
      label: t('sidebarconfig.tsSdkReference'),
      homePageHref: localizePath('/docs/typescript-sdk', locale),
      category: NavigationCategory.TYPESCRIPT_SDK,
      autopopulateFromDir: localizePath('/docs/typescript-sdk', locale),
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/typescript-sdk', locale),
          label: t('sidebarconfig.tsSdkReference'),
          hideInSidebar: true,
        },
        {
          type: 'link',
          href: localizePath('/docs/typescript-sdk/daytona', locale),
          label: t('sidebarconfig.daytona'),
        },
        {
          type: 'link',
          href: localizePath('/docs/typescript-sdk/sandbox', locale),
          label: t('sidebarconfig.sandbox'),
        },
      ],
    },
    {
      type: 'group',
      label: t('sidebarconfig.common'),
      homePageHref: localizePath('/docs/python-sdk', locale),
      category: NavigationCategory.PYTHON_SDK,
      autopopulateFromDir: localizePath('/docs/python-sdk/common', locale),
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/python-sdk', locale),
          label: t('sidebarconfig.pythonSdkReference'),
          hideInSidebar: true,
        },
      ],
    },
    {
      type: 'group',
      label: t('sidebarconfig.syncPython'),
      homePageHref: localizePath('/docs/python-sdk', locale),
      category: NavigationCategory.PYTHON_SDK,
      autopopulateFromDir: localizePath('/docs/python-sdk/sync', locale),
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/python-sdk/sync/daytona', locale),
          label: t('sidebarconfig.daytona'),
        },
        {
          type: 'link',
          href: localizePath('/docs/python-sdk/sync/sandbox', locale),
          label: t('sidebarconfig.sandbox'),
        },
      ],
    },
    {
      type: 'group',
      label: t('sidebarconfig.asyncPython'),
      homePageHref: localizePath('/docs/python-sdk', locale),
      category: NavigationCategory.PYTHON_SDK,
      autopopulateFromDir: localizePath('/docs/python-sdk/async', locale),
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/python-sdk/async/async-daytona', locale),
          label: t('sidebarconfig.asyncDaytona'),
        },
        {
          type: 'link',
          href: localizePath('/docs/python-sdk/async/async-sandbox', locale),
          label: t('sidebarconfig.asyncSandbox'),
        },
      ],
    },
    {
      type: 'group',
      label: t('sidebarconfig.rubySdkReference'),
      homePageHref: localizePath('/docs/ruby-sdk', locale),
      category: NavigationCategory.RUBY_SDK,
      autopopulateFromDir: localizePath('/docs/ruby-sdk', locale),
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/ruby-sdk/daytona', locale),
          label: t('sidebarconfig.daytona'),
        },
        {
          type: 'link',
          href: localizePath('/docs/ruby-sdk/sandbox', locale),
          label: t('sidebarconfig.sandbox'),
        },
      ],
    },
    {
      type: 'group',
      label: t('sidebarconfig.goSdkReference'),
      homePageHref: localizePath('/docs/go-sdk', locale),
      category: NavigationCategory.GO_SDK,
      autopopulateFromDir: localizePath('/docs/go-sdk', locale),
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/go-sdk/daytona', locale),
          label: t('sidebarconfig.daytona'),
        },
      ],
    },
  ]
}
