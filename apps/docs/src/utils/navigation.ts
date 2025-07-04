// Navigation Configuration
// This file defines the relationship between main navigation items and their related pages
import fs from 'node:fs'
import path from 'node:path'

import { NavigationCategory } from '../content/config'

export interface NavigationItem {
  type: 'link' | 'group'
  label?: string
}

export interface NavigationLink extends NavigationItem {
  type: 'link'
  href: string
  label: string
  description?: string
  disablePagination?: boolean
  attrs?: {
    icon?: string
    [key: string]: any
  }
}

export interface MainNavigationLink extends NavigationLink {
  // All links with that category will be shown in the sidebar when the link is active
  relatedGroupCategory: NavigationCategory
}

export interface NavigationGroup extends NavigationItem {
  type: 'group'
  // The category of the group, all links with that category will be shown in the sidebar when
  // the link with that category or the main link that is related to the category is active
  category: NavigationCategory
  // Used to indicate the context of the current page.
  // If the current page is the first item in the list, it is also used as the previous link in the pagination component.
  // The referenced page should be a `MainNavigationLink`. It is ignored for `NavigationCategory.MAIN` groups.
  homePageHref?: string
  autopopulateFromDir?: string
  entries?: (NavigationLink | MainNavigationLink)[]
}

// HELPER FUNCTIONS

function normalizePath(path: string): string {
  return path.replace(/\/$/, '')
}

function getMainNavGroup(sidebarConfig: NavigationGroup[]): NavigationGroup {
  return sidebarConfig.find(
    group => group.category === NavigationCategory.MAIN
  ) as NavigationGroup
}

function getNavGroupsByCategory(
  sidebarConfig: NavigationGroup[],
  category: NavigationCategory
): NavigationGroup[] {
  return sidebarConfig.filter(group => group.category === category)
}

function getNavLinksByCategory(
  sidebarConfig: NavigationGroup[],
  category: NavigationCategory
): NavigationLink[] {
  const groups = getNavGroupsByCategory(sidebarConfig, category)
  return groups.flatMap(group => group.entries || []) as NavigationLink[]
}

function getNavGroupByHref(
  sidebarConfig: NavigationGroup[],
  href: string
): NavigationGroup | undefined {
  for (const group of sidebarConfig) {
    if (!group.entries) continue

    for (const entry of group.entries) {
      if (entry.type === 'link' && comparePaths(entry.href, href)) {
        return group
      }
    }
  }
  return undefined
}

function getNavLinkByHref(
  sidebarConfig: NavigationGroup[],
  href: string,
  group?: NavigationGroup
): NavigationLink | undefined {
  group = group ?? getNavGroupByHref(sidebarConfig, href)

  if (!group) return undefined

  if (!group.entries) return undefined

  return group.entries.find(
    entry => entry.type === 'link' && comparePaths(entry.href, href)
  ) as NavigationLink
}

function toCamelCase(filename: string): string {
  const nameWithoutExt = path.parse(filename).name

  return nameWithoutExt
    .split(/[-_\s]/)
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join('')
}

function populateEntriesFromDir(
  group: NavigationGroup,
  workspaceRoot: string = process.cwd()
): NavigationLink[] {
  if (!group.autopopulateFromDir) return group?.entries || []

  const dirPath = path.join(
    workspaceRoot,
    '/src/content',
    group.autopopulateFromDir
  )

  if (!fs.existsSync(dirPath)) {
    console.warn(
      `Directory ${dirPath} does not exist, cannot autopopulate navigation group`
    )
    return group?.entries || []
  }

  try {
    const files = fs.readdirSync(dirPath, { withFileTypes: true })

    const existingHrefs = new Set(
      (group.entries || [])
        .filter(entry => entry.type === 'link')
        .map(entry => (entry as NavigationLink).href)
    )

    const entries = files
      .filter(file => file.isFile())
      .filter(file => !['index.md', 'index.mdx'].includes(file.name))
      .filter(file => {
        const ext = path.extname(file.name).toLowerCase()
        return ext === '.md' || ext === '.mdx'
      })
      .map(file => {
        const fileName = path.parse(file.name).name
        const pathWithoutExt = path.join(
          group.autopopulateFromDir || '',
          fileName
        )
        const href = `${pathWithoutExt}`

        const label = toCamelCase(fileName)

        return {
          type: 'link',
          href,
          label,
        } as NavigationLink
      })
      .filter(entry => !existingHrefs.has(entry.href))

    return [...(group.entries || []), ...entries]
  } catch (error) {
    console.error(`Error reading directory ${dirPath}:`, error)
    return group?.entries || []
  }
}

