/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Kbd,
  useCommandPalette,
  useCommandPaletteActions,
  useRegisterCommands,
  useRegisterPage,
  type CommandConfig,
} from '@/components/CommandPalette'
import { cn } from '@/lib/utils'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { liteClient as algoliasearch } from 'algoliasearch/lite'
import { BookOpen, Code2, Container, Layers, Terminal } from 'lucide-react'
import { ReactNode, useEffect, useMemo, useState } from 'react'

const ALGOLIA_APP_ID = import.meta.env.VITE_ALGOLIA_APP_ID
const ALGOLIA_API_KEY = import.meta.env.VITE_ALGOLIA_API_KEY
const DOCS_INDEX = import.meta.env.VITE_ALGOLIA_DOCS_INDEX_NAME || 'docs_test'
const CLI_INDEX = import.meta.env.VITE_ALGOLIA_CLI_INDEX_NAME || 'cli_test'
const SDK_INDEX = import.meta.env.VITE_ALGOLIA_SDK_INDEX_NAME || 'sdk_test'

const docSearchEnabled = Boolean(ALGOLIA_APP_ID && ALGOLIA_API_KEY)
const client = docSearchEnabled ? algoliasearch(ALGOLIA_APP_ID, ALGOLIA_API_KEY) : null

export type AlgoliaHit = {
  objectID: string
  url: string
  slug: string
  title: string
  description?: string
  content?: string
  _highlightResult?: {
    title?: { value: string; matchLevel: string }
    description?: { value: string; matchLevel: string }
  }
}

export type SearchResults = {
  docs: AlgoliaHit[]
  cli: AlgoliaHit[]
  sdk: AlgoliaHit[]
}

export const searchDocumentation = async (query: string): Promise<SearchResults> => {
  if (!client || !query.trim()) {
    return { docs: [], cli: [], sdk: [] }
  }

  const commonParams = {
    hitsPerPage: 3,
    attributesToHighlight: ['title', 'description'],
    highlightPreTag: '<em>',
    highlightPostTag: '</em>',
  }

  const { results } = await client.search({
    requests: [
      { indexName: DOCS_INDEX, query, ...commonParams },
      { indexName: CLI_INDEX, query, ...commonParams },
      { indexName: SDK_INDEX, query, ...commonParams },
    ],
  })

  const getHits = (index: number) =>
    results[index] && 'hits' in results[index] ? (results[index].hits as AlgoliaHit[]) : []

  return {
    docs: getHits(0),
    cli: getHits(1),
    sdk: getHits(2),
  }
}

function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value)
  useEffect(() => {
    const handler = setTimeout(() => setDebouncedValue(value), delay)
    return () => clearTimeout(handler)
  }, [value, delay])
  return debouncedValue
}

export const useDocsSearchQuery = ({ search, enabled }: { search: string; enabled: boolean }) => {
  return useQuery({
    queryKey: ['algolia-search', search],
    queryFn: () => searchDocumentation(search),
    enabled: enabled && search.length > 1,
    staleTime: 1000 * 60 * 5,
    placeholderData: keepPreviousData,
  })
}

const openDocs = (path = '') => {
  window.open(`https://www.daytona.io/docs/${path}`, '_blank')
}

const SearchSnippet = ({
  hit,
  attribute,
  className,
}: {
  hit: AlgoliaHit
  attribute: 'title' | 'description'
  className?: string
}) => {
  const content = hit._highlightResult?.[attribute]?.value || hit[attribute] || ''

  return (
    <span
      className={cn(
        '[&_em]:not-italic [&_em]:rounded [&_em]:bg-[#2fcc712b] [&_em]:text-[#058157] dark:[&_em]:text-[#2fcc71] [&_em]:px-[2px]',
        className,
      )}
      dangerouslySetInnerHTML={{ __html: content }}
    />
  )
}

const ResultRow = ({ hit, tag }: { hit: AlgoliaHit; tag?: ReactNode }) => (
  <div className="flex flex-col overflow-hidden">
    <div className="flex items-center gap-2">
      <SearchSnippet hit={hit} attribute="title" className="font-mediumline-clamp-1" />
      {tag}
    </div>
    {hit.description && (
      <SearchSnippet hit={hit} attribute="description" className="text-xs text-muted-foreground line-clamp-1" />
    )}
  </div>
)

