const { createGlobPatternsForDependencies } = require('@nx/react/tailwind')
const { join } = require('path')

/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: ['class'],
  content: [
    join(__dirname, '{src,pages,components,app}/**/*!(*.stories|*.spec).{ts,tsx,html}'),
    ...createGlobPatternsForDependencies(__dirname),
  ],
  theme: {
    extend: {
      colors: {
        background: '#0a0e1a',
        foreground: '#e2e8f0',
        primary: {
          DEFAULT: '#3b82f6',
          foreground: '#ffffff',
        },
        secondary: {
          DEFAULT: '#8b5cf6',
          foreground: '#ffffff',
        },
        accent: {
          DEFAULT: '#10b981',
          foreground: '#ffffff',
        },
        muted: {
          DEFAULT: '#1e293b',
          foreground: '#94a3b8',
        },
        border: '#334155',
        card: {
          DEFAULT: '#1e293b',
          foreground: '#e2e8f0',
        },
      },
      backgroundImage: {
        'space-gradient': 'linear-gradient(135deg, #0a0e1a 0%, #1e293b 50%, #0f172a 100%)',
      },
    },
  },
  plugins: [],
}
