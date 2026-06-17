/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { type CommandConfig, useRegisterCommands } from '@/components/CommandPalette'
import { PageContent, PageFooter, PageHeader, PageIntro, PageLayout } from '@/components/PageLayout'
import { UpsertEndpointSheet } from '@/components/Webhooks/UpsertEndpointSheet'
import { WebhooksEndpointTable } from '@/components/Webhooks/WebhooksEndpointTable'
import { WebhooksMessagesTable } from '@/components/Webhooks/WebhooksMessagesTable/WebhooksMessagesTable'
import { Button } from '@/components/ui/button'
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useDeleteWebhookEndpointMutation } from '@/hooks/mutations/useDeleteWebhookEndpointMutation'
import { useUpdateWebhookEndpointMutation } from '@/hooks/mutations/useUpdateWebhookEndpointMutation'
import { handleApiError } from '@/lib/error-handling'
import { cn } from '@/lib/utils'
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
        label: 'Create Endpoint',
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
        <PageHeader />
        <PageContent>
          <PageIntro title="Webhooks" />
          <Empty className="py-12 max-h-64 border" variant="destructive">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <AlertCircle />
              </EmptyMedia>
              <EmptyTitle>Failed to load webhook endpoints</EmptyTitle>
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
      <PageHeader />

      <PageContent size="full" className="overflow-hidden">
        <PageIntro
          title="Webhooks"
          className="mb-8"
          actions={
            <UpsertEndpointSheet
              onSuccess={handleSuccess}
              ref={createEndpointSheetRef}
              className={cn({
                hidden: activeTab !== 'endpoints',
              })}
            />
          }
        />
        <div className="min-h-0 flex-1 -mx-4 flex flex-col">
          <Tabs value={activeTab} onValueChange={setActiveTab} className="flex min-h-0 flex-1 flex-col gap-0">
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
                data={endpoints.data || []}
                loading={endpoints.loading}
                isLoadingEndpoint={isLoadingEndpoint}
                onDisable={handleDisable}
                onDelete={handleDelete}
              />
            </TabsContent>
            <TabsContent
              value="messages"
              className="min-h-0 p-4 data-[state=active]:flex data-[state=active]:flex-1 flex-col"
            >
              <WebhooksMessagesTable />
            </TabsContent>
          </Tabs>
        </div>
      </PageContent>
      <PageFooter />
    </PageLayout>
  )
}

export default Webhooks
