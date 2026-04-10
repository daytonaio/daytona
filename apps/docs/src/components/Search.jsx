import { liteClient as algoliasearch } from 'algoliasearch/lite'
import { GTProvider, useGT } from 'gt-react'
import { useCallback, useEffect, useRef, useState } from 'react'
import {
  Configure,
  Highlight,
  Index,
  InstantSearch,
  Pagination,
  SearchBox,
  connectHits,
  connectStats,
} from 'react-instantsearch-dom'
import loadTranslations from 'src/i18n/loadTranslations'

import gtConfig from '../../gt.config.json'
import '../styles/components/search.scss'

const ALGOLIA_APP_ID = import.meta.env.PUBLIC_ALGOLIA_APP_ID || null
const ALGOLIA_API_KEY = import.meta.env.PUBLIC_ALGOLIA_API_KEY || null
const DOCS_INDEX_NAME =
  import.meta.env.PUBLIC_ALGOLIA_DOCS_INDEX_NAME || 'docs'
const CLI_INDEX_NAME = import.meta.env.PUBLIC_ALGOLIA_CLI_INDEX_NAME || 'cli'
const SDK_INDEX_NAME = import.meta.env.PUBLIC_ALGOLIA_SDK_INDEX_NAME || 'sdk'
const API_INDEX_NAME = import.meta.env.PUBLIC_ALGOLIA_API_INDEX_NAME || 'api'

const SEARCH_HITS_PER_PAGE = 30

const QUERY_STOPWORDS = new Set([
  'a',
  'an',
  'and',
  'are',
  'as',
  'at',
  'be',
  'by',
  'for',
  'from',
  'in',
  'is',
  'it',
  'of',
  'on',
  'or',
  'the',
  'to',
  'with',
])

const searchClient =
  ALGOLIA_APP_ID && ALGOLIA_API_KEY
    ? algoliasearch(ALGOLIA_APP_ID, ALGOLIA_API_KEY)
    : null

function getSortedIndexes() {
  const path = typeof window !== 'undefined' ? window.location.pathname : ''
  const isApiPage = path.includes('/tools/api')
  const isSdkPage = path.includes('/typescript-sdk') || path.includes('/python-sdk')
  const isCliPage = path.includes('/tools/cli')

  if (isApiPage) {
    return [API_INDEX_NAME, DOCS_INDEX_NAME, CLI_INDEX_NAME, SDK_INDEX_NAME]
  }
  if (isSdkPage) {
    return [SDK_INDEX_NAME, DOCS_INDEX_NAME, CLI_INDEX_NAME, API_INDEX_NAME]
  }
  if (isCliPage) {
    return [CLI_INDEX_NAME, DOCS_INDEX_NAME, SDK_INDEX_NAME, API_INDEX_NAME]
  }
  return [DOCS_INDEX_NAME, CLI_INDEX_NAME, SDK_INDEX_NAME, API_INDEX_NAME]
}

const INDEX_LABELS = {
  [DOCS_INDEX_NAME]: 'Documentation',
  [CLI_INDEX_NAME]: 'CLI',
  [SDK_INDEX_NAME]: 'SDK',
  [API_INDEX_NAME]: 'API',
}

function humanizeSegment(segment) {
  return segment
    .replace(/_/g, ' ')
    .split(/[\s-]+/)
    .filter(Boolean)
    .map(w => w.charAt(0).toUpperCase() + w.slice(1).toLowerCase())
    .join(' ')
}

function firstApiTagSegment(tags) {
  if (tags == null || tags === '') {
    return ''
  }
  if (typeof tags === 'string') {
    return tags.split(',')[0].trim()
  }
  if (Array.isArray(tags)) {
    const first = tags.find(t => t != null && String(t).trim())
    return first != null ? String(first).trim() : ''
  }
  return ''
}

function docPathKey(hit) {
  const s = hit.slug
  if (typeof s === 'string' && s.length > 0) {
    const hash = s.indexOf('#')
    const base = hash >= 0 ? s.slice(0, hash) : s
    const normalized = base.replace(/^\/+/, '').replace(/\/+$/, '')
    return normalized || `__object:${hit.objectID}`
  }
  return `__object:${hit.objectID}`
}