// todo: something more robust here
const parseSDKLanguage = (hit: AlgoliaHit) => {
  const [, lang] = hit.slug.split('/')
  if (lang.includes('python')) {
    return 'Python'
  }
  if (lang.includes('typescript')) {
    return 'TypeScript'
  }

  return lang
}

export function useDocsSearchCommands() {
  const activePageId = useCommandPalette((state) => state.activePageId)
  const search = useCommandPalette((state) => state.searchByPage.get('search-docs') ?? '')

  const { setShouldFilter, setIsLoading } = useCommandPaletteActions()
  const enabled = activePageId === 'search-docs' && docSearchEnabled
  const debouncedQuery = useDebounce(search, 300)

  const { data, isError, isFetching } = useDocsSearchQuery({
    search: debouncedQuery,
    enabled,
  })

  useRegisterPage({ id: 'search-docs', label: 'Search Docs', placeholder: 'Search documentation...' })

  useEffect(() => {
    if (!enabled) {
      return
    }
    setShouldFilter(false)
    return () => setShouldFilter(true)
  }, [enabled, setShouldFilter])

  useEffect(() => {
    if (!enabled) {
      return
    }
    setIsLoading(isFetching)
  }, [enabled, isFetching, setIsLoading])

  const commands: CommandConfig[] = useMemo(() => {
    const handleSelect = (hit: AlgoliaHit) => {
      const url = hit.url || `https://www.daytona.io/${hit.slug}`
      window.open(url, '_blank')
    }

    if (!search || !data) {
      return [
        {
          id: 'suggestion-quickstart',
          label: 'Quick Start',
          icon: <BookOpen className="w-4 h-4" />,
          onSelect: () => openDocs(),
          chainable: true,
        },
        {
          id: 'suggestion-sandboxes',
          label: 'Sandboxes',
          icon: <Container className="w-4 h-4" />,
          onSelect: () => openDocs('/en/sandboxes'),
        },
        {
          id: 'suggestion-snapshots',
          label: 'Snapshots',
          icon: <Layers className="w-4 h-4" />,
          onSelect: () => openDocs('/en/snapshots'),
          chainable: true,
        },
        {
          id: 'suggestion-limits',
          label: 'Limits',
          icon: <Terminal className="w-4 h-4" />,
          onSelect: () => openDocs('/en/limits'),
          chainable: true,
        },
      ]
    }

    if (isError) {
      return [
        {
          id: 'error',
          label: 'Failed to load documentation. Try again.',
          disabled: true,
        },
      ]
    }

    const results: CommandConfig[] = []

    for (const hit of data.docs) {
      results.push({
        id: `docs-${hit.objectID}`,
        label: <ResultRow hit={hit} />,
        value: `docs ${hit.title} ${hit.description || ''}`,
        icon: <BookOpen className="w-4 h-4" />,
        onSelect: () => handleSelect(hit),
        chainable: true,
        className: 'py-2',
      })
    }

    for (const hit of data.cli) {
      results.push({
        id: `cli-${hit.objectID}`,
        label: <ResultRow hit={hit} />,
        value: `cli ${hit.title} ${hit.description || ''}`,
        icon: <Terminal className="w-4 h-4" />,
        onSelect: () => handleSelect(hit),
        chainable: true,
        className: 'py-2',
      })
    }

    for (const hit of data.sdk) {
      const sdkLanguage = parseSDKLanguage(hit)

      results.push({
        id: `sdk-${hit.objectID}`,
        label: <ResultRow hit={hit} tag={<Kbd className="text-xs h-auto">{sdkLanguage}</Kbd>} />,
        value: `sdk ${hit.title} ${hit.description || ''} ${sdkLanguage}`,
        icon: <Code2 className="w-4 h-4" />,
        onSelect: () => handleSelect(hit),
        chainable: true,
        className: 'py-2',
      })
    }

    return results
  }, [search, data, isError])

  useRegisterCommands(commands, {
    pageId: 'search-docs',
    groupId: 'docs-results',
    groupLabel: !search ? 'Suggestions' : data?.docs.length ? 'Results' : undefined,
    groupOrder: 0,
  })
}
