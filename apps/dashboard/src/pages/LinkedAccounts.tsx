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
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { useLinkAccountMutation } from '@/hooks/mutations/useLinkAccountMutation'
import { useUnlinkAccountMutation } from '@/hooks/mutations/useUnlinkAccountMutation'
import { useAccountProvidersQuery } from '@/hooks/queries/useAccountProvidersQuery'
import { useConfig } from '@/hooks/useConfig'
import { handleApiError } from '@/lib/error-handling'
import { ShieldCheck } from 'lucide-react'
import { UserManager } from 'oidc-client-ts'
import React, { useMemo } from 'react'
import { useAuth } from 'react-oidc-context'
import { useLocation } from 'react-router-dom'
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
  const config = useConfig()
  const accountProvidersQuery = useAccountProvidersQuery()

  const linkingUserManager = useMemo(() => {
    return new UserManager({
      authority: config.oidc.issuer,
      client_id: config.oidc.clientId,
      extraQueryParams: {
        audience: config.oidc.audience,
      },
      scope: 'openid profile email offline_access',
      redirect_uri: window.location.origin,
      automaticSilentRenew: false,
    })
  }, [config.oidc.issuer, config.oidc.clientId, config.oidc.audience])

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
      .sort((a, b) => Number(b.isPrimary) - Number(a.isPrimary))
  }, [accountProvidersQuery.data, user])

  return (
    <Card>
      <CardHeader className="p-4">
        <CardTitle>Linked Accounts</CardTitle>
        <CardDescription>Link your accounts to your Daytona account for a seamless login.</CardDescription>
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
            <LinkedAccount key={provider.name} provider={provider} linkingUserManager={linkingUserManager} />
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

const LinkedAccount = ({
  provider,
  linkingUserManager,
}: {
  provider: AccountProvider
  linkingUserManager: UserManager
}) => {
  const linkAccountMutation = useLinkAccountMutation()
  const unlinkAccountMutation = useUnlinkAccountMutation()
  const { user, signinSilent } = useAuth()
  const location = useLocation()

  const locationPath = location.pathname + location.search

  const handleLinkAccount = async () => {
    if (!user?.profile.email_verified) {
      toast.error('Please verify your email before linking an account')
      return
    }

    try {
      const userToLink = await linkingUserManager.signinPopup({
        state: {
          returnTo: locationPath,
        },
        extraQueryParams: {
          connection: provider.name,
          accountLinking: true,
        },
      })

      const secondaryIdentity = (userToLink.profile.identities as UserProfileIdentity[])?.find(
        (i) => i.provider === provider.name,
      )

      if (!secondaryIdentity) {
        throw new Error('Failed to obtain account information')
      }

      await linkAccountMutation.mutateAsync({
        provider: secondaryIdentity.provider,
        userId: secondaryIdentity.userId,
      })

      toast.success(`Successfully linked account`)
      await new Promise((resolve) => setTimeout(resolve, 1500))
      const success = await signinSilent()
      if (!success) {
        window.location.reload()
      }
    } catch (error) {
      handleApiError(error, 'Failed to link account')
    }
  }

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
            <TooltipProvider>
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
            </TooltipProvider>
          )}
        </div>
        <p className="text-sm text-muted-foreground">
          {provider.isLinked
            ? provider.isPrimary
              ? `This is your primary account used for authentication.`
              : `Your ${provider.displayName} account is linked as a secondary login method.`
            : `Link your ${provider.displayName} account for a seamless login.`}
        </p>
      </div>
      {provider.isLinked ? (
        <Button
          variant="outline"
          onClick={handleUnlinkAccount}
          disabled={unlinkAccountMutation.isPending || provider.isPrimary}
        >
          {unlinkAccountMutation.isPending && <Spinner />}
          Unlink
        </Button>
      ) : (
        <Button onClick={handleLinkAccount} disabled={linkAccountMutation.isPending}>
          {linkAccountMutation.isPending && <Spinner />}
          Link Account
        </Button>
      )}
    </div>
  )
}
