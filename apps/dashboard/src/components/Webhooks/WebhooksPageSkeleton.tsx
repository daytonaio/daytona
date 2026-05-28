/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageContent, PageFooter, PageHeader, PageIntro, PageLayout } from '@/components/PageLayout'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { WebhooksEndpointTable } from '@/components/Webhooks/WebhooksEndpointTable'

export function WebhooksPageSkeleton() {
  return (
    <PageLayout contained>
      <PageHeader />

      <PageContent size="full" className="overflow-hidden">
        <PageIntro title="Webhooks" className="mb-8" actions={<Skeleton className="h-9 w-36" />} />
        <div className="min-h-0 flex-1 -mx-4 flex-col flex">
          <Tabs value="endpoints" className="flex min-h-0 flex-1 flex-col gap-0">
            <div className="flex items-center justify-between shadow-[inset_0_-1px] shadow-border">
              <TabsList variant="underline">
                <TabsTrigger value="endpoints">Endpoints</TabsTrigger>
                <TabsTrigger value="messages">Messages</TabsTrigger>
              </TabsList>
            </div>
            <TabsContent
              value="endpoints"
              className="min-h-0 p-4 data-[state=active]:flex data-[state=active]:flex-1 flex-col"
            >
              <WebhooksEndpointTable
                data={[]}
                loading={true}
                isLoadingEndpoint={() => false}
                onDisable={() => undefined}
                onDelete={() => undefined}
              />
            </TabsContent>
          </Tabs>
        </div>
      </PageContent>
      <PageFooter />
    </PageLayout>
  )
}
