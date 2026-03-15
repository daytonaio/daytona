/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#00D4FF',
        background: '#0A0A0F',
        card: '#12121A',
        border: '#1E1E2E',
        success: '#00FF94',
        warning: '#FFB800',
        danger: '#FF4444',
        muted: '#8888AA',
        langchain: '#3B82F6',
        langgraph: '#8B5CF6',
        crewai: '#10B981',
        autogen: '#F59E0B',
      },
      fontFamily: {
        inter: ['Inter', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
