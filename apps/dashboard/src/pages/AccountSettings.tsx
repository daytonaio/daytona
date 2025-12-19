/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Spinner } from '@/components/ui/spinner'
import { useEnrollInSmsMfaMutation } from '@/hooks/mutations/useEnrollInSmsMfaMutation'
import { handleApiError } from '@/lib/error-handling'
import { CheckCircleIcon } from 'lucide-react'
import React, { useCallback } from 'react'
import { useAuth } from 'react-oidc-context'
import LinkedAccounts from './LinkedAccounts'

const AccountSettings: React.FC<{ linkedAccountsEnabled: boolean }> = ({ linkedAccountsEnabled }) => {
  const { user, signinSilent } = useAuth()
  const enrollInSmsMfaMutation = useEnrollInSmsMfaMutation()

  const handleEnrollInSmsMfa = useCallback(async () => {
    try {
      const url = await enrollInSmsMfaMutation.mutateAsync()
      const popup = window.open(url, '_blank', 'width=500,height=700,scrollbars=yes,resizable=yes')
      if (popup) {
        const checkClosed = setInterval(() => {
          if (popup.closed) {
            clearInterval(checkClosed)
            signinSilent().catch((error) => {
              handleApiError(error, 'Failed to refresh user data. Please sign out and sign in again.')
            })
          }
        }, 100)
      }
    } catch (error) {
      handleApiError(error, 'Failed to enroll in SMS MFA')
    }
  }, [enrollInSmsMfaMutation, signinSilent])

  const isPhoneVerified = user?.profile.phone_verified

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Account Settings</PageTitle>
      </PageHeader>

      <PageContent>
        <div className="flex flex-col gap-6">
          {linkedAccountsEnabled && <LinkedAccounts />}
          <Card>
            <CardHeader className="p-4">
              <CardTitle>Verification</CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              <div className="p-4 border-t border-border flex items-center justify-between gap-2">
                <div className="flex flex-col gap-1">
                  <div>Phone Verification</div>
                  <div>
                    {isPhoneVerified ? (
                      <div className="text-sm text-muted-foreground flex items-center gap-2">
                        <CheckCircleIcon className="w-4 h-4 shrink-0" /> Phone number verified
                      </div>
                    ) : (
                      <div className="text-sm text-muted-foreground">
                        Verify your phone number to increase account limits.
                      </div>
                    )}
                  </div>
                </div>
                {!isPhoneVerified && (
                  <Button onClick={handleEnrollInSmsMfa} disabled={enrollInSmsMfaMutation.isPending}>
                    {enrollInSmsMfaMutation.isPending && <Spinner />} Verify
                  </Button>
                )}
              </div>
            </CardContent>
          </Card>
        </div>
      </PageContent>
    </PageLayout>
  )
}

export default AccountSettings
