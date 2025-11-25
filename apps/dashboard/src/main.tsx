/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import ReactDOM from 'react-dom/client'
import { ErrorBoundary } from 'react-error-boundary'
import { BrowserRouter } from 'react-router-dom'
import App from './App'
import { ErrorBoundaryFallback } from './components/ErrorBoundaryFallback'
import { PostHogProviderWrapper } from './components/PostHogProviderWrapper'
import './index.css'
import { ConfigProvider } from './providers/ConfigProvider'

const root = ReactDOM.createRoot(document.getElementById('root') as HTMLElement)

async function enableMocking() {
  if (import.meta.env.VITE_ENABLE_MOCKING !== 'true') {
    return
  }

  const { worker } = await import('./mocks/browser')
  return worker.start()
}

enableMocking().then(() =>
  root.render(
    <React.StrictMode>
      <ErrorBoundary FallbackComponent={ErrorBoundaryFallback}>
        <ConfigProvider>
          <PostHogProviderWrapper>
            <BrowserRouter>
              <App />
            </BrowserRouter>
          </PostHogProviderWrapper>
        </ConfigProvider>
      </ErrorBoundary>
    </React.StrictMode>,
  ),
)