function queryTokens(query) {
  const q = query.trim().toLowerCase()
  if (!q) {
    return []
  }
  const split = q.split(/\s+/).filter(t => t.length >= 2)
  const filtered = split.filter(t => !QUERY_STOPWORDS.has(t))
  return filtered.length > 0 ? filtered : split
}

function countSubstringOccurrences(haystack, needle) {
  if (!haystack || !needle) {
    return 0
  }
  let count = 0
  let pos = 0
  while ((pos = haystack.indexOf(needle, pos)) !== -1) {
    count++
    pos += needle.length
  }
  return count
}

function cappedDescriptionTokenScore(description, token, cap = 26) {
  if (!description.includes(token)) {
    return 0
  }
  const occ = countSubstringOccurrences(description, token)
  return Math.min(8 + 5 * occ, cap)
}

function relevanceScore(hit, tokens, queryLower) {
  if (tokens.length === 0 && !queryLower) {
    return 0
  }

  const pathKey = docPathKey(hit).toLowerCase().replace(/_/g, '-')
  const segments = pathKey.split('/').filter(Boolean)
  const depth = Math.max(0, segments.length - 1)
  const lastSeg = segments.length > 0 ? segments[segments.length - 1] : pathKey

  const title = (hit.title || '').toLowerCase()
  const desc = (hit.description || '').toLowerCase()
  const contentPeek = (hit.content || '').slice(0, 320).toLowerCase()

  let score = 0

  for (const tok of tokens) {
    if (lastSeg.includes(tok)) {
      score += 92
    } else if (pathKey.includes(tok)) {
      score += 26
    }
    if (title.includes(tok)) {
      score += 36
    }
    score += cappedDescriptionTokenScore(desc, tok, 26)
    if (contentPeek.includes(tok)) {
      score += Math.min(6, 3 + countSubstringOccurrences(contentPeek, tok) * 2)
    }
  }

  const lastSegmentMatchesQuery = tokens.some(t => lastSeg.includes(t))
  if (lastSegmentMatchesQuery) {
    score += Math.max(0, 68 - depth * 16)
  }

  if (hit.recordType === 'page' && lastSegmentMatchesQuery) {
    score += 48
  }

  if (queryLower.length > 3 && title.includes(queryLower)) {
    score += 52
  }

  const qParts = queryLower.split(/\s+/).filter(p => p.length > 1)
  if (qParts.length >= 2) {
    const allTokensInTitle = qParts.every(p => title.includes(p))
    if (allTokensInTitle) {
      score += 42
    }
  }

  return score
}

function pickPrimaryDocKey(hits, query) {
  const queryLower = query.trim().toLowerCase()
  const tokens = queryTokens(query)

  const groupScore = new Map()
  const groupMinIndex = new Map()

  hits.forEach((hit, originalIndex) => {
    const key = docPathKey(hit)
    const rowScore = relevanceScore(hit, tokens, queryLower)
    groupScore.set(key, Math.max(groupScore.get(key) || 0, rowScore))
    groupMinIndex.set(
      key,
      Math.min(groupMinIndex.get(key) ?? Infinity, originalIndex)
    )
  })

  let bestKey = null
  let bestScore = -1
  let bestMinIdx = Infinity
  for (const [key, sc] of groupScore) {
    const idx = groupMinIndex.get(key)
    if (sc > bestScore || (sc === bestScore && idx < bestMinIdx)) {
      bestScore = sc
      bestKey = key
      bestMinIdx = idx
    }
  }

  if (bestScore > 0 && bestKey != null) {
    return bestKey
  }

  const firstPage = hits.find(h => h.recordType === 'page')
  if (firstPage) {
    return docPathKey(firstPage)
  }
  return docPathKey(hits[0])
}

