import { z } from 'astro:content'

import defaultI18n from '../content/i18n/en.json'

/**
 * Generates a schema for the i18n data.
 * @returns The schema for the i18n data.
 */
export function generateI18nSchema() {
  return z.object({
    ...Object.fromEntries(
      Object.keys(defaultI18n).map(key => [key, z.string().optional()])
    ),
  })
}
