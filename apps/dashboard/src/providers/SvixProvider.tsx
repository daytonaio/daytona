/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { useWebhookAppPortalAccessQuery } from '@/hooks/queries/useWebhookAppPortalAccessQuery'
import { useWebhookInitializationStatusQuery } from '@/hooks/queries/useWebhookInitializationStatusQuery'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { RefreshCcw } from 'lucide-react'
import React from 'react'
import { SvixProvider as SvixReactProvider } from 'svix-react'

interface SvixProviderProps {
  children: React.ReactNode
}

export function SvixProvider({ children }: SvixProviderProps) {
  const { selectedOrganization } = useSelectedOrganization()
  const {
    data: appPortalAccess,
    isLoading: isLoadingAppPortalAccess,
    error: appPortalAccessError,
    refetch: refetchAppPortalAccess,
  } = useWebhookAppPortalAccessQuery(selectedOrganization?.id)
  const {
    data: initStatus,
    isLoading: isLoadingInitStatus,
    error: initStatusError,
    refetch: refetchInitStatus,
  } = useWebhookInitializationStatusQuery(selectedOrganization?.id)

  const isLoading = isLoadingAppPortalAccess || isLoadingInitStatus
  const error = appPortalAccessError || initStatusError
  const refetch = () => {
    refetchAppPortalAccess()
    refetchInitStatus()
  }

  if (isLoading) {
    return (
      <PageLayout>
        <PageHeader>
          <PageTitle>Webhooks</PageTitle>
        </PageHeader>
        <PageContent>
          <Card>
            <CardHeader>
              <Skeleton className="h-6 w-48" />
            </CardHeader>
            <CardContent className="space-y-4">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </CardContent>
          </Card>
        </PageContent>
      </PageLayout>
    )
  }

  if (error || !appPortalAccess?.token || !initStatus?.svixApplicationId) {
    return (
      <PageLayout>
        <PageHeader>
          <PageTitle>Webhooks</PageTitle>
        </PageHeader>
        <PageContent>
          <Card>
            <CardHeader>
              <CardTitle className="text-center">Oops, something went wrong</CardTitle>
            </CardHeader>
            <CardContent className="flex justify-between items-center flex-col gap-3">
              <div>Failed to load webhooks. Please try again later.</div>
              <Button variant="outline" onClick={() => refetch()}>
                <RefreshCcw className="mr-2 h-4 w-4" />
                Retry
              </Button>
            </CardContent>
          </Card>
        </PageContent>
      </PageLayout>
    )
  }

  return (
    <SvixReactProvider token={appPortalAccess.token} appId={initStatus.svixApplicationId}>
      {children}
    </SvixReactProvider>
  )
}
