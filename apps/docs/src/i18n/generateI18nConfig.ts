import { getLocaleProperties } from 'generaltranslation'

import config from '../../gt.config.json'

/**
 * Generates the i18n config for Astro.
 * @returns The i18n config for Astro.
 */
export function generateI18nConfig() {
  return {
    defaultLocale: config.defaultLocale,
    locales: Object.fromEntries(
      config.locales.map(locale => [
        locale,
        {
          lang: locale,
          label: getLocaleProperties(locale).nativeLanguageName,
        },
      ])
    ),
  }
}
