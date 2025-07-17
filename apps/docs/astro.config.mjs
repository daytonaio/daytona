import node from '@astrojs/node'
import react from '@astrojs/react'
import starlight from '@astrojs/starlight'
import { ExpressiveCodeTheme } from '@astrojs/starlight/expressive-code'
import { defineConfig } from 'astro/config'
import fs from 'node:fs'
import { loadEnv } from 'vite'

const { PUBLIC_WEB_URL } = loadEnv(import.meta.env.MODE, process.cwd(), '')

const jsonDarkString = fs.readFileSync(
  new URL(`src/assets/themes/daytona-code-dark.json`, import.meta.url),
  'utf-8'
)
const jsonLightString = fs.readFileSync(
  new URL(`src/assets/themes/daytona-code-light.json`, import.meta.url),
  'utf-8'
)
const myThemeDark = ExpressiveCodeTheme.fromJSONString(jsonDarkString)
const myThemeLight = ExpressiveCodeTheme.fromJSONString(jsonLightString)

// https://astro.build/config
export default defineConfig({
  site: PUBLIC_WEB_URL,
  base: '/docs',
  integrations: [
    react(),
    starlight({
      favicon: '/favicon.ico',
      title: 'Daytona',
      social: {
        github: 'https://github.com/daytonaio',
      },
      editLink: {
        baseUrl: 'https://github.com/daytonaio/docs/blob/main/',
      },
      tableOfContents: { minHeadingLevel: 2, maxHeadingLevel: 4 },
      customCss: ['./src/fonts/font-face.css', './src/styles/style.scss'],
      components: {
        Footer: './src/components/Footer.astro',
        MarkdownContent: './src/components/MarkdownContent.astro',
        Pagination: './src/components/Pagination.astro',
        Header: './src/components/Header.astro',
        PageSidebar: './src/components/PageSidebar.astro',
        PageFrame: './src/components/PageFrame.astro',
        Sidebar: './src/components/Sidebar.astro',
        TwoColumnContent: './src/components/TwoColumnContent.astro',
        TableOfContents: './src/components/TableOfContents.astro',
        MobileMenuToggle: './src/components/MobileMenuToggle.astro',
        ContentPanel: './src/components/ContentPanel.astro',
        PageTitle: './src/components/PageTitle.astro',
        Hero: './src/components/Hero.astro',
        ThemeProvider: './src/components/ThemeProvider.astro',
        ThemeSelect: './src/components/ThemeSelect.astro',
        Head: './src/components/Head.astro',
        EditLink: './src/components/EditLink.astro',
        ExploreMore: './src/components/ExploreMore.astro',
      },
      expressiveCode: {
        minSyntaxHighlightingColorContrast: 3.0,
        themes: [myThemeDark, myThemeLight],
      },
    }),
  ],
  output: 'hybrid',
  adapter: node({
    mode: 'middleware',
  }),
  outDir: '../../dist/apps/docs',
  vite: {
    ssr: {
      noExternal: ['path-to-regexp'],
    },
  },
})
