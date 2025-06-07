/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'

const LoadingFallback = () => (
  <div className="fixed top-0 left-0 w-full h-full p-6 bg-background z-[3]">
    <div className="flex items-center gap-2">
      <div className="w-4 h-4 border-2 border-foreground border-t-muted rounded-full animate-spin" />
      <p className="text-foreground text-sm">Loading...</p>
    </div>
  </div>
)

export default LoadingFallback
