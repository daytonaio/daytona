/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { Suspense } from 'react'
import ReactDOM from 'react-dom/client'
import { ErrorBoundary } from 'react-error-boundary'
import { NuqsAdapter } from 'nuqs/adapters/react-router/v6'
import { BrowserRouter } from 'react-router-dom'
import App from './App'
import { ErrorBoundaryFallback } from './components/ErrorBoundaryFallback'
import LoadingFallback from './components/LoadingFallback'
import { PostHogProviderWrapper } from './components/PostHogProviderWrapper'
import { ThemeProvider } from './contexts/ThemeContext'
import './index.css'
import { ConfigProvider } from './providers/ConfigProvider'
import { QueryProvider } from './providers/QueryProvider'

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
        <QueryProvider>
          <ThemeProvider>
            <Suspense fallback={<LoadingFallback />}>
              <ConfigProvider>
                <PostHogProviderWrapper>
                  <NuqsAdapter>
                    <BrowserRouter>
                      <App />
                    </BrowserRouter>
                  </NuqsAdapter>
                </PostHogProviderWrapper>
              </ConfigProvider>
            </Suspense>
          </ThemeProvider>
        </QueryProvider>
      </ErrorBoundary>
    </React.StrictMode>,
  ),
)
