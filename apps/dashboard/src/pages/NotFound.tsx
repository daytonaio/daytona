/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { RoutePath } from '@/enums/RoutePath'
import { Home } from 'lucide-react'

const NotFound: React.FC = () => {
  const navigate = useNavigate()

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="text-center space-y-6 max-w-lg">
        <h1 className="text-4xl font-bold text-foreground animate-bounce">404</h1>
        <p className="text-base text-muted-foreground">The page you're looking for doesn't exist or has been moved.</p>
        <Button onClick={() => navigate(RoutePath.DASHBOARD)} className="flex items-center gap-2 mx-auto">
          <Home className="w-4 h-4" />
          Go to Dashboard
        </Button>
      </div>
    </div>
  )
}

export default NotFound
