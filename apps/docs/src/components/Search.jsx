import { liteClient as algoliasearch } from 'algoliasearch/lite'
import { GTProvider, useGT } from 'gt-react'
import { useCallback, useEffect, useRef, useState } from 'react'
import {
  Configure,
  Highlight,
  Hits,
  Index,
  InstantSearch,
  Pagination,
  SearchBox,
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
        hitsPerPage={10}
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
      <Hits
        hitComponent={props => (
          <Hit
            {...props}
            setIsSearchVisible={setIsSearchVisible}
            indexName={indexName}
          />
        )}
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
