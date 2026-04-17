import { docsLoader, i18nLoader } from '@astrojs/starlight/loaders'
import { docsSchema, i18nSchema } from '@astrojs/starlight/schema'
import { defineCollection, z } from 'astro:content'
import { generateI18nSchema } from 'src/i18n/generateI18nSchema'

export const collections = {
  docs: defineCollection({
    loader: docsLoader(),
    schema: docsSchema({
      extend: z.object({
        licence: z.string().optional(),
        distribution: z.string().optional(),
        hideTitleOnPage: z.boolean().optional(),
        llmUrl: z.string().optional(),
        toolboxLlmUrl: z.string().optional(),
      }),
    }),
  }),
  i18n: defineCollection({
    loader: i18nLoader(),
    schema: i18nSchema({
      extend: generateI18nSchema(),
    }),
  }),
}
