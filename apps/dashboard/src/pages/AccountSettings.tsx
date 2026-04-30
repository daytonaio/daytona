/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageBreadcrumbs, PageContent, PageDocsLink, PageHeader, PageIntro, PageLayout } from '@/components/PageLayout'
import { PrivacyPreferencesDialog } from '@/components/PrivacyPreferencesDialog'
import { usePrivacyConsent } from '@/hooks/usePrivacyConsent'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'
import React, { useState } from 'react'
import LinkedAccounts from './LinkedAccounts'

const AccountSettings: React.FC<{ linkedAccountsEnabled: boolean }> = ({ linkedAccountsEnabled }) => {
  const { preferences, saveConsent } = usePrivacyConsent()
  const [showPrivacyDialog, setShowPrivacyDialog] = useState(false)

  return (
    <PageLayout>
      <PageHeader>
        <PageBreadcrumbs current="Account Settings" />
        <PageDocsLink href={`${DAYTONA_DOCS_URL}/en/linked-accounts/`} label="Account Docs" />
      </PageHeader>

      <PageContent>
        <PageIntro title="Account Settings" description="Manage account connections and privacy preferences." />
        <div className="flex flex-col gap-6">
          {linkedAccountsEnabled && <LinkedAccounts />}

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
