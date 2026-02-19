/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AccountProviderIcon } from '@/components/AccountProviderIcon'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { useUnlinkAccountMutation } from '@/hooks/mutations/useUnlinkAccountMutation'
import { useAccountProvidersQuery } from '@/hooks/queries/useAccountProvidersQuery'
import { handleApiError } from '@/lib/error-handling'
import { ShieldCheck } from 'lucide-react'
import React, { useMemo } from 'react'
import { useAuth } from 'react-oidc-context'
import { toast } from 'sonner'

export interface UserProfileIdentity {
  provider: string
  userId: string
}

interface AccountProvider {
  name: string
  displayName: string
  userId: string | null
  isLinked: boolean
  isPrimary: boolean
}

const LinkedAccounts: React.FC = () => {
  const { user } = useAuth()
  const accountProvidersQuery = useAccountProvidersQuery()

  const accountProviders = useMemo<AccountProvider[]>(() => {
    const identities = (user?.profile.identities as UserProfileIdentity[]) || []

    return (accountProvidersQuery.data || [])
      .map((provider) => {
        const identity = identities.find((i) => i.provider === provider.name)
        return {
          name: provider.name,
          displayName: provider.displayName,
          userId: identity?.userId || null,
          isLinked: Boolean(identity),
          isPrimary: Boolean(identity && user?.profile.sub === `${identity.provider}|${identity.userId}`),
        }
      })
      .filter((provider) => provider.isLinked)
      .sort((a, b) => Number(b.isPrimary) - Number(a.isPrimary))
  }, [accountProvidersQuery.data, user])

  return (
    <Card>
      <CardHeader className="p-4">
        <CardTitle>Linked Accounts</CardTitle>
        <CardDescription>View and manage accounts linked to your Daytona account.</CardDescription>
      </CardHeader>
      {accountProvidersQuery.isLoading ? (
        <CardContent className="flex flex-col gap-5">
          {[...Array(2)].map((_, index) => (
            <ProviderSkeleton key={index} />
          ))}
        </CardContent>
      ) : (
        <CardContent className="p-0">
          {accountProviders.map((provider) => (
            <LinkedAccount key={provider.name} provider={provider} />
          ))}
        </CardContent>
      )}
    </Card>
  )
}

const ProviderSkeleton = () => {
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center gap-2">
        <Skeleton className="w-5 h-5" />
        <Skeleton className="w-24 h-5" />
      </div>
      <Skeleton className="w-full h-5" />
    </div>
  )
}

export default LinkedAccounts

const LinkedAccount = ({ provider }: { provider: AccountProvider }) => {
  const unlinkAccountMutation = useUnlinkAccountMutation()
  const { signinSilent } = useAuth()

  const handleUnlinkAccount = async () => {
    if (provider.isPrimary) {
      toast.error('Primary account cannot be unlinked')
      return
    }

    if (!provider.userId) {
      return
    }

    try {
      await unlinkAccountMutation.mutateAsync({ provider: provider.name, userId: provider.userId })
      toast.success('Successfully unlinked account')
      await new Promise((resolve) => setTimeout(resolve, 1500))
      const success = await signinSilent()
      if (!success) {
        window.location.reload()
      }
    } catch (error) {
      handleApiError(error, 'Failed to unlink account')
    }
  }

  return (
    <div className="flex items-center gap-3 p-4 border-t border-border justify-between">
      <div className="flex flex-col">
        <div className="flex items-center gap-2">
          <AccountProviderIcon provider={provider.name} className="h-4 w-4" />
          <div>{provider.displayName}</div>
          {provider.isPrimary && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Badge variant="outline" className="gap-1 text-xs">
                  <ShieldCheck className="h-3 w-3" />
                  Primary
                </Badge>
              </TooltipTrigger>
              <TooltipContent>
                <p>Primary accounts cannot be unlinked</p>
              </TooltipContent>
            </Tooltip>
          )}
        </div>
        <p className="text-sm text-muted-foreground">
          {provider.isPrimary
            ? `This is your primary account used for authentication.`
            : `Your ${provider.displayName} account is linked as a secondary login method.`}
        </p>
      </div>
      {!provider.isPrimary && (
        <Button variant="outline" onClick={handleUnlinkAccount} disabled={unlinkAccountMutation.isPending}>
          {unlinkAccountMutation.isPending && <Spinner />}
          Unlink
        </Button>
      )}
    </div>
  )
}
