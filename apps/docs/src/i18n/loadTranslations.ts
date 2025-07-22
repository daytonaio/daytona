/**
 * Loads the translations for a given locale.
 * @param locale - The locale to load translations for.
 * @returns The translations for the given locale.
 */
export default async function loadTranslations(locale: string) {
  const t = await import(`../data/i18n/${locale}.json`)
  return t.default
}
