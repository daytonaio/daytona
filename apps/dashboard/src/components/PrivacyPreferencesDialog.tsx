/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ConsentPreferences } from '@/hooks/usePrivacyConsent'
import { useEffect, useState } from 'react'
import { Button } from './ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from './ui/dialog'
import { ScrollArea } from './ui/scroll-area'
import { Separator } from './ui/separator'
import { Switch } from './ui/switch'

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

        <ScrollArea fade="mask" className="h-[325px] overflow-auto -mx-5 min-h-0">
          <div className="rounded-md border mx-5">
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
        </ScrollArea>

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
