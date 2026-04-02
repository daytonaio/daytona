/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { type CommandConfig, useRegisterCommands } from '@/components/CommandPalette'
import { PageContent, PageFooter, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { CreateEndpointSheet } from '@/components/Webhooks/CreateEndpointSheet'
import { WebhooksEndpointTable } from '@/components/Webhooks/WebhooksEndpointTable'
import { WebhooksMessagesTable } from '@/components/Webhooks/WebhooksMessagesTable/WebhooksMessagesTable'
import { Button } from '@/components/ui/button'
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useDeleteWebhookEndpointMutation } from '@/hooks/mutations/useDeleteWebhookEndpointMutation'
import { useUpdateWebhookEndpointMutation } from '@/hooks/mutations/useUpdateWebhookEndpointMutation'
import { handleApiError } from '@/lib/error-handling'
import { AlertCircle, PlusIcon, RefreshCw } from 'lucide-react'
import { useCallback, useMemo, useRef, useState } from 'react'
import { toast } from 'sonner'
import { EndpointOut } from 'svix'
import { useEndpoints } from 'svix-react'

const Webhooks: React.FC = () => {
  const endpoints = useEndpoints()
  const [mutatingEndpointId, setMutatingEndpointId] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState('endpoints')
  const createEndpointSheetRef = useRef<{ open: () => void }>(null)

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

  const rootCommands: CommandConfig[] = useMemo(() => {
    if (endpoints.error) {
      return []
    }

    return [
      {
        id: 'add-endpoint',
        label: 'Add Endpoint',
        icon: <PlusIcon className="w-4 h-4" />,
        onSelect: () => {
          setActiveTab('endpoints')
          createEndpointSheetRef.current?.open()
        },
      },
    ]
  }, [endpoints.error])

  useRegisterCommands(rootCommands, { groupId: 'webhook-actions', groupLabel: 'Webhook actions', groupOrder: 0 })

  if (endpoints.error) {
    return (
      <PageLayout>
        <PageHeader>
          <PageTitle>Webhooks</PageTitle>
        </PageHeader>
        <PageContent>
          <Empty className="py-12 max-h-64 border">
            <EmptyHeader>
              <EmptyMedia variant="icon" className="bg-destructive-background text-destructive">
                <AlertCircle />
              </EmptyMedia>
              <EmptyTitle className="text-destructive">Failed to load webhook endpoints</EmptyTitle>
              <EmptyDescription>Something went wrong while fetching your endpoints. Please try again.</EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
              <Button variant="secondary" size="sm" onClick={() => endpoints.reload()}>
                <RefreshCw />
                Retry
              </Button>
            </EmptyContent>
          </Empty>
        </PageContent>
      </PageLayout>
    )
  }

  return (
    <PageLayout contained>
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
        <CreateEndpointSheet
          onSuccess={handleSuccess}
          ref={createEndpointSheetRef}
          className={activeTab === 'endpoints' ? '' : 'hidden'}
        />
      </PageHeader>

      <PageContent size="full" className="pt-0 overflow-hidden">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="flex min-h-0 flex-1 flex-col gap-6">
          <TabsList
            className="shadow-none bg-transparent w-auto p-0 pb-0 justify-start rounded-none"
            variant="underline"
          >
            <TabsTrigger value="endpoints" className="-mb-0.5 pb-1.5">
              Endpoints
            </TabsTrigger>
            <TabsTrigger value="messages" className="-mb-0.5 pb-1.5">
              Messages
            </TabsTrigger>
          </TabsList>
          <TabsContent
            value="endpoints"
            className="min-h-0 data-[state=active]:flex data-[state=active]:flex-1 flex-col"
          >
            <WebhooksEndpointTable
              data={endpoints.data || []}
              loading={endpoints.loading}
              isLoadingEndpoint={isLoadingEndpoint}
              onDisable={handleDisable}
              onDelete={handleDelete}
            />
          </TabsContent>
          <TabsContent
            value="messages"
            className="min-h-0 data-[state=active]:flex data-[state=active]:flex-1 flex-col"
          >
            <WebhooksMessagesTable />
          </TabsContent>
        </Tabs>
      </PageContent>
      <PageFooter />
    </PageLayout>
  )
}

export default Webhooks
