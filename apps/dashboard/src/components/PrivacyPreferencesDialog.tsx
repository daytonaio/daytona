/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { usePostHog } from 'posthog-js/react'
import { useCallback, useEffect, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { Button } from './ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from './ui/dialog'
import { Separator } from './ui/separator'
import { Switch } from './ui/switch'

type ConsentPreferences = {
  necessary: boolean
  analytics: boolean
  preferences: boolean
  marketing: boolean
}

const CONSENT_KEY_PREFIX = 'privacy_consent_'

const DEFAULT_PREFERENCES: ConsentPreferences = {
  necessary: true,
  analytics: false,
  preferences: false,
  marketing: false,
}

export function usePrivacyConsent() {
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
      console.error('Error parsing privacy consent')
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

type PrivacyPreferencesDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  preferences: ConsentPreferences
  onSave: (preferences: ConsentPreferences) => void
}

export function PrivacyPreferencesDialog({ open, onOpenChange, preferences, onSave }: PrivacyPreferencesDialogProps) {
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
          <DialogTitle>Privacy Preferences</DialogTitle>
        </DialogHeader>

        <DialogDescription>
          Choose which tracking technologies you allow. Essential tracking is always required.
        </DialogDescription>

        <div className="rounded-md border">
          <div className="flex items-start justify-between space-x-2 p-3 bg-muted/50">
            <div className="space-y-1">
              <p className="font-medium text-sm">Essential</p>
              <p className="text-xs text-muted-foreground">
                Required for login sessions and core functionality. Cannot be disabled.
              </p>
            </div>
            <Switch checked disabled />
          </div>

          <Separator />

          <div className="flex items-start justify-between space-x-2 p-3">
            <div className="space-y-1">
              <p className="font-medium text-sm">Analytics</p>
              <p className="text-xs text-muted-foreground">
                Collects anonymous usage data to help us understand how the product is used and improve it.
              </p>
            </div>
            <Switch
              checked={tempPreferences.analytics}
              onCheckedChange={(checked) => setTempPreferences((prev) => ({ ...prev, analytics: checked }))}
            />
          </div>

          <Separator />

          <div className="flex items-start justify-between space-x-2 p-3">
            <div className="space-y-1">
              <p className="font-medium text-sm">Preferences</p>
              <p className="text-xs text-muted-foreground">
                Remembers your settings like theme and layout across sessions.
              </p>
            </div>
            <Switch
              checked={tempPreferences.preferences}
              onCheckedChange={(checked) => setTempPreferences((prev) => ({ ...prev, preferences: checked }))}
            />
          </div>

          <Separator />

          <div className="flex items-start justify-between space-x-2 p-3">
            <div className="space-y-1">
              <p className="font-medium text-sm">Marketing</p>
              <p className="text-xs text-muted-foreground">
                Used for communications about Daytona features and updates.
              </p>
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
