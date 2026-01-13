/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import React from 'react'
import LinkedAccounts from './LinkedAccounts'

const AccountSettings: React.FC<{ linkedAccountsEnabled: boolean }> = ({ linkedAccountsEnabled }) => {
  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Account Settings</PageTitle>
      </PageHeader>

      <PageContent>
        <div className="flex flex-col gap-6">{linkedAccountsEnabled && <LinkedAccounts />}</div>
      </PageContent>
    </PageLayout>
  )
}

export default AccountSettings
