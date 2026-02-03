/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { CreateEndpointDialog } from '@/components/Webhooks/CreateEndpointDialog'
import { WebhooksEndpointTable } from '@/components/Webhooks/WebhooksEndpointTable'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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
  const [endpointToDelete, setEndpointToDelete] = useState<EndpointOut | null>(null)
  const [deleteDialogIsOpen, setDeleteDialogIsOpen] = useState(false)

  const updateMutation = useUpdateWebhookEndpointMutation()
  const deleteMutation = useDeleteWebhookEndpointMutation()

  const handleDisable = useCallback(
    async (endpoint: EndpointOut) => {
      setMutatingEndpointId(endpoint.id)
      try {
        await updateMutation.mutateAsync({
          endpointId: endpoint.id,
          update: { disabled: !endpoint.disabled, url: endpoint.url },
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
        setEndpointToDelete(null)
        setDeleteDialogIsOpen(false)
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
    toast.success('Endpoint created')
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
        <CreateEndpointDialog onSuccess={handleSuccess} className="ml-auto" />
      </PageHeader>

      <PageContent>
        <WebhooksEndpointTable
          data={endpoints.data || []}
          loading={endpoints.loading}
          isLoadingEndpoint={isLoadingEndpoint}
          onDisable={handleDisable}
          onDelete={(endpoint) => {
            setEndpointToDelete(endpoint)
            setDeleteDialogIsOpen(true)
          }}
        />
      </PageContent>

      <AlertDialog
        open={deleteDialogIsOpen}
        onOpenChange={(isOpen) => {
          setDeleteDialogIsOpen(isOpen)
          if (!isOpen) {
            setEndpointToDelete(null)
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Webhook Endpoint</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete this endpoint? This action cannot be undone. All webhook history for this
              endpoint will be permanently deleted.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => endpointToDelete && handleDelete(endpointToDelete)}
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </PageLayout>
  )
}

export default Webhooks
