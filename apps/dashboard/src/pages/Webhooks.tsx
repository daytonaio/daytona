/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { CreateEndpointDialog } from '@/components/Webhooks/CreateEndpointDialog'
import { WebhooksEndpointTable } from '@/components/Webhooks/WebhooksEndpointTable'
import { WebhooksMessagesTable } from '@/components/Webhooks/WebhooksMessagesTable/WebhooksMessagesTable'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useDeleteWebhookEndpointMutation } from '@/hooks/mutations/useDeleteWebhookEndpointMutation'
import { useUpdateWebhookEndpointMutation } from '@/hooks/mutations/useUpdateWebhookEndpointMutation'
import { handleApiError } from '@/lib/error-handling'
import { RefreshCcw } from 'lucide-react'
import React, { useCallback, useState } from 'react'
import { toast } from 'sonner'
import { EndpointOut } from 'svix'
import { useEndpoints } from 'svix-react'

const Webhooks: React.FC = () => {
  const endpoints = useEndpoints()
  const [mutatingEndpointId, setMutatingEndpointId] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState('endpoints')

  const updateMutation = useUpdateWebhookEndpointMutation()
  const deleteMutation = useDeleteWebhookEndpointMutation()

  const handleDisable = useCallback(
    async (endpoint: EndpointOut) => {
      setMutatingEndpointId(endpoint.id)
      try {
        await updateMutation.mutateAsync({
          endpointId: endpoint.id,
          update: { disabled: !endpoint.disabled },
        })
        toast.success('Endpoint updated')
        endpoints.reload()
      } catch (error) {
        handleApiError(error, 'Failed to update endpoint')
      } finally {
        setMutatingEndpointId(null)
      }
    },
    [updateMutation, endpoints],
  )

  const handleDelete = useCallback(
    async (endpoint: EndpointOut) => {
      setMutatingEndpointId(endpoint.id)
      try {
        await deleteMutation.mutateAsync({ endpointId: endpoint.id })
        toast.success('Endpoint deleted')
        endpoints.reload()
      } catch (error) {
        handleApiError(error, 'Failed to delete endpoint')
      } finally {
        setMutatingEndpointId(null)
      }
    },
    [deleteMutation, endpoints],
  )

  const handleSuccess = useCallback(() => {
    endpoints.reload()
  }, [endpoints])

  const isLoadingEndpoint = useCallback(
    (endpoint: EndpointOut) => {
      return mutatingEndpointId === endpoint.id && (updateMutation.isPending || deleteMutation.isPending)
    },
    [mutatingEndpointId, updateMutation.isPending, deleteMutation.isPending],
  )

  if (endpoints.error) {
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
              <div>There was an error loading your webhook endpoints.</div>
              <Button variant="outline" onClick={() => endpoints.reload()}>
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
    <PageLayout>
      <PageHeader>
        <PageTitle>Webhooks</PageTitle>
        <a
          href="https://www.daytona.io/docs/en/tools/api/#daytona/webhook/undefined/"
          target="_blank"
          rel="noopener noreferrer"
          className="ml-auto"
        >
          <Button variant="link" size="sm">
            Docs
          </Button>
        </a>
        {activeTab === 'endpoints' && <CreateEndpointDialog onSuccess={handleSuccess} />}
      </PageHeader>

      <PageContent>
        <Tabs value={activeTab} onValueChange={setActiveTab} className="gap-6">
          <TabsList className="shadow-none bg-transparent w-auto p-0 pb-0 justify-start border-b rounded-none">
            <TabsTrigger
              value="endpoints"
              className="data-[state=inactive]:border-b-transparent data-[state=active]:border-b-foreground border-b rounded-none !shadow-none -mb-0.5 pb-1.5"
            >
              Endpoints
            </TabsTrigger>
            <TabsTrigger
              value="messages"
              className="data-[state=inactive]:border-b-transparent data-[state=active]:border-b-foreground border-b rounded-none !shadow-none -mb-0.5 pb-1.5"
            >
              Messages
            </TabsTrigger>
          </TabsList>
          <TabsContent value="endpoints">
            <WebhooksEndpointTable
              data={endpoints.data || []}
              loading={endpoints.loading}
              isLoadingEndpoint={isLoadingEndpoint}
              onDisable={handleDisable}
              onDelete={handleDelete}
            />
          </TabsContent>
          <TabsContent value="messages">
            <WebhooksMessagesTable />
          </TabsContent>
        </Tabs>
      </PageContent>
    </PageLayout>
  )
}

export default Webhooks
