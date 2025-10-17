/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useEffect, useState } from 'react'
import { Sandbox, SandboxState } from '@daytonaio/api-client'
import { SandboxState as SandboxStateComponent } from './SandboxTable/SandboxState'
import { Button } from '@/components/ui/button'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { getRelativeTimeString } from '@/lib/utils'
import { Archive, Camera, X, GitFork, Trash, Play, Tag, Copy } from 'lucide-react'
import { toast } from 'sonner'

interface SandboxDetailsSheetProps {
  sandbox: Sandbox | null
  open: boolean
  onOpenChange: (open: boolean) => void
  sandboxIsLoading: Record<string, boolean>
  handleStart: (id: string) => void
  handleStop: (id: string) => void
  handleDelete: (id: string) => void
  handleArchive: (id: string) => void
  getWebTerminalUrl: (id: string) => Promise<string | null>
  writePermitted: boolean
  deletePermitted: boolean
}

const SandboxDetailsSheet: React.FC<SandboxDetailsSheetProps> = ({
  sandbox,
  open,
  onOpenChange,
  sandboxIsLoading,
  handleStart,
  handleStop,
  handleDelete,
  handleArchive,
  getWebTerminalUrl,
  writePermitted,
  deletePermitted,
}) => {
  const [terminalUrl, setTerminalUrl] = useState<string | null>(null)

  // TODO: uncomment when we enable the terminal tab
  // useEffect(() => {
  //   const getTerminalUrl = async () => {
  //     if (!sandbox?.id) {
  //       setTerminalUrl(null)
  //       return
  //     }

  //     const url = await getWebTerminalUrl(sandbox.id)
  //     setTerminalUrl(url)
  //   }

  //   getTerminalUrl()
  // }, [sandbox?.id, getWebTerminalUrl])

  if (!sandbox) return null

  const getLastEvent = (sandbox: Sandbox): { date: Date; relativeTimeString: string } => {
    return getRelativeTimeString(sandbox.updatedAt)
  }

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      toast.success('Copied to clipboard')
    } catch (err) {
      console.error('Failed to copy text:', err)
      toast.error('Failed to copy to clipboard')
    }
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-dvw sm:w-[800px] p-0 flex flex-col gap-0 [&>button]:hidden">
        <SheetHeader className="space-y-0 flex flex-row justify-between items-center p-6">
          <SheetTitle className="text-2xl font-medium">Sandbox Details</SheetTitle>
          <div className="flex gap-2 items-center">
            {writePermitted && (
              <>
                {sandbox.state === SandboxState.STARTED && (
                  <Button
                    variant="outline"
                    onClick={() => handleStop(sandbox.id)}
                    disabled={sandboxIsLoading[sandbox.id]}
                  >
                    Stop
                  </Button>
                )}
                {(sandbox.state === SandboxState.STOPPED || sandbox.state === SandboxState.ARCHIVED) && (
                  <Button
                    variant="outline"
                    onClick={() => handleStart(sandbox.id)}
                    disabled={sandboxIsLoading[sandbox.id]}
                  >
                    <Play className="w-4 h-4" />
                    Start
                  </Button>
                )}
                {/* {(sandbox.state === SandboxState.STOPPED || sandbox.state === SandboxState.ARCHIVED) && (
                  <Button
                    variant="outline"
                    onClick={() => handleFork(sandbox.id)}
                    disabled={sandboxIsLoading[sandbox.id]}
                  >
                    <GitFork className="w-4 h-4" />
                    Fork
                  </Button>
                )}
                {(sandbox.state === SandboxState.STOPPED || sandbox.state === SandboxState.ARCHIVED) && (
                  <Button
                    variant="outline"
                    onClick={() => handleSnapshot(sandbox.id)}
                    disabled={sandboxIsLoading[sandbox.id]}
                  >
                    <Camera className="w-4 h-4" />
                    Snapshot
                  </Button>
                )} */}
                {sandbox.state === SandboxState.STOPPED && (
                  <Button
                    variant="outline"
                    className="w-8 h-8"
                    onClick={() => handleArchive(sandbox.id)}
                    disabled={sandboxIsLoading[sandbox.id]}
                  >
                    <Archive className="w-4 h-4" />
                  </Button>
                )}
              </>
            )}
            {deletePermitted && (
              <Button
                variant="outline"
                className="w-8 h-8"
                onClick={() => handleDelete(sandbox.id)}
                disabled={sandboxIsLoading[sandbox.id]}
              >
                <Trash className="w-4 h-4" />
              </Button>
            )}
            <Button
              variant="outline"
              className="w-8 h-8"
              onClick={() => onOpenChange(false)}
              disabled={sandboxIsLoading[sandbox.id]}
            >
              <X className="w-4 h-4" />
            </Button>
          </div>
        </SheetHeader>

        <Tabs defaultValue="overview" className="flex-1 flex flex-col min-h-0">
          {/* TODO: Add terminal tab */}
          {/* <TabsList className="px-4 w-full flex-shrink-0">
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="terminal">Terminal</TabsTrigger>
          </TabsList> */}
          <TabsContent value="overview" className="flex-1 p-6 space-y-10 overflow-y-auto min-h-0">
            <div className="grid grid-cols-2 gap-6">
              <div>
                <h3 className="text-sm text-muted-foreground">Name</h3>
                <div className="mt-1 flex items-center gap-2">
                  <p className="text-sm font-medium truncate">{sandbox.name}</p>
                  <button
                    onClick={() => copyToClipboard(sandbox.name)}
                    className="text-muted-foreground hover:text-foreground transition-colors"
                    aria-label="Copy name"
                  >
                    <Copy className="w-3 h-3" />
                  </button>
                </div>
              </div>
              <div>
                <h3 className="text-sm text-muted-foreground">UUID</h3>
                <div className="mt-1 flex items-center gap-2">
                  <p className="text-sm font-medium truncate">{sandbox.id}</p>
                  <button
                    onClick={() => copyToClipboard(sandbox.id)}
                    className="text-muted-foreground hover:text-foreground transition-colors"
                    aria-label="Copy UUID"
                  >
                    <Copy className="w-3 h-3" />
                  </button>
                </div>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
              <div>
                <h3 className="text-sm text-muted-foreground">State</h3>
                <div className="mt-1 text-sm">
                  <SandboxStateComponent state={sandbox.state} errorReason={sandbox.errorReason} />
                </div>
              </div>
              <div>
                <h3 className="text-sm text-muted-foreground">Snapshot</h3>
                <div className="mt-1 flex items-center gap-2">
                  <p className="text-sm font-medium truncate">{sandbox.snapshot || '-'}</p>
                  {sandbox.snapshot && (
                    <button
                      onClick={() => copyToClipboard(sandbox.snapshot || '')}
                      className="text-muted-foreground hover:text-foreground transition-colors"
                      aria-label="Copy snapshot"
                    >
                      <Copy className="w-3 h-3" />
                    </button>
                  )}
                </div>
              </div>
              <div>
                <h3 className="text-sm text-muted-foreground">Region</h3>
                <div className="mt-1 flex items-center gap-2">
                  <p className="text-sm font-medium truncate">{sandbox.target}</p>
                  <button
                    onClick={() => copyToClipboard(sandbox.target)}
                    className="text-muted-foreground hover:text-foreground transition-colors"
                    aria-label="Copy region"
                  >
                    <Copy className="w-3 h-3" />
                  </button>
                </div>
              </div>
              <div>
                <h3 className="text-sm text-muted-foreground">Last used</h3>
                <p className="mt-1 text-sm font-medium">{getLastEvent(sandbox).relativeTimeString}</p>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div>
                <h3 className="text-sm text-muted-foreground">Resources</h3>
                <div className="mt-1 text-sm font-medium flex items-center gap-1">
                  <div className="flex items-center gap-1 bg-blue-100 text-blue-600 dark:bg-blue-950 dark:text-blue-200 rounded-full px-2">
                    {sandbox.cpu} vCPU
                  </div>
                  <div className="flex items-center gap-1 bg-blue-100 text-blue-600 dark:bg-blue-950 dark:text-blue-200 rounded-full px-2">
                    {sandbox.memory} GiB
                  </div>
                  <div className="flex items-center gap-1 bg-blue-100 text-blue-600 dark:bg-blue-950 dark:text-blue-200 rounded-full px-2">
                    {sandbox.disk} GiB
                  </div>
                </div>
              </div>
            </div>
            <div>
              <h3 className="text-lg font-medium">Labels</h3>
              <div className="mt-3 space-y-4">
                {Object.entries(sandbox.labels ?? {}).length > 0 ? (
                  Object.entries(sandbox.labels ?? {}).map(([key, value]) => (
                    <div key={key} className="text-sm">
                      <div>{key}</div>
                      <div className="font-medium p-2 bg-muted rounded-md mt-1 border border-border">{value}</div>
                    </div>
                  ))
                ) : (
                  <div className="flex flex-col border border-border rounded-md items-center justify-center gap-2 text-muted-foreground w-full min-h-40">
                    <Tag className="w-4 h-4" />
                    <span className="text-sm">No labels found</span>
                  </div>
                )}
              </div>
            </div>
          </TabsContent>

          <TabsContent value="terminal" className="p-4">
            <iframe title="Terminal" src={terminalUrl || undefined} className="w-full h-full"></iframe>
          </TabsContent>
        </Tabs>
      </SheetContent>
    </Sheet>
  )
}

export default SandboxDetailsSheet
