/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageContent, PageHeader, PageIntro, PageLayout } from '@/components/PageLayout'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { EndpointEventsTable } from '@/components/Webhooks/EndpointEventsTable'
import { ArrowLeft } from 'lucide-react'

export function WebhookEndpointDetailsSkeleton({ onBack }: { onBack?: () => void }) {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-3 min-w-0">
        <Button variant="ghost" size="icon-sm" className="shrink-0" onClick={onBack}>
          <ArrowLeft className="w-4 h-4" />
        </Button>
        <Skeleton className="h-6 w-48" />
        <Skeleton className="h-5 w-16" />
        <Skeleton className="h-4 w-24" />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Endpoint Configuration</CardTitle>
        </CardHeader>
        <CardContent className="p-4 flex flex-col gap-4">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <div className="flex flex-col">
              <div className="text-muted-foreground text-xs mb-1">URL</div>
              <Skeleton className="h-9 w-full" />
            </div>
            <div className="flex flex-col">
              <div className="text-muted-foreground text-xs mb-1">Signing Secret</div>
              <Skeleton className="h-9 w-full" />
            </div>
          </div>
          <div>
            <div className="text-muted-foreground text-xs mb-1">Listening For</div>
            <div className="flex flex-wrap gap-1.5">
              <Skeleton className="h-5 w-24 rounded-full" />
              <Skeleton className="h-5 w-32 rounded-full" />
              <Skeleton className="h-5 w-20 rounded-full" />
            </div>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Delivery Stats</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-2">
            <Skeleton className="h-2 w-full rounded-full" />
            <div className="flex items-center gap-4">
              <Skeleton className="h-3 w-20" />
              <Skeleton className="h-3 w-16" />
              <Skeleton className="h-3 w-18" />
              <Skeleton className="h-3 w-18" />
            </div>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Event History</CardTitle>
        </CardHeader>
        <CardContent>
          <EndpointEventsTable data={[]} loading={true} onReplay={() => undefined} />
        </CardContent>
      </Card>
    </div>
  )
}

export function WebhookEndpointDetailsPageSkeleton() {
  return (
    <PageLayout>
      <PageHeader />

      <PageContent className="gap-6">
        <PageIntro title="Webhooks" />
        <WebhookEndpointDetailsSkeleton />
      </PageContent>
    </PageLayout>
  )
}
