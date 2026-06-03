/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageContent, PageHeader, PageIntro, PageLayout } from '@/components/PageLayout'
import { PrivacyPreferencesDialog } from '@/components/PrivacyPreferencesDialog'
import { usePrivacyConsent } from '@/hooks/usePrivacyConsent'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { useConfig } from '@/hooks/useConfig'
import React, { useState } from 'react'
import LinkedAccounts from './LinkedAccounts'

const AccountSettings: React.FC = () => {
  const config = useConfig()
  const { preferences, saveConsent } = usePrivacyConsent()
  const [showPrivacyDialog, setShowPrivacyDialog] = useState(false)

  return (
    <PageLayout>
      <PageHeader />

      <PageContent>
        <PageIntro title="Account Settings" />
        <div className="flex flex-col gap-6">
          {config.linkedAccountsEnabled && <LinkedAccounts />}

          <Card>
            <CardContent>
              <div className="flex sm:flex-row flex-col justify-between sm:items-center gap-2">
                <div className="text-sm">
                  <div className="text-muted-foreground">
                    <p className="font-semibold text-foreground">Privacy Preferences</p>
                    Manage which tracking technologies are used for analytics and marketing.
                  </div>
                </div>
                <Button variant="outline" onClick={() => setShowPrivacyDialog(true)}>
                  Manage Preferences
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>

        <PrivacyPreferencesDialog
          open={showPrivacyDialog}
          onOpenChange={setShowPrivacyDialog}
          preferences={preferences}
          onSave={saveConsent}
        />
      </PageContent>
    </PageLayout>
  )
}

export default AccountSettings
