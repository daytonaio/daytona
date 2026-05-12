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
import { DAYTONA_DOCS_URL } from '@/constants/ExternalLinks'
import { useDeleteWebhookEndpointMutation } from '@/hooks/mutations/useDeleteWebhookEndpointMutation'
import { useUpdateWebhookEndpointMutation } from '@/hooks/mutations/useUpdateWebhookEndpointMutation'
import { handleApiError } from '@/lib/error-handling'
import { AlertCircle, BookOpen, PlusIcon, RefreshCw } from 'lucide-react'
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
        <PageHeader />
        <PageContent>
          <PageIntro
            title="Webhooks"
            actions={
              <Button
                variant="link"
                size="sm"
                className="w-8 gap-0 px-0 text-muted-foreground hover:text-foreground xs:w-auto xs:gap-1.5 xs:px-3"
                asChild
              >
                <a href={`${DAYTONA_DOCS_URL}/en/webhooks/`} target="_blank" rel="noopener noreferrer">
                  <BookOpen className="size-4" />
                  <span className="sr-only xs:not-sr-only">Docs</span>
                </a>
              </Button>
            }
          />
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
          actions={
            <>
              <Button
                variant="link"
                size="sm"
                className="w-8 gap-0 px-0 text-muted-foreground hover:text-foreground xs:w-auto xs:gap-1.5 xs:px-3"
                asChild
              >
                <a href={`${DAYTONA_DOCS_URL}/en/webhooks/`} target="_blank" rel="noopener noreferrer">
                  <BookOpen className="size-4" />
                  <span className="sr-only xs:not-sr-only">Docs</span>
                </a>
              </Button>
              <UpsertEndpointSheet
                onSuccess={handleSuccess}
                ref={createEndpointSheetRef}
                className={activeTab === 'endpoints' ? '' : 'hidden'}
              />
            </>
          }
        />
        <Tabs value={activeTab} onValueChange={setActiveTab} className="flex min-h-0 flex-1 flex-col gap-0">
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
      </PageContent>
      <PageFooter />
    </PageLayout>
  )
}

export default Webhooks
