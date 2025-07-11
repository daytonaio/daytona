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
  labels: Record<string, string>
): NavigationGroup[] => {
  if (!labels) return []
  return [
    {
      type: 'group',
      category: NavigationCategory.MAIN,
      entries: [
        {
          type: 'link',
          href: localizePath('/docs', locale),
          label: labels['sidebarconfig.documentation'],
          attrs: {
            icon: 'home.svg',
          },
          relatedGroupCategory: NavigationCategory.GENERAL,
        },
        {
          type: 'link',
          href: localizePath('/docs/typescript-sdk', locale),
          label: labels['sidebarconfig.tsSdkReference'],
          attrs: {
            icon: 'package.svg',
          },
          relatedGroupCategory: NavigationCategory.TYPESCRIPT_SDK,
        },
        {
          type: 'link',
          href: localizePath('/docs/python-sdk', locale),
          label: labels['sidebarconfig.pythonSdkReference'],
          attrs: {
            icon: 'package.svg',
          },
          relatedGroupCategory: NavigationCategory.PYTHON_SDK,
        },
        {
          type: 'link',
          href: localizePath('/docs/tools/api', locale),
          label: labels['sidebarconfig.apiReference'],
          disablePagination: true,
          attrs: {
            icon: 'server.svg',
          },
          relatedGroupCategory: NavigationCategory.GENERAL,
        },
        {
          type: 'link',
          href: localizePath('/docs/tools/cli', locale),
          label: labels['sidebarconfig.cliReference'],
          disablePagination: true,
          attrs: {
            icon: 'terminal.svg',
          },
          relatedGroupCategory: NavigationCategory.GENERAL,
        },
      ],
    },
    {
      type: 'group',
      label: labels['sidebarconfig.introduction'],
      homePageHref: localizePath('/docs', locale),
      category: NavigationCategory.GENERAL,
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/getting-started', locale),
          label: labels['sidebarconfig.gettingStarted'],
          description: labels['sidebarconfig.gettingStartedDescription'],
          attrs: {
            icon: 'bookmark.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/configuration', locale),
          label: labels['sidebarconfig.configuration'],
          description: labels['sidebarconfig.configurationDescription'],
          attrs: {
            icon: 'git-commit.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/sandbox-management', locale),
          label: labels['sidebarconfig.sandboxes'],
          description: labels['sidebarconfig.sandboxesDescription'],
          attrs: {
            icon: 'rectangle.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/snapshots', locale),
          label: labels['sidebarconfig.snapshots'],
          description: labels['sidebarconfig.snapshotsDescription'],
          attrs: {
            icon: 'layers.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/declarative-builder', locale),
          label: labels['sidebarconfig.declarativeBuilder'],
          description: labels['sidebarconfig.declarativeBuilderDescription'],
          attrs: {
            icon: 'prebuilds.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/volumes', locale),
          label: labels['sidebarconfig.volumes'],
          description: labels['sidebarconfig.volumesDescription'],
          attrs: {
            icon: 'container-registries.svg',
          },
        },
      ],
    },
    {
      type: 'group',
      label: labels['sidebarconfig.accountManagement'],
      homePageHref: localizePath('/docs', locale),
      category: NavigationCategory.GENERAL,
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/api-keys', locale),
          label: labels['sidebarconfig.apiKeys'],
          description: labels['sidebarconfig.apiKeysDescription'],
          attrs: {
            icon: 'tag.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/organizations', locale),
          label: labels['sidebarconfig.organizations'],
          description: labels['sidebarconfig.organizationsDescription'],
          attrs: {
            icon: 'building.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/limits', locale),
          label: labels['sidebarconfig.limits'],
          description: labels['sidebarconfig.limitsDescription'],
          attrs: {
            icon: 'log.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/billing', locale),
          label: labels['sidebarconfig.billing'],
          description: labels['sidebarconfig.billingDescription'],
          attrs: {
            icon: 'credit-card.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/linked-accounts', locale),
          label: labels['sidebarconfig.linkedAccounts'],
          description: labels['sidebarconfig.linkedAccountsDescription'],
          attrs: {
            icon: 'link.svg',
          },
        },
      ],
    },
    {
      type: 'group',
      label: labels['sidebarconfig.agentToolbox'],
      homePageHref: localizePath('/docs', locale),
      category: NavigationCategory.GENERAL,
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/file-system-operations', locale),
          label: labels['sidebarconfig.fileSystem'],
          description: labels['sidebarconfig.fileSystemDescription'],
          attrs: {
            icon: 'folder.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/git-operations', locale),
          label: labels['sidebarconfig.gitOperations'],
          description: labels['sidebarconfig.gitOperationsDescription'],
          attrs: {
            icon: 'git-branch.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/language-server-protocol', locale),
          label: labels['sidebarconfig.languageServerProtocol'],
          description:
            labels['sidebarconfig.languageServerProtocolDescription'],
          attrs: {
            icon: 'pulse.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/process-code-execution', locale),
          label: labels['sidebarconfig.processCodeExecution'],
          description: labels['sidebarconfig.processCodeExecutionDescription'],
          attrs: {
            icon: 'computer.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/log-streaming', locale),
          label: labels['sidebarconfig.logStreaming'],
          description: labels['sidebarconfig.logStreamingDescription'],
          attrs: {
            icon: 'log.svg',
          },
        },
      ],
    },
    {
      type: 'group',
      label: labels['sidebarconfig.other'],
      homePageHref: localizePath('/docs', locale),
      category: NavigationCategory.GENERAL,
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/web-terminal', locale),
          label: labels['sidebarconfig.webTerminal'],
          description: labels['sidebarconfig.webTerminalDescription'],
          attrs: {
            icon: 'terminal.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/preview-and-authentication', locale),
          label: labels['sidebarconfig.previewAuthentication'],
          description: labels['sidebarconfig.previewAuthenticationDescription'],
          attrs: {
            icon: 'shield.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/regions', locale),
          label: labels['sidebarconfig.regions'],
          description: labels['sidebarconfig.regionsDescription'],
          attrs: {
            icon: 'globe.svg',
          },
        },
        {
          type: 'link',
          href: localizePath('/docs/mcp', locale),
          label: labels['sidebarconfig.mcpServer'],
          disablePagination: true,
          attrs: {
            icon: 'server.svg',
          },
        },
      ],
    },
    {
      type: 'group',
      label: labels['sidebarconfig.tsSdkReference'],
      homePageHref: localizePath('/docs/typescript-sdk', locale),
      category: NavigationCategory.TYPESCRIPT_SDK,
      autopopulateFromDir: localizePath('/docs/typescript-sdk', locale),
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/typescript-sdk/daytona', locale),
          label: labels['sidebarconfig.daytona'],
        },
        {
          type: 'link',
          href: localizePath('/docs/typescript-sdk/sandbox', locale),
          label: labels['sidebarconfig.sandbox'],
        },
      ],
    },
    {
      type: 'group',
      label: labels['sidebarconfig.pythonSdkReference'],
      homePageHref: localizePath('/docs/python-sdk', locale),
      category: NavigationCategory.PYTHON_SDK,
    },
    {
      type: 'group',
      label: labels['sidebarconfig.common'],
      homePageHref: localizePath('/docs/python-sdk', locale),
      category: NavigationCategory.PYTHON_SDK,
      autopopulateFromDir: localizePath('/docs/python-sdk/common', locale),
    },
    {
      type: 'group',
      label: labels['sidebarconfig.syncPython'],
      homePageHref: localizePath('/docs/python-sdk', locale),
      category: NavigationCategory.PYTHON_SDK,
      autopopulateFromDir: localizePath('/docs/python-sdk/sync', locale),
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/python-sdk/sync/daytona', locale),
          label: labels['sidebarconfig.daytona'],
        },
        {
          type: 'link',
          href: localizePath('/docs/python-sdk/sync/sandbox', locale),
          label: labels['sidebarconfig.sandbox'],
        },
      ],
    },
    {
      type: 'group',
      label: labels['sidebarconfig.asyncPython'],
      homePageHref: localizePath('/docs/python-sdk', locale),
      category: NavigationCategory.PYTHON_SDK,
      autopopulateFromDir: localizePath('/docs/python-sdk/async', locale),
      entries: [
        {
          type: 'link',
          href: localizePath('/docs/python-sdk/async/async-daytona', locale),
          label: labels['sidebarconfig.asyncDaytona'],
        },
        {
          type: 'link',
          href: localizePath('/docs/python-sdk/async/async-sandbox', locale),
          label: labels['sidebarconfig.asyncSandbox'],
        },
      ],
    },
  ]
}
