/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Loader2, ShieldCheck } from 'lucide-react'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { useApi } from '@/hooks/useApi'
import { handleApiError } from '@/lib/error-handling'
import { useLocation } from 'react-router-dom'
import { AccountProviderIcon } from '@/components/AccountProviderIcon'
import { AccountProvider as AccountProviderApi } from '@daytonaio/api-client'

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
  const { user, signinPopup } = useAuth()
  const location = useLocation()

  const [loadingProviders, setLoadingProviders] = useState(true)
  const [providers, setProviders] = useState<AccountProviderApi[]>([])

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
      await signinPopup({
        state: {
          returnTo: location.pathname + location.search,
        },
        extraQueryParams: {
          connection: provider.name,
        },
      })
      toast.success(`Successfully linked account`)
      window.location.reload()
    } catch (error) {
      handleApiError(error, 'Failed to link account')
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
      window.location.reload()
    } catch (error) {
      handleApiError(error, 'Failed to unlink account')
      setProcessingProviderActions((prev) => ({ ...prev, [provider.name]: false }))
    }
  }

  return (
    <div className="p-6">
      <div className="mb-6 flex justify-between items-center">
        <h1 className="text-2xl font-bold">Linked Accounts</h1>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {accountProviders.map((provider) => (
          <Card key={provider.name}>
            <CardHeader>
              <div className="flex items-center gap-3">
                <AccountProviderIcon provider={provider.name} className="h-5 w-5" />
                <div className="flex flex-col">
                  <div className="flex items-center gap-3">
                    <CardTitle className="text-lg">{provider.displayName}</CardTitle>
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
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground">
                {provider.isLinked
                  ? provider.isPrimary
                    ? `This is your primary account used for authentication.`
                    : `Your ${provider.displayName} account is linked as a secondary login method.`
                  : `Link your ${provider.displayName} account with the same email for a seamless login.`}
              </p>
            </CardContent>
            <CardFooter>
              {provider.isLinked ? (
                <Button
                  variant="outline"
                  className="w-full"
                  onClick={() => handleUnlinkAccount(provider)}
                  disabled={processingProviderActions[provider.name] || provider.isPrimary}
                >
                  {processingProviderActions[provider.name] ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                  Unlink
                </Button>
              ) : (
                <Button
                  className="w-full"
                  onClick={() => handleLinkAccount(provider)}
                  disabled={processingProviderActions[provider.name]}
                >
                  {processingProviderActions[provider.name] ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                  Link Account
                </Button>
              )}
            </CardFooter>
          </Card>
        ))}
      </div>
    </div>
  )
}

export default LinkedAccounts
