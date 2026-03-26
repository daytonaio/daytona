/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CopyButton } from '@/components/CopyButton'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { TimestampTooltip } from '@/components/TimestampTooltip'
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
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { InputGroup, InputGroupButton, InputGroupInput } from '@/components/ui/input-group'
import { Skeleton } from '@/components/ui/skeleton'
import DeliveryStatsLine from '@/components/Webhooks/DeliveryStatsLine'
import { EditEndpointDialog } from '@/components/Webhooks/EditEndpointDialog'
import { EndpointEventsTable } from '@/components/Webhooks/EndpointEventsTable'
import { RoutePath } from '@/enums/RoutePath'
import { useDeleteWebhookEndpointMutation } from '@/hooks/mutations/useDeleteWebhookEndpointMutation'
import { useReplayWebhookEventMutation } from '@/hooks/mutations/useReplayWebhookEventMutation'
import { useRotateWebhookSecretMutation } from '@/hooks/mutations/useRotateWebhookSecretMutation'
import { useUpdateWebhookEndpointMutation } from '@/hooks/mutations/useUpdateWebhookEndpointMutation'
import { handleApiError } from '@/lib/error-handling'
import { getMaskedToken, getRelativeTimeString } from '@/lib/utils'
import { ArrowLeft, Eye, EyeOff, Loader2, MoreHorizontal, RefreshCcw } from 'lucide-react'
import React, { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { toast } from 'sonner'
import { useAttemptedMessages, useEndpoint, useEndpointSecret, useEndpointStats } from 'svix-react'

const WebhookEndpointDetails: React.FC = () => {
  const { endpointId } = useParams<{ endpointId: string }>()
  const navigate = useNavigate()
  const [isSecretRevealed, setIsSecretRevealed] = useState(false)
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [disableDialogOpen, setDisableDialogOpen] = useState(false)
  const [rotateSecretDialogOpen, setRotateSecretDialogOpen] = useState(false)

  const endpoint = useEndpoint(endpointId || '')
  const secret = useEndpointSecret(endpointId || '')
  const messages = useAttemptedMessages(endpointId || '', {})
  const stats = useEndpointStats(endpointId || '')

  const updateMutation = useUpdateWebhookEndpointMutation()
  const deleteMutation = useDeleteWebhookEndpointMutation()
  const rotateSecretMutation = useRotateWebhookSecretMutation()
  const replayMutation = useReplayWebhookEventMutation()

  const isMutating = updateMutation.isPending || deleteMutation.isPending || rotateSecretMutation.isPending

  const isRefreshing = endpoint.loading || secret.loading || messages.loading || stats.loading

  const statsIsLoading = !stats.data && stats.loading
  const statsIsFetching = !!stats.data && stats.loading

  const handleRetry = () => {
    endpoint.reload()
    secret.reload()
    messages.reload()
    stats.reload()
  }

  const handleDisable = async () => {
    if (!endpoint.data) return
    setDisableDialogOpen(false)
    try {
      await updateMutation.mutateAsync({
        endpointId: endpoint.data.id,
        update: { disabled: !endpoint.data.disabled },
      })
      toast.success('Endpoint updated')
      endpoint.reload()
    } catch (error) {
      handleApiError(error, 'Failed to update endpoint')
    }
  }

  const handleDelete = async () => {
    if (!endpoint.data) return
    try {
      await deleteMutation.mutateAsync({ endpointId: endpoint.data.id })
      toast.success('Endpoint deleted')
      setDeleteDialogOpen(false)
      navigate(RoutePath.WEBHOOKS)
    } catch (error) {
      handleApiError(error, 'Failed to delete endpoint')
    }
  }

  const handleRotateSecret = async () => {
    if (!endpoint.data) return
    try {
      await rotateSecretMutation.mutateAsync({ endpointId: endpoint.data.id })
      toast.success('Secret rotated')
      secret.reload()
      setRotateSecretDialogOpen(false)
    } catch (error) {
      handleApiError(error, 'Failed to rotate secret')
    }
  }

  const handleReplay = async (msgId: string) => {
    if (!endpointId) return
    try {
      await replayMutation.mutateAsync({ endpointId, msgId })
      toast.success('Event replayed')
      messages.reload()
      stats.reload()
    } catch (error) {
      handleApiError(error, 'Failed to replay event')
    }
  }

  const handleEditSuccess = () => {
    endpoint.reload()
  }

  const endpointData = endpoint.data
  const relativeTime = endpointData ? getRelativeTimeString(endpointData.createdAt).relativeTimeString : null

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Webhooks</PageTitle>
      </PageHeader>

      <PageContent className="gap-6">
        <div className="flex items-center gap-3 min-w-0">
          <Button variant="ghost" size="icon-sm" className="shrink-0" onClick={() => navigate(RoutePath.WEBHOOKS)}>
            <ArrowLeft className="w-4 h-4" />
          </Button>
          {endpoint.loading ? (
            <>
              <Skeleton className="h-6 w-48" />
              <Skeleton className="h-5 w-16" />
              <Skeleton className="h-4 w-24" />
            </>
          ) : endpointData ? (
            <>
              <h2 className="text-lg font-medium truncate min-w-0">{endpointData.description || 'Unnamed Endpoint'}</h2>
              <Badge variant={endpointData.disabled ? 'secondary' : 'success'} className="shrink-0">
                {endpointData.disabled ? 'Disabled' : 'Enabled'}
              </Badge>
              <span className="text-sm text-muted-foreground shrink-0 hidden sm:inline">•</span>
              <TimestampTooltip
                timestamp={
                  typeof endpointData.createdAt === 'string'
                    ? endpointData.createdAt
                    : endpointData.createdAt.toISOString()
                }
              >
                <span className="text-sm text-muted-foreground cursor-default shrink-0 hidden sm:inline">
                  {relativeTime}
                </span>
              </TimestampTooltip>
              <div className="ml-auto flex items-center gap-2 shrink-0">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="outline" size="icon-sm" disabled={isMutating}>
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => setEditDialogOpen(true)} className="cursor-pointer">
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => setDisableDialogOpen(true)} className="cursor-pointer">
                      {endpointData.disabled ? 'Enable' : 'Disable'}
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onClick={() => setRotateSecretDialogOpen(true)} className="cursor-pointer">
                      Rotate Secret
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      variant="destructive"
                      onClick={() => setDeleteDialogOpen(true)}
                      className="cursor-pointer"
                    >
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
                <Button variant="ghost" size="icon-sm" onClick={handleRetry} disabled={isRefreshing}>
                  {isRefreshing ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCcw className="h-4 w-4" />}
                </Button>
              </div>
            </>
          ) : null}
        </div>

        {endpoint.loading ? (
          <div className="flex flex-col gap-4">
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
                <EndpointEventsTable data={[]} loading={true} onReplay={handleReplay} />
              </CardContent>
            </Card>
          </div>
        ) : endpoint.error || !endpointData ? (
          <Card>
            <CardHeader>
              <CardTitle className="text-center">Oops, something went wrong</CardTitle>
            </CardHeader>
            <CardContent className="flex justify-between items-center flex-col gap-3">
              <div>There was an error loading the endpoint details.</div>
              <Button variant="outline" onClick={handleRetry}>
                <RefreshCcw className="mr-2 h-4 w-4" />
                Retry
              </Button>
            </CardContent>
          </Card>
        ) : (
          <div className="flex flex-col gap-4">
            <Card>
              <CardHeader>
                <CardTitle>Endpoint Configuration</CardTitle>
              </CardHeader>
              <CardContent className="p-4 flex flex-col gap-4">
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                  <div className="flex flex-col">
                    <div className="text-muted-foreground text-xs mb-1">URL</div>
                    <InputGroup className="pr-1">
                      <InputGroupInput value={endpointData.url} readOnly className="font-mono text-sm" />
                      <CopyButton value={endpointData.url} size="icon-xs" tooltipText="Copy URL" />
                    </InputGroup>
                  </div>
                  <div className="flex flex-col">
                    <div className="text-muted-foreground text-xs mb-1">Signing Secret</div>
                    {secret.loading ? (
                      <Skeleton className="h-9 w-full" />
                    ) : secret.error ? (
                      <span className="text-sm text-muted-foreground">Failed to load</span>
                    ) : secret.data ? (
                      <InputGroup className="pr-1">
                        <InputGroupInput
                          value={isSecretRevealed ? secret.data.key : getMaskedToken(secret.data.key)}
                          readOnly
                          className="font-mono text-sm"
                        />
                        <InputGroupButton
                          variant="ghost"
                          size="icon-xs"
                          onClick={() => setIsSecretRevealed(!isSecretRevealed)}
                          title={isSecretRevealed ? 'Hide secret' : 'Reveal secret'}
                        >
                          {isSecretRevealed ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                        </InputGroupButton>
                        <CopyButton value={secret.data.key} size="icon-xs" tooltipText="Copy Signing Secret" />
                      </InputGroup>
                    ) : null}
                  </div>
                </div>
                {endpointData.filterTypes && endpointData.filterTypes.length > 0 && (
                  <div>
                    <div className="text-muted-foreground text-xs mb-1">Listening For</div>
                    <div className="flex flex-wrap gap-1.5">
                      {endpointData.filterTypes.map((eventType) => (
                        <Badge key={eventType} variant="secondary" className="font-normal text-xs">
                          {eventType}
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle>Delivery Stats</CardTitle>
              </CardHeader>
              <CardContent>
                {statsIsLoading ? (
                  <div className="flex flex-col gap-2">
                    <Skeleton className="h-2 w-full rounded-full" />
                    <div className="flex items-center gap-4">
                      <Skeleton className="h-3 w-20" />
                      <Skeleton className="h-3 w-16" />
                      <Skeleton className="h-3 w-18" />
                      <Skeleton className="h-3 w-18" />
                    </div>
                  </div>
                ) : stats.data ? (
                  <div className={statsIsFetching ? 'opacity-50 transition-opacity' : ''}>
                    <DeliveryStatsLine stats={stats.data} />
                  </div>
                ) : null}
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle>Event History</CardTitle>
              </CardHeader>
              <CardContent>
                <EndpointEventsTable data={messages.data || []} loading={messages.loading} onReplay={handleReplay} />
              </CardContent>
            </Card>
          </div>
        )}
      </PageContent>

      <EditEndpointDialog
        endpoint={endpoint.data || null}
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
        onSuccess={handleEditSuccess}
      />

      <AlertDialog open={disableDialogOpen} onOpenChange={setDisableDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{endpoint.data?.disabled ? 'Enable' : 'Disable'} Webhook Endpoint</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to {endpoint.data?.disabled ? 'enable' : 'disable'} this webhook endpoint?
              {endpoint.data?.disabled
                ? ' The endpoint will start receiving webhook events again.'
                : ' The endpoint will stop receiving webhook events.'}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDisable}>
              {endpoint.data?.disabled ? 'Enable' : 'Disable'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
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
            <AlertDialogAction variant="destructive" onClick={handleDelete} disabled={deleteMutation.isPending}>
              {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog open={rotateSecretDialogOpen} onOpenChange={setRotateSecretDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Rotate Signing Secret</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to rotate the signing secret? The current secret will be invalidated and you will
              need to update your webhook handler with the new secret.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleRotateSecret} disabled={rotateSecretMutation.isPending}>
              {rotateSecretMutation.isPending ? 'Rotating...' : 'Rotate Secret'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </PageLayout>
  )
}

export default WebhookEndpointDetails
