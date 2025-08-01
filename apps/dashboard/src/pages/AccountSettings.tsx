/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { useApi } from '@/hooks/useApi'
import LinkedAccounts from './LinkedAccounts'
import { handleApiError } from '@/lib/error-handling'
import { Button } from '@/components/ui/button'
import { Loader2 } from 'lucide-react'

const AccountSettings: React.FC<{ linkedAccountsEnabled: boolean }> = ({ linkedAccountsEnabled }) => {
  const { userApi } = useApi()
  const { user, signinSilent } = useAuth()

  const [enrollInSmsMfaLoading, setEnrollInSmsMfaLoading] = useState(false)

  const handleEnrollInSmsMfa = useCallback(async () => {
    try {
      setEnrollInSmsMfaLoading(true)
      const response = await userApi.enrollInSmsMfa()
      const popup = window.open(response.data, '_blank', 'width=500,height=700,scrollbars=yes,resizable=yes')
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
    } finally {
      setEnrollInSmsMfaLoading(false)
    }
  }, [userApi, signinSilent])

  return (
    <div className="p-6">
      <div className="mb-6 flex justify-between items-center">
        <h1 className="text-2xl font-medium">Account Settings</h1>
      </div>

      {linkedAccountsEnabled && <LinkedAccounts />}

      <div className="mt-8">
        <div className="mb-6 flex justify-between items-center">
          <h1 className="text-xl font-medium">Phone Verification</h1>
        </div>

        {user?.profile.phone_verified ? (
          <div className="p-4 border rounded-lg bg-green-50 border-green-200 max-w-xs">
            <p className="text-sm text-green-700">
              {user?.profile.phone_name ? <span>{user?.profile.phone_name as string}</span> : 'Phone number verified'}
            </p>
          </div>
        ) : (
          <div className="p-4 border rounded-lg max-w-xs">
            <div className="space-y-4">
              <p className="text-sm text-muted-foreground">Verify your phone number to increase account limits.</p>
              <Button onClick={handleEnrollInSmsMfa} disabled={enrollInSmsMfaLoading} className="w-full sm:w-auto">
                {enrollInSmsMfaLoading ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Setting up verification...
                  </>
                ) : (
                  'Verify Phone Number'
                )}
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default AccountSettings
