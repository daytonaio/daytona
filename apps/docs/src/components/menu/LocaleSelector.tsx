import { GTProvider, useLocaleSelector } from 'gt-react'
import type { ChangeEvent, ComponentProps } from 'react'
import loadTranslations from 'src/i18n/loadTranslations'
import { localizePath } from 'src/i18n/utils'

import gtConfig from '../../../gt.config.json'
import styles from './LocaleSelector.module.scss'

function capitalizeLanguageName(language: string): string {
  if (!language) return ''
  return (
    language.charAt(0).toUpperCase() +
    (language.length > 1 ? language.slice(1) : '')
  )
}

type Props = ComponentProps<'select'>

function LocaleSelector({
  locales: _locales,
  ...props
}: Props & { locales?: string[] }): React.JSX.Element | null {
  const {
    locale: currentLocale,
    locales,
    getLocaleProperties,
  } = useLocaleSelector(_locales ? _locales : undefined)

  if (!locales || locales.length === 0 || !currentLocale) {
    return null
  }

  const getDisplayName = (locale: string) =>
    capitalizeLanguageName(getLocaleProperties(locale).nativeNameWithRegionCode)

  const handleChange = (event: ChangeEvent<HTMLSelectElement>) => {
    const nextLocale = event.target.value
    if (!nextLocale || nextLocale === currentLocale) return

    window.location.href = localizePath(
      window.location.pathname,
      nextLocale,
      currentLocale
    )
  }

  const { className, ...restProps } = props

  return (
    <select
      className={styles.localeSelector}
      value={currentLocale}
      onChange={handleChange}
    >
      {locales.map(locale => (
        <option key={locale} value={locale}>
          {getDisplayName(locale)}
        </option>
      ))}
    </select>
  )
}

export const LocaleSelect = ({
  locale,
  ...props
}: { locale: string } & Props) => {
  return (
    <GTProvider
      config={gtConfig}
      loadTranslations={loadTranslations}
      locale={locale}
      projectId={import.meta.env.PUBLIC_VITE_GT_PROJECT_ID}
      devApiKey={import.meta.env.PUBLIC_VITE_GT_API_KEY}
    >
      <LocaleSelector {...props} />
    </GTProvider>
  )
}
