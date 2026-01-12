/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useState, useEffect } from 'react'
import { usePostHog } from 'posthog-js/react'
import { useAuth } from 'react-oidc-context'
import { Button } from './ui/button'
import { Cookie } from 'lucide-react'

const CONSENT_KEY_PREFIX = 'cookie_consent_'

export function CookieBanner() {
  const posthog = usePostHog()
  const { user } = useAuth()
  const [showBanner, setShowBanner] = useState(false)

  const consentKey = user?.profile?.sub ? `${CONSENT_KEY_PREFIX}${user.profile.sub}` : null

  useEffect(() => {
    if (posthog && consentKey) {
      const savedConsent = localStorage.getItem(consentKey)
      if (savedConsent === 'accepted') {
        posthog.opt_in_capturing()
        setShowBanner(false)
      } else if (savedConsent === 'rejected') {
        posthog.opt_out_capturing()
        setShowBanner(false)
      } else {
        setShowBanner(true)
      }
    }
  }, [posthog, consentKey])

  if (!showBanner) {
    return null
  }

  const handleAccept = () => {
    posthog?.opt_in_capturing()
    if (consentKey) {
      localStorage.setItem(consentKey, 'accepted')
    }
    setShowBanner(false)
  }

  const handleReject = () => {
    posthog?.opt_out_capturing()
    if (consentKey) {
      localStorage.setItem(consentKey, 'rejected')
    }
    setShowBanner(false)
  }

  return (
    <div className="fixed bottom-0 left-0 right-0 z-50 p-4 md:p-6 pointer-events-none">
      <div className="mx-auto max-w-2xl rounded-lg border border-border bg-background p-4 shadow-lg pointer-events-auto">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-start gap-3">
            <Cookie className="mt-0.5 h-5 w-5 flex-shrink-0 text-muted-foreground" />
            <p className="text-sm text-muted-foreground">
              We use cookies to analyze site usage and improve your experience.
            </p>
          </div>
          <div className="flex gap-2 flex-shrink-0">
            <Button variant="outline" size="sm" onClick={handleReject}>
              Reject
            </Button>
            <Button size="sm" onClick={handleAccept}>
              Accept
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
