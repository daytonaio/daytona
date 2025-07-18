import { useLocaleSelector } from 'gt-react'
import { useEffect, useRef, useState } from 'react'
import { localizePath } from 'src/i18n/utils'

/**
 * Capitalizes the first letter of a language name if applicable.
 * For languages that do not use capitalization, it returns the name unchanged.
 * @param {string} language - The name of the language.
 * @returns {string} The language name with the first letter capitalized if applicable.
 */
function capitalizeLanguageName(language: string): string {
  if (!language) return ''
  return (
    language.charAt(0).toUpperCase() +
    (language.length > 1 ? language.slice(1) : '')
  )
}

/**
 * A component that allows the user to select a locale.
 * @param locales - The locales to display.
 * @param props - The props to pass to the dropdown element.
 * @returns A custom dropdown with the locales.
 */
export default function LocaleSelector({
  locales: _locales,
  ...props
}: {
  locales?: string[]
  [key: string]: any
}): React.JSX.Element | null {
  const [isOpen, setIsOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  // Get locale selector properties
  const {
    locale: currentLocale,
    locales,
    getLocaleProperties,
  } = useLocaleSelector(_locales ? _locales : undefined)

  // Get display name
  const getDisplayName = (locale: string) => {
    return capitalizeLanguageName(
      getLocaleProperties(locale).nativeNameWithRegionCode
    )
  }

  // Set the locale and redirect to the new URL
  const setLocale = (locale: string) => {
    setIsOpen(false)
    window.location.href = localizePath(
      window.location.pathname,
      locale,
      currentLocale
    )
  }

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [])

  // If no locales are returned, just render nothing or handle gracefully
  if (!locales || locales.length === 0) {
    return null
  }

  const currentDisplayName = currentLocale ? getDisplayName(currentLocale) : ''

  return (
    <div {...props} className="locale-selector-wrapper" ref={dropdownRef}>
      <button
        className="locale-selector"
        onClick={() => setIsOpen(!isOpen)}
        type="button"
      >
        {currentDisplayName}
        <svg
          width="12"
          height="8"
          viewBox="0 0 12 8"
          fill="none"
          style={{
            marginLeft: '8px',
          }}
        >
          <path
            d="M1 1L6 6L11 1"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </button>

      {isOpen && (
        <div className="locale-dropdown">
          {locales.map(locale => (
            <button
              key={locale}
              className={`locale-option ${locale === currentLocale ? 'active' : ''}`}
              onClick={() => setLocale(locale)}
              type="button"
            >
              {getDisplayName(locale)}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
