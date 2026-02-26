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
import { useConfig } from '@/hooks/useConfig'
import { useMatchMedia } from '@/hooks/useMatchMedia'
import { useRegions } from '@/hooks/useRegions'
import { useSandboxWsSync } from '@/hooks/useSandboxWsSync'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { isTransitioning } from '@/lib/utils/sandbox'
import { SandboxSessionProvider } from '@/providers/SandboxSessionProvider'
import { OrganizationUserRoleEnum } from '@daytonaio/api-client'
import { RefreshCw } from 'lucide-react'
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
import { SandboxTerminalTab } from './SandboxTerminalTab'
import { SandboxTracesTab } from './SandboxTracesTab'
import { SandboxVncTab } from './SandboxVncTab'
import { tabParser, TabValue } from './SearchParams'

const TAB_TRIGGER =
  'rounded-none border-b-2 border-transparent data-[state=active]:border-foreground data-[state=active]:bg-transparent data-[state=active]:shadow-none px-4 py-2.5 text-sm'
const TABS_LIST = 'w-full bg-transparent border-b border-border rounded-none h-auto p-0 gap-0 justify-start shrink-0'

export default function SandboxDetails() {
  const { sandboxId } = useParams<{ sandboxId: string }>()
  const navigate = useNavigate()
  const config = useConfig()
  const { authenticatedUserOrganizationMember } = useSelectedOrganization()
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

  const { data: sandbox, isLoading, isError, refetch, isFetching } = useSandboxQuery(sandboxId ?? '')

  useSandboxWsSync({ sandboxId })

  const startMutation = useStartSandboxMutation()
  const stopMutation = useStopSandboxMutation()
  const archiveMutation = useArchiveSandboxMutation()
  const recoverMutation = useRecoverSandboxMutation()
  const deleteMutation = useDeleteSandboxMutation()

  const isOwner = authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER
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

  return (
    <SandboxSessionProvider>
      <PageLayout className="max-h-screen overflow-hidden">
        <PageHeader>
          <PageTitle>Sandboxes</PageTitle>
        </PageHeader>

        <SandboxHeader
          sandbox={sandbox}
          isLoading={isLoading}
          isOwner={isOwner}
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
          mutations={{
            start: startMutation.isPending,
            stop: stopMutation.isPending,
            archive: archiveMutation.isPending,
            recover: recoverMutation.isPending,
          }}
        />

        <div className="flex flex-1 min-h-0 overflow-hidden">
          <aside className="hidden lg:flex w-72 shrink-0 border-r border-border flex-col overflow-hidden">
            <div className="flex items-center px-5 border-b border-border shrink-0 h-[41px]">
              <span className="text-sm font-medium">Overview</span>
            </div>
            <ScrollArea fade="mask" className="flex-1">
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
                <TabsList className={`${TABS_LIST} h-[41px] overflow-x-auto overflow-y-hidden scrollbar-sm`}>
                  <TabsTrigger value="overview" className={`${TAB_TRIGGER} lg:hidden`}>
                    Overview
                  </TabsTrigger>
                  {experimentsEnabled && (
                    <>
                      <TabsTrigger value="logs" className={TAB_TRIGGER}>
                        Logs
                      </TabsTrigger>
                      <TabsTrigger value="traces" className={TAB_TRIGGER}>
                        Traces
                      </TabsTrigger>
                      <TabsTrigger value="metrics" className={TAB_TRIGGER}>
                        Metrics
                      </TabsTrigger>
                    </>
                  )}
                  <TabsTrigger value="terminal" className={TAB_TRIGGER}>
                    Terminal
                  </TabsTrigger>
                  <TabsTrigger value="vnc" className={TAB_TRIGGER}>
                    VNC
                  </TabsTrigger>
                </TabsList>

                <TabsContent value="overview" className="relative flex-1 min-h-0 m-0 lg:hidden">
                  <div className="absolute inset-0 overflow-y-auto scrollbar-sm">
                    <SandboxInfoPanel sandbox={sandbox} getRegionName={getRegionName} />
                  </div>
                </TabsContent>

                {experimentsEnabled && (
                  <>
                    <TabsContent value="logs" className="relative flex-1 min-h-0 m-0">
                      <div className="absolute inset-0 flex flex-col overflow-hidden">
                        <SandboxLogsTab sandboxId={sandbox.id} />
                      </div>
                    </TabsContent>
                    <TabsContent value="traces" className="relative flex-1 min-h-0 m-0">
                      <div className="absolute inset-0 flex flex-col overflow-hidden">
                        <SandboxTracesTab sandboxId={sandbox.id} />
                      </div>
                    </TabsContent>
                    <TabsContent value="metrics" className="relative flex-1 min-h-0 m-0">
                      <div className="absolute inset-0 flex flex-col overflow-hidden">
                        <SandboxMetricsTab sandboxId={sandbox.id} />
                      </div>
                    </TabsContent>
                  </>
                )}
                <TabsContent value="terminal" className="relative flex-1 min-h-0 m-0">
                  <div className="absolute inset-0 flex flex-col overflow-hidden">
                    <SandboxTerminalTab sandbox={sandbox} />
                  </div>
                </TabsContent>
                <TabsContent value="vnc" className="relative flex-1 min-h-0 m-0">
                  <div className="absolute inset-0 flex flex-col overflow-hidden">
                    <SandboxVncTab sandbox={sandbox} />
                  </div>
                </TabsContent>
              </Tabs>
            )}
          </div>
        </div>

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
