import { docsSchema, i18nSchema } from '@astrojs/starlight/schema'
import { defineCollection, z } from 'astro:content'

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
  i18n: defineCollection({ type: 'data', schema: i18nSchema() }),
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
export const sidebarConfig: NavigationGroup[] = [
  {
    type: 'group',
    category: NavigationCategory.MAIN,
    entries: [
      {
        type: 'link',
        href: '/docs',
        label: 'Documentation',
        attrs: {
          icon: 'home.svg',
        },
        relatedGroupCategory: NavigationCategory.GENERAL,
      },
      {
        type: 'link',
        href: '/docs/typescript-sdk',
        label: 'TS SDK Reference',
        attrs: {
          icon: 'package.svg',
        },
        relatedGroupCategory: NavigationCategory.TYPESCRIPT_SDK,
      },
      {
        type: 'link',
        href: '/docs/python-sdk',
        label: 'Python SDK Reference',
        attrs: {
          icon: 'package.svg',
        },
        relatedGroupCategory: NavigationCategory.PYTHON_SDK,
      },
      {
        type: 'link',
        href: '/docs/tools/api',
        label: 'API Reference',
        disablePagination: true,
        attrs: {
          icon: 'server.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/tools/cli',
        label: 'CLI Reference',
        disablePagination: true,
        attrs: {
          icon: 'terminal.svg',
        },
      },
    ],
  },
  {
    type: 'group',
    label: 'Introduction',
    homePageHref: '/docs',
    category: NavigationCategory.GENERAL,
    entries: [
      {
        type: 'link',
        href: '/docs',
        label: 'Home',
        attrs: {
          icon: 'home.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/getting-started',
        label: 'Getting Started',
        description:
          'Learn about Daytona SDK and how it can help you manage your development environments.',
        attrs: {
          icon: 'bookmark.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/configuration',
        label: 'Configuration',
        description:
          'Get started with Daytona SDK and learn how to use and configure your development environments.',
        attrs: {
          icon: 'git-commit.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/sandbox-management',
        label: 'Sandboxes',
        description:
          'Learn how to create, manage, and remove Sandboxes using the Daytona SDK.',
        attrs: {
          icon: 'rectangle.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/snapshots',
        label: 'Snapshots',
        description:
          'Learn how to create, manage and remove Snapshots using the Daytona SDK.',
        attrs: {
          icon: 'layers.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/declarative-builder',
        label: 'Declarative Builder',
        description:
          'Learn how to dynamically build Snapshots from Docker/OCI compatible images using the Daytona SDK.',
        attrs: {
          icon: 'prebuilds.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/volumes',
        label: 'Volumes',
        description: 'Learn how to manage volumes in your Daytona Sandboxes.',
        attrs: {
          icon: 'container-registries.svg',
        },
      },
    ],
  },
  {
    type: 'group',
    label: 'Account management',
    homePageHref: '/docs',
    category: NavigationCategory.GENERAL,
    entries: [
      {
        type: 'link',
        href: '/docs/api-keys',
        label: 'API Keys',
        description: 'Daytona API Key management and scopes.',
        attrs: {
          icon: 'tag.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/organizations',
        label: 'Organizations',
        description:
          'Learn how to create, manage, and remove Organizations using the Daytona SDK.',
        attrs: {
          icon: 'building.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/limits',
        label: 'Limits',
        description: 'Limits and tiers assigned to Organizations.',
        attrs: {
          icon: 'log.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/billing',
        label: 'Billing',
        description: 'Billing management for Organizations.',
        attrs: {
          icon: 'credit-card.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/linked-accounts',
        label: 'Linked Accounts',
        description: 'Linked Accounts for Users.',
        attrs: {
          icon: 'link.svg',
        },
      },
    ],
  },
  {
    type: 'group',
    label: 'Agent Toolbox',
    homePageHref: '/docs',
    category: NavigationCategory.GENERAL,
    entries: [
      {
        type: 'link',
        href: '/docs/file-system-operations',
        label: 'File System',
        description:
          'Learn how to manage files and directories in your Sandboxes using the Daytona SDK.',
        attrs: {
          icon: 'folder.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/git-operations',
        label: 'Git Operations',
        description:
          'Learn how to manage Git repositories in your Sandboxes using the Daytona SDK.',
        attrs: {
          icon: 'git-branch.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/language-server-protocol',
        label: 'Language Server Protocol',
        description:
          'Learn how to use Language Server Protocol (LSP) support in your Sandboxes using the Daytona SDK.',
        attrs: {
          icon: 'pulse.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/process-code-execution',
        label: 'Process & Code Execution',
        description:
          'Learn about running commands and code in isolated environments using the Daytona SDK.',
        attrs: {
          icon: 'computer.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/log-streaming',
        label: 'Log Streaming',
        description:
          'Learn how to stream logs from your Sandboxes using the Daytona SDK.',
        attrs: {
          icon: 'log.svg',
        },
      },
    ],
  },
  {
    type: 'group',
    label: 'Other',
    homePageHref: '/docs',
    category: NavigationCategory.GENERAL,
    entries: [
      {
        type: 'link',
        href: '/docs/web-terminal',
        label: 'Web Terminal',
        description: 'Web Terminal access to Daytona Sandboxes.',
        attrs: {
          icon: 'terminal.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/preview-and-authentication',
        label: 'Preview & Authentication',
        description: 'Preview URLs and authentication tokens.',
        attrs: {
          icon: 'shield.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/regions',
        label: 'Regions',
        description: 'Setting the region to spin up Daytona Sandboxes in.',
        attrs: {
          icon: 'globe.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/mcp',
        label: 'MCP Server',
        disablePagination: true,
        attrs: {
          icon: 'server.svg',
        },
      },
      {
        type: 'link',
        href: '/docs/data-analysis-with-ai',
        label: 'Data Analysis with AI',
        disablePagination: true,
        attrs: {
          icon: 'chart.svg',
        },
      },
    ],
  },
  {
    type: 'group',
    label: 'TS SDK Reference',
    homePageHref: '/docs/typescript-sdk',
    category: NavigationCategory.TYPESCRIPT_SDK,
    autopopulateFromDir: '/docs/typescript-sdk',
    entries: [
      {
        type: 'link',
        href: '/docs/typescript-sdk',
        label: 'Home',
      },
      {
        type: 'link',
        href: '/docs/typescript-sdk/daytona',
        label: 'Daytona',
      },
      {
        type: 'link',
        href: '/docs/typescript-sdk/sandbox',
        label: 'Sandbox',
      },
    ],
  },
  {
    type: 'group',
    label: 'Python SDK Reference',
    homePageHref: '/docs/python-sdk',
    category: NavigationCategory.PYTHON_SDK,
    entries: [
      {
        type: 'link',
        href: '/docs/python-sdk',
        label: 'Home',
      },
    ],
  },
  {
    type: 'group',
    label: 'Common',
    homePageHref: '/docs/python-sdk',
    category: NavigationCategory.PYTHON_SDK,
    autopopulateFromDir: '/docs/python-sdk/common',
  },
  {
    type: 'group',
    label: 'Sync Python',
    homePageHref: '/docs/python-sdk',
    category: NavigationCategory.PYTHON_SDK,
    autopopulateFromDir: '/docs/python-sdk/sync',
    entries: [
      {
        type: 'link',
        href: '/docs/python-sdk/sync/daytona',
        label: 'Daytona',
      },
      {
        type: 'link',
        href: '/docs/python-sdk/sync/sandbox',
        label: 'Sandbox',
      },
    ],
  },
  {
    type: 'group',
    label: 'Async Python',
    homePageHref: '/docs/python-sdk',
    category: NavigationCategory.PYTHON_SDK,
    autopopulateFromDir: '/docs/python-sdk/async',
    entries: [
      {
        type: 'link',
        href: '/docs/python-sdk/async/async-daytona',
        label: 'AsyncDaytona',
      },
      {
        type: 'link',
        href: '/docs/python-sdk/async/async-sandbox',
        label: 'AsyncSandbox',
      },
    ],
  },
]
