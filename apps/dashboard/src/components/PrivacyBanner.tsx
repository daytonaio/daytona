/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Cookie, Settings } from 'lucide-react'
import { useState } from 'react'
import { PrivacyPreferencesDialog, usePrivacyConsent } from './PrivacyPreferencesDialog'
import { Button } from './ui/button'

export function PrivacyBanner() {
  const { hasConsented, preferences, saveConsent } = usePrivacyConsent()
  const [showCustomize, setShowCustomize] = useState(false)

  const handleAcceptAll = () => {
    saveConsent({
      necessary: true,
      analytics: true,
      preferences: true,
      marketing: true,
    })
  }

  const handleRejectAll = () => {
    saveConsent({
      necessary: true,
      analytics: false,
      preferences: false,
      marketing: false,
    })
  }

  if (hasConsented) return null

  return (
    <>
      <div className="fixed bottom-0 left-0 right-0 z-50 p-4 md:p-6 pointer-events-none">
        <div className="mx-auto max-w-4xl rounded-lg border border-border bg-background p-5 shadow-xl pointer-events-auto">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div className="flex items-start gap-3">
              <Cookie className="mt-1 h-6 w-6 flex-shrink-0 text-primary" />
              <div className="space-y-1">
                <h4 className="font-semibold text-sm">We value your privacy</h4>
                <p className="text-sm text-muted-foreground max-w-xl">
                  We use tracking technologies for essential functionality like authentication, and optionally for
                  analytics to improve our product.
                </p>
              </div>
            </div>
            <div className="flex flex-col sm:flex-row gap-2 flex-shrink-0 pt-2 lg:pt-0">
              <Button variant="outline" size="sm" onClick={() => setShowCustomize(true)}>
                <Settings className="mr-2 h-4 w-4" />
                Customize
              </Button>
              <Button variant="secondary" size="sm" onClick={handleRejectAll}>
                Essential Only
              </Button>
              <Button size="sm" onClick={handleAcceptAll}>
                Accept All
              </Button>
            </div>
          </div>
        </div>
      </div>

      <PrivacyPreferencesDialog
        open={showCustomize}
        onOpenChange={setShowCustomize}
        preferences={preferences}
        onSave={saveConsent}
      />
    </>
  )
}
