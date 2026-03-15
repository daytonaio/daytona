import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Map each agent's Render backend for the Vite dev proxy
// This completely bypasses CORS by routing through localhost
const agentProxies = {
  '/proxy/atlas':       { target: 'https://atlas-agent-ynj0.onrender.com',       changeOrigin: true, rewrite: (p) => p.replace(/^\/proxy\/atlas/, '') },
  '/proxy/sniper':      { target: 'https://sniper-agent-ynj0.onrender.com',      changeOrigin: true, rewrite: (p) => p.replace(/^\/proxy\/sniper/, '') },
  '/proxy/harvester':   { target: 'https://harvester-agent-ynj0.onrender.com',   changeOrigin: true, rewrite: (p) => p.replace(/^\/proxy\/harvester/, '') },
  '/proxy/airdrop':     { target: 'https://airdrop-agent-ynj0.onrender.com',     changeOrigin: true, rewrite: (p) => p.replace(/^\/proxy\/airdrop/, '') },
  '/proxy/chrono':      { target: 'https://chrono-agent-ynj0.onrender.com',      changeOrigin: true, rewrite: (p) => p.replace(/^\/proxy\/chrono/, '') },
  '/proxy/consigliere': { target: 'https://consigliere-agent-ynj0.onrender.com', changeOrigin: true, rewrite: (p) => p.replace(/^\/proxy\/consigliere/, '') },
  '/proxy/summary':     { target: 'https://summary-agent-ynj0.onrender.com',     changeOrigin: true, rewrite: (p) => p.replace(/^\/proxy\/summary/, '') },
  '/proxy/scribe':      { target: 'https://scribe-agent-ynj0.onrender.com',      changeOrigin: true, rewrite: (p) => p.replace(/^\/proxy\/scribe/, '') },
  '/proxy/trend':       { target: 'https://trend-agent-ynj0.onrender.com',       changeOrigin: true, rewrite: (p) => p.replace(/^\/proxy\/trend/, '') },
  '/proxy/whale':       { target: 'https://whale-agent-ynj0.onrender.com',       changeOrigin: true, rewrite: (p) => p.replace(/^\/proxy\/whale/, '') },
  '/proxy/guardian':    { target: 'https://guardian-agent-ynj0.onrender.com',     changeOrigin: true, rewrite: (p) => p.replace(/^\/proxy\/guardian/, '') },
  '/proxy/tax':         { target: 'https://tax-agent-ynj0.onrender.com',         changeOrigin: true, rewrite: (p) => p.replace(/^\/proxy\/tax/, '') },
  '/proxy/bid':         { target: 'https://bid-engine-ynj0.onrender.com',        changeOrigin: true, rewrite: (p) => p.replace(/^\/proxy\/bid/, '') },
}

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: agentProxies
  }
})
