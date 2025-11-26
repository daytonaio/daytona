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
import { useApi } from '@/hooks/useApi'
import { useConfig } from '@/hooks/useConfig'
import { handleApiError } from '@/lib/error-handling'
import { AccountProvider as AccountProviderApi } from '@daytonaio/api-client'
import { ShieldCheck } from 'lucide-react'
import { UserManager } from 'oidc-client-ts'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
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
  const { userApi } = useApi()
  const { user, signinSilent } = useAuth()
  const location = useLocation()
  const config = useConfig()

  const [loadingProviders, setLoadingProviders] = useState(true)
  const [providers, setProviders] = useState<AccountProviderApi[]>([])

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

  const fetchAvailableProviders = useCallback(async () => {
    try {
      const response = await userApi.getAvailableAccountProviders()
      setProviders(response.data)
    } catch (error) {
      handleApiError(error, 'Failed to fetch available account providers')
    } finally {
      setLoadingProviders(false)
    }
  }, [userApi])

  useEffect(() => {
    fetchAvailableProviders()
  }, [fetchAvailableProviders])

  const accountProviders = useMemo<AccountProvider[]>(() => {
    const identities = (user?.profile.identities as UserProfileIdentity[]) || []

    return providers
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
  }, [providers, user])

  const [processingProviderActions, setProcessingProviderActions] = useState<Record<string, boolean>>({})

  const handleLinkAccount = async (provider: AccountProvider) => {
    if (!user?.profile.email_verified) {
      toast.error('Please verify your email before linking an account')
      return
    }

    setProcessingProviderActions((prev) => ({ ...prev, [provider.name]: true }))
    try {
      const userToLink = await linkingUserManager.signinPopup({
        state: {
          returnTo: location.pathname + location.search,
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

      await userApi.linkAccount({
        provider: secondaryIdentity.provider,
        userId: secondaryIdentity.userId,
      })

      toast.success(`Successfully linked account`)
      // signinSilent triggers top level loader, so we need to wait a short interval to make sure the toast is visible to the user
      await new Promise((resolve) => setTimeout(resolve, 1500))
      const success = await signinSilent()
      if (!success) {
        window.location.reload()
      }
    } catch (error) {
      handleApiError(error, 'Failed to link account')
    } finally {
      setProcessingProviderActions((prev) => ({ ...prev, [provider.name]: false }))
    }
  }

  const handleUnlinkAccount = async (provider: AccountProvider) => {
    if (provider.isPrimary) {
      toast.error('Primary account cannot be unlinked')
      return
    }

    if (!provider.userId) {
      return
    }

    setProcessingProviderActions((prev) => ({ ...prev, [provider.name]: true }))
    try {
      await userApi.unlinkAccount(provider.name, provider.userId)
      toast.success('Successfully unlinked account')
      // signinSilent triggers top level loader, so we need to wait a short interval to make sure the toast is visible to the user
      await new Promise((resolve) => setTimeout(resolve, 1500))
      const success = await signinSilent()
      if (!success) {
        window.location.reload()
      }
    } catch (error) {
      handleApiError(error, 'Failed to unlink account')
    } finally {
      setProcessingProviderActions((prev) => ({ ...prev, [provider.name]: false }))
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Linked Accounts</CardTitle>
        <CardDescription>Link your accounts to your Daytona account for a seamless login.</CardDescription>
      </CardHeader>
      {loadingProviders ? (
        <CardContent className="flex flex-col gap-5">
          {[...Array(2)].map((_, index) => (
            <ProviderSkeleton key={index} />
          ))}
        </CardContent>
      ) : (
        <CardContent className="p-0 mt-5">
          {accountProviders.map((provider) => (
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
                  onClick={() => handleUnlinkAccount(provider)}
                  disabled={processingProviderActions[provider.name] || provider.isPrimary}
                >
                  {processingProviderActions[provider.name] && <Spinner />}
                  Unlink
                </Button>
              ) : (
                <Button onClick={() => handleLinkAccount(provider)} disabled={processingProviderActions[provider.name]}>
                  {processingProviderActions[provider.name] && <Spinner />}
                  Link Account
                </Button>
              )}
            </div>
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