function sortHitsByDocumentStructure(hits, query = '') {
  if (!hits || hits.length < 2) {
    return hits
  }

  const primaryDocKey = pickPrimaryDocKey(hits, query)
  const queryLower = query.trim().toLowerCase()
  const tokens = queryTokens(query)

  const annotated = hits.map((hit, originalIndex) => ({ hit, originalIndex }))
  const groups = new Map()
  for (const row of annotated) {
    const key = docPathKey(row.hit)
    if (!groups.has(key)) {
      groups.set(key, [])
    }
    groups.get(key).push(row)
  }

  const groupKeys = [...groups.keys()].sort((a, b) => {
    if (a === primaryDocKey) {
      return -1
    }
    if (b === primaryDocKey) {
      return 1
    }
    const minA = Math.min(...groups.get(a).map(r => r.originalIndex))
    const minB = Math.min(...groups.get(b).map(r => r.originalIndex))
    return minA - minB
  })

  const ordered = []
  for (const key of groupKeys) {
    const members = groups.get(key)
    members.sort((a, b) => {
      const pa = a.hit.recordType === 'page' ? -1 : 0
      const pb = b.hit.recordType === 'page' ? -1 : 0
      if (pa !== pb) {
        return pa - pb
      }
      const ra = relevanceScore(a.hit, tokens, queryLower)
      const rb = relevanceScore(b.hit, tokens, queryLower)
      if (ra !== rb) {
        return rb - ra
      }
      const ha =
        typeof a.hit.headingOrder === 'number'
          ? a.hit.headingOrder
          : Number.MAX_SAFE_INTEGER
      const hb =
        typeof b.hit.headingOrder === 'number'
          ? b.hit.headingOrder
          : Number.MAX_SAFE_INTEGER
      if (ha !== hb) {
        return ha - hb
      }
      return a.originalIndex - b.originalIndex
    })
    ordered.push(...members.map(r => r.hit))
  }

  const positions = hits
    .map(h => h.__position)
    .filter(p => typeof p === 'number' && Number.isFinite(p))
  if (positions.length > 0) {
    const minPos = Math.min(...positions)
    return ordered.map((hit, i) => ({
      ...hit,
      __position: minPos + i,
    }))
  }

  return ordered
}

