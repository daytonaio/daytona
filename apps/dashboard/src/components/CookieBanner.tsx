/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Cookie, Settings } from 'lucide-react'
import { usePostHog } from 'posthog-js/react'
import { useEffect, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { Button } from './ui/button'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from './ui/dialog'
import { Switch } from './ui/switch'

type ConsentPreferences = {
  necessary: boolean
  analytics: boolean
  marketing: boolean
}

const CONSENT_KEY_PREFIX = 'cookie_consent_'

const DEFAULT_PREFERENCES: ConsentPreferences = {
  necessary: true,
  analytics: false,
  marketing: false,
}

export function CookieBanner() {
  const posthog = usePostHog()
  const { user } = useAuth()

  const [showBanner, setShowBanner] = useState(false)
  const [showCustomize, setShowCustomize] = useState(false)
  const [tempPreferences, setTempPreferences] = useState<ConsentPreferences>(DEFAULT_PREFERENCES)

  const consentKey = user?.profile?.sub ? `${CONSENT_KEY_PREFIX}${user.profile.sub}` : null

  useEffect(() => {
    if (!consentKey || !posthog) return

    try {
      const saved = localStorage.getItem(consentKey)
      if (saved) {
        const parsedConsent: ConsentPreferences = JSON.parse(saved)
        setTempPreferences(parsedConsent)
        if (parsedConsent.analytics) {
          posthog.opt_in_capturing()
        } else {
          posthog.opt_out_capturing()
        }
        setShowBanner(false)
      } else {
        setShowBanner(true)
      }
    } catch {
      console.error('Error parsing cookie consent')
      setShowBanner(true)
    }
  }, [posthog, consentKey])

  const saveConsent = (preferences: ConsentPreferences) => {
    if (!consentKey) return

    localStorage.setItem(consentKey, JSON.stringify(preferences))

    if (preferences.analytics) {
      posthog?.opt_in_capturing()
    } else {
      posthog?.opt_out_capturing()
    }

    setShowBanner(false)
    setShowCustomize(false)
  }

  const handleAcceptAll = () => {
    saveConsent({
      necessary: true,
      analytics: true,
      marketing: true,
    })
  }

  const handleRejectAll = () => {
    saveConsent({
      necessary: true,
      analytics: false,
      marketing: false,
    })
  }

  const handleOpenCustomize = () => {
    setShowCustomize(true)
  }

  if (!showBanner) return null

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
                  We use cookies for authentication and analytics to improve our product.
                </p>
              </div>
            </div>
            <div className="flex flex-col sm:flex-row gap-2 flex-shrink-0 pt-2 lg:pt-0">
              <Button variant="outline" size="sm" onClick={handleOpenCustomize}>
                <Settings className="mr-2 h-4 w-4" />
                Customize
              </Button>
              <Button variant="secondary" size="sm" onClick={handleRejectAll}>
                Reject All
              </Button>
              <Button size="sm" onClick={handleAcceptAll}>
                Accept All
              </Button>
            </div>
          </div>
        </div>
      </div>

      <Dialog open={showCustomize} onOpenChange={setShowCustomize}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Cookie Preferences</DialogTitle>
          </DialogHeader>

          <div className="space-y-4">
            <div className="flex items-start justify-between space-x-2 rounded-md border p-3 bg-muted/50">
              <div className="space-y-1">
                <p className="font-medium text-sm">Strictly Necessary</p>
                <p className="text-xs text-muted-foreground">Required for authentication.</p>
              </div>
              <Switch checked disabled />
            </div>

            <div className="flex items-start justify-between space-x-2 rounded-md border p-3">
              <div className="space-y-1">
                <p className="font-medium text-sm">Analytics</p>
                <p className="text-xs text-muted-foreground">Helps us improve the product.</p>
              </div>
              <Switch
                checked={tempPreferences.analytics}
                onCheckedChange={(checked) => setTempPreferences((prev) => ({ ...prev, analytics: checked }))}
              />
            </div>

            <div className="flex items-start justify-between space-x-2 rounded-md border p-3">
              <div className="space-y-1">
                <p className="font-medium text-sm">Preferences & Marketing</p>
                <p className="text-xs text-muted-foreground">Personalization and targeted content.</p>
              </div>
              <Switch
                checked={tempPreferences.marketing}
                onCheckedChange={(checked) => setTempPreferences((prev) => ({ ...prev, marketing: checked }))}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCustomize(false)}>
              Cancel
            </Button>
            <Button onClick={() => saveConsent(tempPreferences)}>Save Preferences</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
