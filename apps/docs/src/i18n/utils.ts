import config from '../../gt.config.json'

const defaultLocale = config.defaultLocale

/**
 * Localizes a path to a given locale.
 * @param path - The path to localize.
 * @param locale - The locale to localize to.
 * @param currentLocale - The current locale.
 * @returns The localized path.
 */
export function localizePath(
  path: string,
  locale: string = defaultLocale,
  currentLocale?: string
): string {
  if (!currentLocale || !path.startsWith(`/docs/${currentLocale}`)) {
    return path.replace(`/docs`, `/docs/${locale}`)
  }
  return path.replace(`/docs/${currentLocale}`, `/docs/${locale}`)
}