function formatSearchBreadcrumb(indexName, hit) {
  const root = INDEX_LABELS[indexName] ?? hit.tag ?? 'Search'
  if (indexName === API_INDEX_NAME && hit.tags) {
    const firstTag = firstApiTagSegment(hit.tags)
    if (firstTag) {
      return `${root} > ${humanizeSegment(firstTag)}`
    }
  }
  const slug = (hit.slug || '').replace(/^\//, '')
  if (!slug) {
    return root
  }
  const rawParts = slug.split(/[/+#]+/).filter(Boolean)
  const humanized = rawParts.slice(0, 3).map(humanizeSegment)
  if (humanized.length === 0) {
    return root
  }
  return [root, ...humanized].join(' > ')
}

function SearchContent() {
  const [isSearchVisible, setIsSearchVisible] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [debounceQuery, setDebounceQuery] = useState('')
  const [displayHits, setDisplayHits] = useState(false)
  const [totalHits, setTotalHits] = useState(0)
  const [sortedIndexes, setSortedIndexes] = useState(getSortedIndexes())
  const [selectedIndexFilter, setSelectedIndexFilter] = useState(null)
  const [indexesWithHits, setIndexesWithHits] = useState([])
  const debounceTimeoutRef = useRef(null)
  const searchWrapperRef = useRef(null)
  const t = useGT()

  const currentPath =
    typeof window !== 'undefined' ? window.location.pathname : ''

  useEffect(() => {
    if (isSearchVisible) {
      setSortedIndexes(getSortedIndexes())
    }
  }, [isSearchVisible, currentPath])

  useEffect(() => {
    const toggleSearch = () => {
      setIsSearchVisible(prev => {
        if (prev) {
          setSearchQuery('')
          setDebounceQuery('')
          setDisplayHits(false)
          setTotalHits(0)
          setSelectedIndexFilter(null)
          setIndexesWithHits([])
        }
        return !prev
      })
    }

    const handleKeyDown = event => {
      if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
        event.preventDefault()
        toggleSearch()
      } else if (event.key === 'Escape') {
        setIsSearchVisible(false)
        setSearchQuery('')
        setDebounceQuery('')
        setDisplayHits(false)
        setTotalHits(0)
        setSelectedIndexFilter(null)
        setIndexesWithHits([])
      }
    }

    const handleSearchClick = event => {
      if (event.target.closest('.search-click')) {
        event.preventDefault()
        event.stopPropagation()
        toggleSearch()
      }
    }

    const handleClickOutside = event => {
      if (
        searchWrapperRef.current &&
        !searchWrapperRef.current.contains(event.target) &&
        !event.target.closest('.search-click')
      ) {
        setIsSearchVisible(false)
        setSearchQuery('')
        setDebounceQuery('')
        setDisplayHits(false)
        setTotalHits(0)
        setSelectedIndexFilter(null)
        setIndexesWithHits([])
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    document.addEventListener('click', handleSearchClick)
    document.addEventListener('mousedown', handleClickOutside)

    return () => {
      document.removeEventListener('keydown', handleKeyDown)
      document.removeEventListener('click', handleSearchClick)
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [])

  useEffect(() => {
    if (isSearchVisible && debounceQuery && displayHits) {
      document.body.classList.add('no-scroll')
    } else {
      document.body.classList.remove('no-scroll')
    }
  }, [isSearchVisible, debounceQuery, displayHits])

  useEffect(() => {
    if (debounceTimeoutRef.current) {
      clearTimeout(debounceTimeoutRef.current)
    }

    debounceTimeoutRef.current = setTimeout(() => {
      setTotalHits(0)
      setDebounceQuery(searchQuery)
      setSelectedIndexFilter(null)
    }, 400)

    return () => {
      if (debounceTimeoutRef.current) {
        clearTimeout(debounceTimeoutRef.current)
      }
    }
  }, [searchQuery])

  const handleIndexHitsChange = useCallback((indexName, nbHits) => {
    setIndexesWithHits(prev =>
      nbHits > 0
        ? prev.includes(indexName)
          ? prev
          : [...prev, indexName]
        : prev.filter(i => i !== indexName)
    )
  }, [])

  return (
    isSearchVisible && (
      <>
        <div className="search-modal-backdrop" aria-hidden="true" />
        <div
          id="searchbox-wrapper"
          className="searchbox-wrapper"
          ref={searchWrapperRef}
          role="dialog"
          aria-label={t('Search documentation', {
            $context: 'As in a search bar on a website',
          })}
        >
          <InstantSearch indexName={DOCS_INDEX_NAME} searchClient={searchClient}>
          <div className="search-bar-container">
            <div className="search-input-shell">
              <SearchBox
                translations={{
                  placeholder: t('Search documentation', {
                    $context: 'As in a search bar on a website',
                  }),
                }}
                autoFocus
                onChange={event => setSearchQuery(event.currentTarget.value)}
                value={searchQuery}
              />
              <kbd className="search-input-shell__esc">Esc</kbd>
            </div>
          </div>
          <div className="search-content">
            {debounceQuery && (
              <>
                <div className="search-index-filters">
                  <button
                    type="button"
                    className={`search-index-filter-chip ${selectedIndexFilter === null ? 'active' : ''}`}
                    onClick={() => setSelectedIndexFilter(null)}
                  >
                    All
                  </button>
                  {sortedIndexes
                    .filter(indexName => indexesWithHits.includes(indexName))
                    .map(indexName => (
                      <button
                        key={indexName}
                        type="button"
                        className={`search-index-filter-chip ${selectedIndexFilter === indexName ? 'active' : ''}`}
                        onClick={() => setSelectedIndexFilter(indexName)}
                      >
                        {INDEX_LABELS[indexName] ?? indexName}
                      </button>
                    ))}
                </div>
                {sortedIndexes.map(indexName => (
                  <div
                    key={`${indexName}-${debounceQuery}`}
                    className="search-index-panel"
                    hidden={selectedIndexFilter !== null && selectedIndexFilter !== indexName}
                  >
                    <SearchIndex
                      indexName={indexName}
                      setDisplayHits={setDisplayHits}
                      setIsSearchVisible={setIsSearchVisible}
                      setTotalHits={setTotalHits}
                      onIndexHitsChange={handleIndexHitsChange}
                      debounceQuery={debounceQuery}
                    />
                  </div>
                ))}
                {(totalHits === 0 ||
                  (selectedIndexFilter !== null && !indexesWithHits.includes(selectedIndexFilter))) && (
                  <div className="search-empty">
                    No results found for &quot;<strong>{debounceQuery}</strong>&quot;
                  </div>
                )}
              </>
            )}
          </div>
          </InstantSearch>
        </div>
      </>
    )
  )
}

function SearchIndex({ indexName, setDisplayHits, setIsSearchVisible, setTotalHits, onIndexHitsChange, debounceQuery }) {
  return (
    <Index indexName={indexName}>
      <Configure
        hitsPerPage={SEARCH_HITS_PER_PAGE}
        clickAnalytics
        getRankingInfo={false}
        optionalFilters={['recordType:page']}
      />
      <ConditionalSearchIndex
        indexName={indexName}
        setDisplayHits={setDisplayHits}
        setIsSearchVisible={setIsSearchVisible}
        setTotalHits={setTotalHits}
        onIndexHitsChange={onIndexHitsChange}
        debounceQuery={debounceQuery}
      />
    </Index>
  )
}

const ConditionalSearchIndexComponent = ({ indexName, setDisplayHits, setIsSearchVisible, nbHits, setTotalHits, onIndexHitsChange, debounceQuery }) => {
  useEffect(() => {
    setDisplayHits(nbHits > 0)
    setTotalHits(prev => prev + nbHits)
  }, [nbHits, setDisplayHits, setTotalHits, debounceQuery])

  useEffect(() => {
    if (onIndexHitsChange && typeof nbHits === 'number') {
      onIndexHitsChange(indexName, nbHits)
    }
  }, [indexName, nbHits, onIndexHitsChange])

  if (nbHits === 0) {
    return null
  }

  return (
    <div data-index={indexName} className="search-index-block">
      <div className="stats-pagination-wrapper">
        <Stats setDisplayHits={setDisplayHits} indexName={indexName} />
        <Pagination
          showFirst={false}
          showPrevious
          showNext
          showLast={false}
          padding={1}
        />
      </div>
      <DocOrderedHits
        debounceQuery={debounceQuery}
        setIsSearchVisible={setIsSearchVisible}
        indexName={indexName}
      />
    </div>
  )
}

const ConditionalSearchIndex = connectStats(ConditionalSearchIndexComponent)

function Hit({ hit, setIsSearchVisible, indexName }) {
  const isDocsFamily =
    [DOCS_INDEX_NAME, CLI_INDEX_NAME, SDK_INDEX_NAME, API_INDEX_NAME].includes(
      indexName
    ) || indexName === 'website'
  const isBlogs = indexName === 'blogs_test'
  const isSection =
    !isBlogs &&
    isDocsFamily &&
    typeof hit.url === 'string' &&
    hit.url.includes('#')

  const handleClick = e => {
    e.preventDefault()
    let hitUrl = hit.url

    if (indexName === 'blogs_test') {
      hitUrl = `https://www.daytona.io/dotfiles/${hit.slug}`
    } else if (indexName === 'website') {
      hitUrl = `https://www.daytona.io/${hit.slug}`
    }

    const currentUrl = window.location.href

    if (currentUrl.includes(hitUrl)) {
      const element = document.querySelector(`[data-slug='${hit.slug}']`)
      if (element) {
        element.scrollIntoView({ behavior: 'smooth' })
      }
    } else {
      window.location.href = hitUrl
    }
    setIsSearchVisible(false)
  }

  return (
    <div
      className={`search-hit ${isBlogs ? 'search-hit--blog' : ''} ${isSection ? 'search-hit--section' : 'search-hit--page'}`}
      tabIndex="0"
      onKeyDown={e => {
        if (e.key === 'Enter') {
          handleClick(e)
        }
      }}
    >
      <a
        href={hit.url}
        tabIndex="-1"
        className="search-hit__link"
        onClick={handleClick}
      >
        {isDocsFamily && (
          <>
            <span
              className="search-hit__glyph"
              aria-hidden="true"
            />
            <div className="search-hit__body">
              <div className="search-hit__breadcrumb">
                {formatSearchBreadcrumb(indexName, hit)}
              </div>
              <div className="search-hit__title">
                <Highlight attribute="title" hit={hit} />
              </div>
              {hit.description ? (
                <div className="search-hit__snippet">
                  <Highlight attribute="description" hit={hit} />
                </div>
              ) : null}
            </div>
          </>
        )}
        {isBlogs && hit.featuredImage?.url && (
          <img
            src={hit.featuredImage.url}
            alt={hit.featuredImage.alt || 'Blog image'}
            className="search-hit__thumb"
          />
        )}
        {isBlogs && (
          <div className="search-hit__main-row">
            <span className="search-hit__glyph" aria-hidden="true" />
            <div className="search-hit__body">
              <div className="search-hit__breadcrumb">{hit.tag || 'Blog'}</div>
              <div className="search-hit__title">
                <Highlight attribute="title" hit={hit} />
              </div>
              {hit.author?.name && hit.publishedDate && (
                <div className="search-hit__meta">
                  {hit.publishedDate} · {hit.author.name}
                </div>
              )}
              <div className="search-hit__snippet">
                <Highlight attribute="description" hit={hit} />
              </div>
            </div>
          </div>
        )}
        {!isDocsFamily && !isBlogs && (
          <>
            <span className="search-hit__glyph" aria-hidden="true" />
            <div className="search-hit__body">
              <div className="search-hit__title">
                <Highlight attribute="title" hit={hit} />
              </div>
              <div className="search-hit__snippet">
                <Highlight attribute="description" hit={hit} />
              </div>
            </div>
          </>
        )}
      </a>
    </div>
  )
}

const DocOrderedHits = connectHits(
  ({ hits, debounceQuery, setIsSearchVisible, indexName }) => {
  const sorted = sortHitsByDocumentStructure(hits, debounceQuery || '')
  return (
    <div className="ais-Hits">
      <ul className="ais-Hits-list">
        {sorted.map(hit => (
          <li key={hit.objectID} className="ais-Hits-item">
            <Hit
              hit={hit}
              setIsSearchVisible={setIsSearchVisible}
              indexName={indexName}
            />
          </li>
        ))}
      </ul>
    </div>
  )
}
)

const CustomStats = ({ nbHits, indexName, setDisplayHits }) => {
  useEffect(() => {
    setDisplayHits(nbHits > 0)
  }, [nbHits, setDisplayHits])

  const getIndexLabel = () => {
    switch (indexName) {
      case DOCS_INDEX_NAME:
        return 'Documentation'
      case 'blogs_test':
        return 'Blog'
      case 'website':
        return 'Website'
      case CLI_INDEX_NAME:
        return 'CLI'
      case SDK_INDEX_NAME:
        return 'SDK'
      case API_INDEX_NAME:
        return 'API'
      default:
        return 'Results'
    }
  }

  return (
    <div className="custom-stats">
      <span className="custom-stats__label">{getIndexLabel()}</span>
      <span className="custom-stats__count">
        {' '}
        ({nbHits})
      </span>
    </div>
  )
}

const Stats = connectStats(CustomStats)

const Search = ({ locale }) => {
  return (
    <GTProvider
      config={gtConfig}
      loadTranslations={loadTranslations}
      locale={locale}
      projectId={import.meta.env.PUBLIC_VITE_GT_PROJECT_ID}
      devApiKey={import.meta.env.PUBLIC_VITE_GT_API_KEY}
    >
      <SearchContent />
    </GTProvider>
  )
}

export default Search
