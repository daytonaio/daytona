/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { PostHogProviderWrapper } from './components/PostHogProviderWrapper'
import App from './App'
import './index.css'
import { ErrorBoundary } from 'react-error-boundary'
import { ErrorBoundaryFallback } from './components/ErrorBoundaryFallback'
import { ConfigProvider } from './providers/ConfigProvider'

const root = ReactDOM.createRoot(document.getElementById('root') as HTMLElement)

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
)
