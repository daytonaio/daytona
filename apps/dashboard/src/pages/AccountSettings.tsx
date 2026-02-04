/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DeleteAccountDialog } from '@/components/DeleteAccountDialog'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { Card, CardContent } from '@/components/ui/card'
import { useDeleteAccountMutation } from '@/hooks/mutations/useDeleteAccountMutation'
import React from 'react'
import { useAuth } from 'react-oidc-context'
import { toast } from 'sonner'
import LinkedAccounts from './LinkedAccounts'

const AccountSettings: React.FC<{ linkedAccountsEnabled: boolean }> = ({ linkedAccountsEnabled }) => {
  const { signoutRedirect } = useAuth()

  const deleteAccountMutation = useDeleteAccountMutation()

  const parseDeleteAccountError = (error: unknown): string[] => {
    const rawMessage = String((error as { message?: string } | undefined)?.message || error || '')
    return rawMessage
      .split(';')
      .map((reason) => reason.trim())
      .filter(Boolean)
  }

  const handleDeleteAccount = async (): Promise<{ success: boolean; reasons: string[] }> => {
    try {
      await deleteAccountMutation.mutateAsync()
      toast.success('Account deleted successfully')
      // Sign out and redirect after a short delay
      setTimeout(() => {
        signoutRedirect()
      }, 1500)
      return { success: true, reasons: [] }
    } catch (error) {
      return { success: false, reasons: parseDeleteAccountError(error) }
    }
  }

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Account Settings</PageTitle>
      </PageHeader>

      <PageContent>
        <div className="flex flex-col gap-6">
          {linkedAccountsEnabled && <LinkedAccounts />}

          <Card className="bg-destructive-background border-destructive-separator">
            <CardContent>
              <div className="flex sm:flex-row flex-col justify-between sm:items-center gap-2">
                <div className="text-sm">
                  <div className="text-muted-foreground">
                    <p className="font-semibold text-destructive-foreground">Danger Zone</p>
                    Delete your account and all associated data.
                  </div>
                </div>
                <DeleteAccountDialog onDeleteAccount={handleDeleteAccount} loading={deleteAccountMutation.isPending} />
              </div>
            </CardContent>
          </Card>
        </div>
      </PageContent>
    </PageLayout>
  )
}

export default AccountSettings
