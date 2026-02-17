/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { usePostHog } from 'posthog-js/react'
import { useCallback, useEffect, useState } from 'react'
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

export function useCookieConsent() {
  const posthog = usePostHog()
  const { user } = useAuth()

  const [hasConsented, setHasConsented] = useState(false)
  const [preferences, setPreferences] = useState<ConsentPreferences>(DEFAULT_PREFERENCES)

  const consentKey = user?.profile?.sub ? `${CONSENT_KEY_PREFIX}${user.profile.sub}` : null

  useEffect(() => {
    if (!consentKey || !posthog) return

    try {
      const saved = localStorage.getItem(consentKey)
      if (saved) {
        const parsedConsent: ConsentPreferences = JSON.parse(saved)
        setPreferences(parsedConsent)
        if (parsedConsent.analytics) {
          posthog.opt_in_capturing()
        } else {
          posthog.opt_out_capturing()
        }
        setHasConsented(true)
      } else {
        setHasConsented(false)
      }
    } catch {
      console.error('Error parsing cookie consent')
      setHasConsented(false)
    }
  }, [posthog, consentKey])

  const saveConsent = useCallback(
    (newPreferences: ConsentPreferences) => {
      if (!consentKey) return

      localStorage.setItem(consentKey, JSON.stringify(newPreferences))
      setPreferences(newPreferences)

      if (newPreferences.analytics) {
        posthog?.opt_in_capturing()
      } else {
        posthog?.opt_out_capturing()
      }

      setHasConsented(true)
    },
    [consentKey, posthog],
  )

  return { hasConsented, preferences, saveConsent }
}

type CookiePreferencesDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  preferences: ConsentPreferences
  onSave: (preferences: ConsentPreferences) => void
}

export function CookiePreferencesDialog({ open, onOpenChange, preferences, onSave }: CookiePreferencesDialogProps) {
  const [tempPreferences, setTempPreferences] = useState<ConsentPreferences>(preferences)

  useEffect(() => {
    if (open) {
      setTempPreferences(preferences)
    }
  }, [open, preferences])

  const handleSave = () => {
    onSave(tempPreferences)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
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
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSave}>Save Preferences</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
