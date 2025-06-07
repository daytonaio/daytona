/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { PostHogProviderWrapper } from './components/PostHogProviderWrapper'
import { oidcConfig } from './auth/oidc-config'
import { AuthProvider } from 'react-oidc-context'
import App from './App'
import './index.css'
import { ErrorBoundary } from 'react-error-boundary'
import { ErrorBoundaryFallback } from './components/ErrorBoundaryFallback'

const root = ReactDOM.createRoot(document.getElementById('root') as HTMLElement)

root.render(
  <React.StrictMode>
    <ErrorBoundary FallbackComponent={ErrorBoundaryFallback}>
      <PostHogProviderWrapper>
        <AuthProvider {...oidcConfig}>
          <BrowserRouter>
            <App />
          </BrowserRouter>
        </AuthProvider>
      </PostHogProviderWrapper>
    </ErrorBoundary>
  </React.StrictMode>,
)
