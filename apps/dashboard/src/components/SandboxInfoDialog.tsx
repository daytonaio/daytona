/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState } from 'react'
import { Copy, Check } from 'lucide-react'
import { Sandbox } from '@daytonaio/api-client'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'

interface SandboxInfoDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  sandbox: Sandbox | null
}

export const SandboxInfoDialog: React.FC<SandboxInfoDialogProps> = ({ open, onOpenChange, sandbox }) => {
  const [copied, setCopied] = useState(false)

  if (!sandbox) return null

  const formatSandboxInfo = (sandbox: Sandbox) => {
    // Note: Sandbox type doesn't have providerMetadata like Workspace had
    // So we'll create the object without trying to parse non-existent metadata

    return {
      id: sandbox.id,
      organizationId: sandbox.organizationId,
      state: sandbox.state,
      target: sandbox.target,
      user: sandbox.user,
      cpu: sandbox.cpu,
      gpu: sandbox.gpu,
      memory: sandbox.memory,
      disk: sandbox.disk,
      public: sandbox.public,
      env: sandbox.env,
      labels: sandbox.labels,
      volumes: sandbox.volumes,
      errorReason: sandbox.errorReason,
      autoStopInterval: sandbox.autoStopInterval,
      autoArchiveInterval: sandbox.autoArchiveInterval,
      created: sandbox.createdAt,
      updatedAt: sandbox.updatedAt,
      snapshot: sandbox.snapshot,
      runnerDomain: sandbox.runnerDomain,
      backupState: sandbox.backupState,
      backupCreatedAt: sandbox.backupCreatedAt,
      buildInfo: sandbox.buildInfo,
    }
  }

  const copyId = async () => {
    try {
      const sandboxInfo = formatSandboxInfo(sandbox)
      const formattedText = JSON.stringify(sandboxInfo, null, 2)
      await navigator.clipboard.writeText(formattedText)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy to clipboard:', err)
    }
  }

  const getStateColor = (state?: string) => {
    switch (state) {
      case 'started':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300 border-green-300 dark:border-green-700'
      case 'error':
      case 'stopped':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300 border-red-300 dark:border-red-700'
      case 'archiving':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300 border-yellow-300 dark:border-yellow-700'
      case 'archived':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300 border-blue-300 dark:border-blue-700'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300 border-gray-300 dark:border-gray-700'
    }
  }

  const formatLabels = (labels?: { [key: string]: string }) => {
    if (!labels || Object.keys(labels).length === 0) return 'None'
    return Object.entries(labels)
      .map(([key, value]) => `${key}: ${value}`)
      .join(', ')
  }

  const formatEnv = (env?: { [key: string]: string }) => {
    if (!env || Object.keys(env).length === 0) return 'None'
    return Object.entries(env)
      .map(([key, value]) => `${key}=${value}`)
      .join(', ')
  }

  const formatVolumes = (volumes?: Array<{ volumeId: string; mountPath: string }>) => {
    if (!volumes || volumes.length === 0) return 'None'
    return volumes.map((volume) => `${volume.volumeId} → ${volume.mountPath}`).join(', ')
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-6xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Sandbox Information</DialogTitle>
          <DialogDescription>Complete overview of the Sandbox details, resources and configuration</DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Left Column */}
          <div className="space-y-4">
            {/* Basic Information */}
            <div>
              <h3 className="text-sm font-semibold mb-2">Basic Information</h3>
              <div className="grid grid-cols-[100px_1fr] gap-2 text-sm">
                <span className="text-muted-foreground">ID:</span>
                <span className="font-mono break-all">{sandbox.id}</span>

                <span className="text-muted-foreground">Org ID:</span>
                <span className="font-mono break-all">{sandbox.organizationId}</span>

                <span className="text-muted-foreground">State:</span>
                <span
                  className={`inline-flex items-center px-2 py-1 rounded-md text-xs font-medium border w-fit ${getStateColor(sandbox.state)}`}
                >
                  {sandbox.state?.charAt(0).toUpperCase() + (sandbox.state?.slice(1) || '')}
                </span>

                <span className="text-muted-foreground">Target:</span>
                <span>{sandbox.target}</span>

                <span className="text-muted-foreground">User:</span>
                <span>{sandbox.user}</span>

                <span className="text-muted-foreground">Snapshot:</span>
                <span className="font-mono break-all">{sandbox.snapshot || 'Not specified'}</span>

                <span className="text-muted-foreground">Public:</span>
                <span>{sandbox.public ? 'Yes' : 'No'}</span>
              </div>
            </div>

            <Separator />

            {/* Resources */}
            <div>
              <h3 className="text-sm font-semibold mb-2">Resources</h3>
              <div className="grid grid-cols-[100px_1fr] gap-2 text-sm">
                <span className="text-muted-foreground">CPU:</span>
                <span>{sandbox.cpu ? `${sandbox.cpu} cores` : 'Not specified'}</span>

                <span className="text-muted-foreground">Memory:</span>
                <span>{sandbox.memory ? `${sandbox.memory} GB` : 'Not specified'}</span>

                <span className="text-muted-foreground">Disk:</span>
                <span>{sandbox.disk ? `${sandbox.disk} GB` : 'Not specified'}</span>

                <span className="text-muted-foreground">GPU:</span>
                <span>{sandbox.gpu ? `${sandbox.gpu} units` : 'Not specified'}</span>
              </div>
            </div>
          </div>

          {/* Right Column */}
          <div className="space-y-4">
            {/* Configuration */}
            <div>
              <h3 className="text-sm font-semibold mb-2">Configuration</h3>
              <div className="grid grid-cols-[100px_1fr] gap-2 text-sm">
                <span className="text-muted-foreground">Labels:</span>
                <span className="break-words">{formatLabels(sandbox.labels)}</span>

                <span className="text-muted-foreground">Environment:</span>
                <span className="break-words font-mono">{formatEnv(sandbox.env)}</span>

                <span className="text-muted-foreground">Volumes:</span>
                <span className="break-words font-mono">{formatVolumes(sandbox.volumes)}</span>

                <span className="text-muted-foreground">Auto-stop:</span>
                <span>{sandbox.autoStopInterval ? `${sandbox.autoStopInterval} minutes` : 'Disabled'}</span>

                <span className="text-muted-foreground">Auto-archive:</span>
                <span>{sandbox.autoArchiveInterval ? `${sandbox.autoArchiveInterval} minutes` : 'Disabled'}</span>
              </div>
            </div>

            {/* Error Information */}
            {sandbox.errorReason && (
              <>
                <Separator />
                <div>
                  <h3 className="text-sm font-semibold mb-2 text-red-600 dark:text-red-400">Error Information</h3>
                  <div className="text-sm">
                    <span className="text-red-600 dark:text-red-400 break-words">{sandbox.errorReason}</span>
                  </div>
                </div>
              </>
            )}
          </div>
        </div>

        <DialogFooter className="flex-col sm:flex-row gap-2">
          <Button onClick={copyId} className="flex items-center gap-2">
            {copied ? (
              <>
                <Check className="w-4 h-4" />
                Copied!
              </>
            ) : (
              <>
                <Copy className="w-4 h-4" />
                Copy ID
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
