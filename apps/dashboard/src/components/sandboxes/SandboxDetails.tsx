/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationSuspendedError } from '@/api/errors'
import { PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
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
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { FeatureFlags } from '@/enums/FeatureFlags'
import { RoutePath } from '@/enums/RoutePath'
import { useArchiveSandboxMutation } from '@/hooks/mutations/useArchiveSandboxMutation'
import { useDeleteSandboxMutation } from '@/hooks/mutations/useDeleteSandboxMutation'
import { useRecoverSandboxMutation } from '@/hooks/mutations/useRecoverSandboxMutation'
import { useStartSandboxMutation } from '@/hooks/mutations/useStartSandboxMutation'
import { useStopSandboxMutation } from '@/hooks/mutations/useStopSandboxMutation'
import { useSandboxQuery } from '@/hooks/queries/useSandboxQuery'
import { useApi } from '@/hooks/useApi'
import { useConfig } from '@/hooks/useConfig'
import { useMatchMedia } from '@/hooks/useMatchMedia'
import { useRegions } from '@/hooks/useRegions'
import { useSandboxWsSync } from '@/hooks/useSandboxWsSync'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { isStoppable, isTransitioning } from '@/lib/utils/sandbox'
import { SandboxSessionProvider } from '@/providers/SandboxSessionProvider'
import { OrganizationRolePermissionsEnum, OrganizationUserRoleEnum } from '@daytonaio/api-client'
import { isAxiosError } from 'axios'
import { Container, RefreshCw } from 'lucide-react'
import { useQueryState } from 'nuqs'
import { useFeatureFlagEnabled } from 'posthog-js/react'
import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { toast } from 'sonner'
import { CreateSshAccessDialog } from './CreateSshAccessDialog'
import { RevokeSshAccessDialog } from './RevokeSshAccessDialog'
import { SandboxHeader } from './SandboxHeader'
import { InfoPanelSkeleton, SandboxInfoPanel } from './SandboxInfoPanel'
import { SandboxLogsTab } from './SandboxLogsTab'
import { SandboxMetricsTab } from './SandboxMetricsTab'
import { SandboxSpendingTab } from './SandboxSpendingTab'
import { SandboxTerminalTab } from './SandboxTerminalTab'
import { SandboxTracesTab } from './SandboxTracesTab'
import { SandboxVncTab } from './SandboxVncTab'
import { tabParser, TabValue } from './SearchParams'

export default function SandboxDetails() {
  const { sandboxId } = useParams<{ sandboxId: string }>()
  const navigate = useNavigate()
  const config = useConfig()
  const { sandboxApi } = useApi()
  const { authenticatedUserOrganizationMember, selectedOrganization, authenticatedUserHasPermission } =
    useSelectedOrganization()
  const { getRegionName } = useRegions()

  const experimentsEnabled = useFeatureFlagEnabled(FeatureFlags.ORGANIZATION_EXPERIMENTS)

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [createSshDialogOpen, setCreateSshDialogOpen] = useState(false)
  const [revokeSshDialogOpen, setRevokeSshDialogOpen] = useState(false)
  const [tab, setTab] = useQueryState('tab', tabParser)
  const isDesktop = useMatchMedia('(min-width: 1024px)')

  // On desktop (lg+), the overview tab is hidden in the sidebar, so switch to a content tab
  useEffect(() => {
    if (isDesktop && tab === 'overview') {
      setTab(experimentsEnabled ? 'logs' : 'terminal')
    }
  }, [isDesktop, tab, setTab, experimentsEnabled])

  // When experiments are disabled, coerce experimental tabs back to a supported default
  useEffect(() => {
    if (!experimentsEnabled && (tab === 'logs' || tab === 'traces' || tab === 'metrics' || tab === 'spending')) {
      setTab('terminal')
    }
  }, [experimentsEnabled, tab, setTab])

  const { data: sandbox, isLoading, isError, error, refetch, isFetching } = useSandboxQuery(sandboxId ?? '')
  const isNotFound = isError && isAxiosError(error.cause) && error.cause?.status === 404

  useSandboxWsSync({ sandboxId })

  const startMutation = useStartSandboxMutation()
  const stopMutation = useStopSandboxMutation()
  const archiveMutation = useArchiveSandboxMutation()
  const recoverMutation = useRecoverSandboxMutation()
  const deleteMutation = useDeleteSandboxMutation()

  const writePermitted = authenticatedUserHasPermission(OrganizationRolePermissionsEnum.WRITE_SANDBOXES)
  const deletePermitted = authenticatedUserHasPermission(OrganizationRolePermissionsEnum.DELETE_SANDBOXES)
  const transitioning = sandbox ? isTransitioning(sandbox) : false
  const anyMutating =
    startMutation.isPending ||
    stopMutation.isPending ||
    archiveMutation.isPending ||
    recoverMutation.isPending ||
    deleteMutation.isPending
  const actionsDisabled = anyMutating || transitioning

  const handleStart = async () => {
    if (!sandbox) return
    try {
      await startMutation.mutateAsync({ sandboxId: sandbox.id })
      toast.success('Sandbox started')
    } catch (error) {
      handleApiError(
        error,
        'Failed to start sandbox',
        error instanceof OrganizationSuspendedError &&
          config.billingApiUrl &&
          authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER ? (
          <Button variant="secondary" onClick={() => navigate(RoutePath.BILLING_WALLET)}>
            Go to billing
          </Button>
        ) : undefined,
      )
    }
  }

  const handleStop = async () => {
    if (!sandbox) return
    try {
      await stopMutation.mutateAsync({ sandboxId: sandbox.id })
      toast.success('Sandbox stopped')
    } catch (error) {
      handleApiError(error, 'Failed to stop sandbox')
    }
  }

  const handleArchive = async () => {
    if (!sandbox) return
    try {
      await archiveMutation.mutateAsync({ sandboxId: sandbox.id })
      toast.success('Sandbox archived')
    } catch (error) {
      handleApiError(error, 'Failed to archive sandbox')
    }
  }

  const handleRecover = async () => {
    if (!sandbox) return
    try {
      await recoverMutation.mutateAsync({ sandboxId: sandbox.id })
      toast.success('Sandbox recovery started')
    } catch (error) {
      handleApiError(error, 'Failed to recover sandbox')
    }
  }

  const handleDelete = async () => {
    if (!sandbox) return
    try {
      await deleteMutation.mutateAsync({ sandboxId: sandbox.id })
      toast.success('Sandbox deleted')
      setDeleteDialogOpen(false)
      navigate(RoutePath.SANDBOXES)
    } catch (error) {
      handleApiError(error, 'Failed to delete sandbox')
    }
  }

  const handleScreenRecordings = async () => {
    if (!sandbox || !isStoppable(sandbox)) {
      toast.error('Sandbox must be started to access Screen Recordings')
      return
    }
    try {
      const response = await sandboxApi.getSignedPortPreviewUrl(sandbox.id, 33333, selectedOrganization?.id)
      window.open(response.data.url, '_blank', 'noopener,noreferrer')
      toast.success('Opening Screen Recordings dashboard...')
    } catch (error) {
      handleApiError(error, 'Failed to open Screen Recordings')
    }
  }

  return (
    <SandboxSessionProvider>
      <PageLayout className="max-h-screen overflow-hidden">
        <PageHeader>
          <PageTitle>Sandboxes</PageTitle>
        </PageHeader>

        <SandboxHeader
          sandbox={sandbox}
          isLoading={isLoading}
          writePermitted={writePermitted}
          deletePermitted={deletePermitted}
          actionsDisabled={actionsDisabled}
          isFetching={isFetching}
          onStart={handleStart}
          onStop={handleStop}
          onArchive={handleArchive}
          onRecover={handleRecover}
          onDelete={() => setDeleteDialogOpen(true)}
          onRefresh={() => refetch()}
          onBack={() => navigate(RoutePath.SANDBOXES)}
          onCreateSshAccess={() => setCreateSshDialogOpen(true)}
          onRevokeSshAccess={() => setRevokeSshDialogOpen(true)}
          onScreenRecordings={handleScreenRecordings}
          mutations={{
            start: startMutation.isPending,
            stop: stopMutation.isPending,
            archive: archiveMutation.isPending,
            recover: recoverMutation.isPending,
          }}
        />

        {isNotFound ? (
          <div className="flex flex-1 min-h-0 items-center justify-center">
            <Empty>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <Container className="size-4" />
                </EmptyMedia>
                <EmptyTitle>Sandbox not found</EmptyTitle>
                <EmptyDescription>Are you sure you're in the right organization?</EmptyDescription>
              </EmptyHeader>
              <Button variant="outline" size="sm" onClick={() => navigate(RoutePath.SANDBOXES)}>
                Back to Sandboxes
              </Button>
            </Empty>
          </div>
        ) : (
          <div className="flex flex-1 min-h-0 overflow-hidden">
            <aside className="hidden lg:flex w-72 shrink-0 border-r border-border flex-col overflow-hidden">
              <div className="flex items-center px-5 border-b border-border shrink-0 h-[41px]">
                <span className="text-sm font-medium">Overview</span>
              </div>
              <ScrollArea fade="mask" className="flex-1 min-h-0">
                {isLoading ? (
                  <InfoPanelSkeleton />
                ) : isError || !sandbox ? (
                  <div className="flex flex-col items-center justify-center gap-3 p-8 text-center text-muted-foreground">
                    <p className="text-sm">Failed to load sandbox details.</p>
                    <Button variant="outline" size="sm" onClick={() => refetch()}>
                      <RefreshCw className="size-4" />
                      Retry
                    </Button>
                  </div>
                ) : (
                  <SandboxInfoPanel sandbox={sandbox} getRegionName={getRegionName} />
                )}
              </ScrollArea>
            </aside>

            <div className="flex-1 min-w-0 flex flex-col overflow-hidden">
              {isLoading ? (
                <div className="flex flex-col h-full">
                  <div className="flex items-center gap-0 border-b border-border h-[41px] px-4 shrink-0">
                    <Skeleton className="h-4 w-16 lg:hidden" />
                    <Skeleton className="h-4 w-10 ml-4 lg:ml-0" />
                    <Skeleton className="h-4 w-12 ml-4" />
                    <Skeleton className="h-4 w-14 ml-4" />
                    <Skeleton className="h-4 w-16 ml-4" />
                    <Skeleton className="h-4 w-10 ml-4" />
                  </div>
                  <div className="flex-1 flex items-center justify-center text-muted-foreground">
                    <Spinner className="size-5" />
                  </div>
                </div>
              ) : !sandbox ? null : (
                <Tabs value={tab} onValueChange={(v) => setTab(v as TabValue)} className="flex flex-col h-full gap-0">
                  <TabsList variant="underline" className="h-[41px] overflow-x-auto overflow-y-hidden scrollbar-sm">
                    <TabsTrigger value="overview" className="lg:hidden">
                      Overview
                    </TabsTrigger>
                    {experimentsEnabled && (
                      <>
                        <TabsTrigger value="logs">Logs</TabsTrigger>
                        <TabsTrigger value="traces">Traces</TabsTrigger>
                        <TabsTrigger value="metrics">Metrics</TabsTrigger>
                        <TabsTrigger value="spending">Spending</TabsTrigger>
                      </>
                    )}
                    <TabsTrigger value="terminal">Terminal</TabsTrigger>
                    <TabsTrigger value="vnc">VNC</TabsTrigger>
                  </TabsList>

                  <TabsContent value="overview" className="flex-1 min-h-0 m-0 overflow-y-auto scrollbar-sm lg:hidden">
                    <SandboxInfoPanel sandbox={sandbox} getRegionName={getRegionName} />
                  </TabsContent>
                  {experimentsEnabled && (
                    <>
                      <TabsContent
                        value="logs"
                        className="flex-1 min-h-0 m-0 data-[state=active]:flex flex-col overflow-hidden"
                      >
                        <SandboxLogsTab sandboxId={sandbox.id} />
                      </TabsContent>
                      <TabsContent
                        value="traces"
                        className="flex-1 min-h-0 m-0 data-[state=active]:flex flex-col overflow-hidden"
                      >
                        <SandboxTracesTab sandboxId={sandbox.id} />
                      </TabsContent>
                      <TabsContent
                        value="metrics"
                        className="flex-1 min-h-0 m-0 data-[state=active]:flex flex-col overflow-hidden"
                      >
                        <SandboxMetricsTab sandboxId={sandbox.id} />
                      </TabsContent>
                      <TabsContent
                        value="spending"
                        className="flex-1 min-h-0 m-0 data-[state=active]:flex flex-col overflow-hidden"
                      >
                        <SandboxSpendingTab sandboxId={sandbox.id} />
                      </TabsContent>
                    </>
                  )}
                  <TabsContent
                    value="terminal"
                    className="flex-1 min-h-0 m-0 data-[state=active]:flex flex-col overflow-hidden"
                  >
                    <SandboxTerminalTab sandbox={sandbox} />
                  </TabsContent>
                  <TabsContent
                    value="vnc"
                    className="flex-1 min-h-0 m-0 data-[state=active]:flex flex-col overflow-hidden"
                  >
                    <SandboxVncTab sandbox={sandbox} />
                  </TabsContent>
                </Tabs>
              )}
            </div>
          </div>
        )}

        <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Delete Sandbox</AlertDialogTitle>
              <AlertDialogDescription>
                Are you sure you want to delete this sandbox? This action cannot be undone.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel disabled={deleteMutation.isPending}>Cancel</AlertDialogCancel>
              <AlertDialogAction variant="destructive" onClick={handleDelete} disabled={deleteMutation.isPending}>
                {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>

        {sandboxId && (
          <>
            <CreateSshAccessDialog
              sandboxId={sandboxId}
              open={createSshDialogOpen}
              onOpenChange={setCreateSshDialogOpen}
            />
            <RevokeSshAccessDialog
              sandboxId={sandboxId}
              open={revokeSshDialogOpen}
              onOpenChange={setRevokeSshDialogOpen}
            />
          </>
        )}
      </PageLayout>
    </SandboxSessionProvider>
  )
}
