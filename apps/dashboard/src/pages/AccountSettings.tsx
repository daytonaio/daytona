/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CookiePreferencesDialog, useCookieConsent } from '@/components/CookiePreferencesDialog'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import React, { useState } from 'react'
import LinkedAccounts from './LinkedAccounts'

const AccountSettings: React.FC<{ linkedAccountsEnabled: boolean }> = ({ linkedAccountsEnabled }) => {
  const { preferences, saveConsent } = useCookieConsent()
  const [showCookieDialog, setShowCookieDialog] = useState(false)

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Account Settings</PageTitle>
      </PageHeader>

      <PageContent>
        <div className="flex flex-col gap-6">
          {linkedAccountsEnabled && <LinkedAccounts />}

          <Card>
            <CardContent>
              <div className="flex sm:flex-row flex-col justify-between sm:items-center gap-2">
                <div className="text-sm">
                  <div className="text-muted-foreground">
                    <p className="font-semibold text-foreground">Cookie Preferences</p>
                    Manage which cookies are used for analytics and marketing.
                  </div>
                </div>
                <Button variant="outline" onClick={() => setShowCookieDialog(true)}>
                  Cookie Settings
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>

        <CookiePreferencesDialog
          open={showCookieDialog}
          onOpenChange={setShowCookieDialog}
          preferences={preferences}
          onSave={saveConsent}
        />
      </PageContent>
    </PageLayout>
  )
}

export default AccountSettings