function processAutopopulateGroups(sidebarConfig: NavigationGroup[]) {
  sidebarConfig.forEach(group => {
    if (group.autopopulateFromDir) {
      group.entries = populateEntriesFromDir(group)
      group.autopopulateFromDir = undefined
    }
  })
}

export function getPagination(
  sidebarConfig: NavigationGroup[],
  currentPath: string
): {
  prev?: { href: string; label: string }
  next?: { href: string; label: string }
} {
  processAutopopulateGroups(sidebarConfig)

  currentPath = currentPath.replace(/\/$/, '')
  const result: {
    prev?: { href: string; label: string }
    next?: { href: string; label: string }
  } = {}

  const link = getNavLinkByHref(sidebarConfig, currentPath)
  if (!link || link.disablePagination) return result

  const group = getNavGroupByHref(sidebarConfig, currentPath)

  if (!group) return result

  if (group.category === NavigationCategory.MAIN) {
    const links = getNavLinksByCategory(
      sidebarConfig,
      (link as MainNavigationLink).relatedGroupCategory
    )
    if (links && links.length > 0) {
      result.next = { href: links[0].href, label: links[0].label }
    }
  } else {
    const links = getNavLinksByCategory(sidebarConfig, group.category)

    const index = links.findIndex(link => comparePaths(link.href, currentPath))

    if (index === 0) {
      if (group.homePageHref) {
        const homePageLink = getNavLinkByHref(sidebarConfig, group.homePageHref)
        if (homePageLink) {
          result.prev = { href: homePageLink.href, label: homePageLink.label }
        }
      }
    } else {
      result.prev = {
        href: links[index - 1].href,
        label: links[index - 1].label,
      }
    }

    if (index != links.length - 1) {
      result.next = {
        href: links[index + 1].href,
        label: links[index + 1].label,
      }
    }
  }
  return result
}

export function getSidebar(
  sidebarConfig: NavigationGroup[],
  currentPath: string
): NavigationGroup[] {
  processAutopopulateGroups(sidebarConfig)

  currentPath = currentPath.replace(/\/$/, '')
  const mainGroup = getMainNavGroup(sidebarConfig)
  const currentGroup = getNavGroupByHref(sidebarConfig, currentPath)

  if (!currentGroup) return [mainGroup]

  let contextHref: string | null = null
  let relatedGroups: NavigationGroup[] = []

  if (currentGroup.category === NavigationCategory.MAIN) {
    const currentLink = getNavLinkByHref(
      sidebarConfig,
      currentPath,
      currentGroup
    ) as MainNavigationLink

    if (!currentLink) return [mainGroup]

    contextHref = currentPath
    relatedGroups = getNavGroupsByCategory(
      sidebarConfig,
      currentLink.relatedGroupCategory
    )
  } else {
    relatedGroups = getNavGroupsByCategory(sidebarConfig, currentGroup.category)
    contextHref = currentGroup.homePageHref || null
  }

  if (contextHref && mainGroup.entries) {
    mainGroup.entries = mainGroup.entries.map(entry => ({
      ...entry,
      context: comparePaths(entry.href, contextHref as string),
    }))
  }

  return [mainGroup, ...relatedGroups]
}

export function getExploreMoreData(
  sidebarConfig: NavigationGroup[],
  currentPath: string
) {
  processAutopopulateGroups(sidebarConfig)

  currentPath = currentPath.replace(/\/$/, '')
  const link = getNavLinkByHref(
    sidebarConfig,
    currentPath
  ) as MainNavigationLink
  if (!link) {
    return []
  }

  const relatedGroups = getNavGroupsByCategory(
    sidebarConfig,
    link.relatedGroupCategory
  )

  return relatedGroups.map(group => {
    const items = (group.entries || []).map(navLink => {
      return {
        title: navLink.label,
        subtitle: navLink.description || '',
        href: navLink.href,
      }
    })

    return {
      title: group.label || '',
      items,
    }
  })
}

export function comparePaths(path1: string, path2: string): boolean {
  return normalizePath(path1) === normalizePath(path2)
}
