import { liteClient as algoliasearch } from 'algoliasearch/lite'
import { GTProvider, useGT } from 'gt-react'
import { useEffect, useRef, useState } from 'react'
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
  import.meta.env.PUBLIC_ALGOLIA_DOCS_INDEX_NAME || 'docs_test'
const CLI_INDEX_NAME = import.meta.env.PUBLIC_ALGOLIA_CLI_INDEX_NAME || 'cli_test'
const SDK_INDEX_NAME = import.meta.env.PUBLIC_ALGOLIA_SDK_INDEX_NAME || 'sdk_test'

const searchClient =
  ALGOLIA_APP_ID && ALGOLIA_API_KEY
    ? algoliasearch(ALGOLIA_APP_ID, ALGOLIA_API_KEY)
    : null

function SearchContent() {
  const [isSearchVisible, setIsSearchVisible] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [debounceQuery, setDebounceQuery] = useState('')
  const [displayHits, setDisplayHits] = useState(false)
  const [totalHits, setTotalHits] = useState(0)
  const debounceTimeoutRef = useRef(null)
  const searchWrapperRef = useRef(null)
  const t = useGT()

  useEffect(() => {
    const toggleSearch = () => {
      setIsSearchVisible(prev => {
        if (prev) {
          setSearchQuery('')
          setDebounceQuery('')
          setDisplayHits(false)
          setTotalHits(0)
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
    }, 400)

    return () => {
      if (debounceTimeoutRef.current) {
        clearTimeout(debounceTimeoutRef.current)
      }
    }
  }, [searchQuery])

  return (
    isSearchVisible && (
      <div
        id="searchbox-wrapper"
        className="searchbox-wrapper"
        ref={searchWrapperRef}
      >
        <InstantSearch indexName={DOCS_INDEX_NAME} searchClient={searchClient}>
          <div className="search-bar-container">
            <SearchBox
              translations={{
                placeholder: t('Search daytona.io', {
                  $context: 'As in a search bar on a website',
                }),
              }}
              autoFocus
              onChange={event => setSearchQuery(event.currentTarget.value)}
              value={searchQuery}
            />
          </div>
          <div className="search-content">
            {debounceQuery && (
              <>
                <SearchIndex
                  indexName={DOCS_INDEX_NAME}
                  setDisplayHits={setDisplayHits}
                  setIsSearchVisible={setIsSearchVisible}
                  setTotalHits={setTotalHits}
                  debounceQuery={debounceQuery}
                />
                <SearchIndex
                  indexName={CLI_INDEX_NAME}
                  setDisplayHits={setDisplayHits}
                  setIsSearchVisible={setIsSearchVisible}
                  setTotalHits={setTotalHits}
                  debounceQuery={debounceQuery}
                />
                <SearchIndex
                  indexName={SDK_INDEX_NAME}
                  setDisplayHits={setDisplayHits}
                  setIsSearchVisible={setIsSearchVisible}
                  setTotalHits={setTotalHits}
                  debounceQuery={debounceQuery}
                />
                {totalHits === 0 && (
                  <div style={{ 
                    textAlign: 'center', 
                    padding: '20px',
                    color: 'var(--primary-text-color)',
                    fontSize: '16px'
                  }}>
                    No results found for "<strong>{debounceQuery}</strong>"
                  </div>
                )}
              </>
            )}
            <Configure hitsPerPage={10} clickAnalytics getRankingInfo={false} />
          </div>
        </InstantSearch>
      </div>
    )
  )
}

function SearchIndex({ indexName, setDisplayHits, setIsSearchVisible, setTotalHits, debounceQuery }) {
  return (
    <Index indexName={indexName}>
      <ConditionalSearchIndex 
        indexName={indexName}
        setDisplayHits={setDisplayHits}
        setIsSearchVisible={setIsSearchVisible}
        setTotalHits={setTotalHits}
        debounceQuery={debounceQuery}
      />
    </Index>
  )
}

const ConditionalSearchIndexComponent = ({ indexName, setDisplayHits, setIsSearchVisible, nbHits, setTotalHits, debounceQuery }) => {
  useEffect(() => {
    setDisplayHits(nbHits > 0)
    setTotalHits(prev => prev + nbHits)
  }, [nbHits, setDisplayHits, setTotalHits, debounceQuery])

  if (nbHits === 0) {
    return null
  }

  return (
    <>
      <div data-index={indexName}>
        <div
          className="stats-pagination-wrapper"
          style={indexName === 'blogs_test' ? { marginTop: '24px' } : {}}
        >
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
      <hr style={{ marginBottom: '40px' }} />
    </>
  )
}

const ConditionalSearchIndex = connectStats(ConditionalSearchIndexComponent)

function Hit({ hit, setIsSearchVisible, indexName }) {
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
      tabIndex="0"
      onKeyDown={e => {
        if (e.key === 'Enter') {
          handleClick(e)
        }
      }}
    >
      <a href={hit.url} tabIndex="-1" onClick={handleClick}>
        {([DOCS_INDEX_NAME, CLI_INDEX_NAME, SDK_INDEX_NAME].includes(indexName) || indexName === 'website') && (
          <>
            <h5
              style={{
                fontSize: '20px',
                display: 'flex',
                alignItems: 'center',
              }}
            >
              <span style={{ fontSize: '10px', marginRight: '8px' }}>🟦</span>
              <span style={{ marginLeft: '4px' }}>
                <Highlight attribute="title" hit={hit} />
              </span>
            </h5>
            <h6
              style={{
                fontSize: '12px',
                color: '#686868',
                fontWeight: 500,
                paddingLeft: '24px',
              }}
            >
              {hit.slug}
            </h6>
          </>
        )}
        {indexName === 'blogs_test' && hit.featuredImage?.url && (
          <img
            src={hit.featuredImage.url}
            alt={hit.featuredImage.alt || 'Blog image'}
            style={{
              width: '100%',
              maxWidth: '500px',
              marginBottom: '12px',
              border: '1px solid var(--border-color)',
            }}
          />
        )}
        {indexName === 'blogs_test' && (
          <h5
            style={{ fontSize: '20px', display: 'flex', alignItems: 'center' }}
          >
            <span style={{ fontSize: '10px', marginRight: '8px' }}>🟦</span>
            <span style={{ marginLeft: '4px' }}>
              <Highlight attribute="title" hit={hit} />
            </span>
          </h5>
        )}
        {indexName === 'blogs_test' &&
          hit.author?.name &&
          hit.publishedDate && (
            <p
              style={{
                fontSize: '14px',
                paddingLeft: '24px',
                paddingBottom: '8px',
              }}
            >
              {hit.publishedDate} :: {hit.author.name}
            </p>
          )}
        <p
          style={{
            fontSize: '12px',
            paddingBottom: '16px',
            paddingLeft: '24px',
          }}
        >
          <Highlight attribute="description" hit={hit} />
        </p>
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
      default:
        return 'Results'
    }
  }

  return (
    <div className="custom-stats">
      <span style={{ color: 'var(--primary-text-color)' }}>
        {getIndexLabel()}{' '}
      </span>
      ({nbHits} results)
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
