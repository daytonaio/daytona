/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import { useSandboxLogs } from '@/hooks/useSandboxLogs'
import { Loader2, RefreshCw, Archive, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { SandboxState } from '@daytonaio/api-client'

interface SandboxLogsProps {
  sandboxId: string
  sandboxState?: SandboxState
}

const SandboxLogs: React.FC<SandboxLogsProps> = ({ sandboxId, sandboxState }) => {
  const { data: logs, isLoading, error, refetch, isRefetching } = useSandboxLogs(sandboxId)

  // Show appropriate message for archived or destroyed sandboxes
  if (sandboxState === SandboxState.ARCHIVED) {
    return (
      <div className="flex flex-col items-center justify-center h-96 bg-black font-mono text-sm">
        <div className="text-center">
          <Archive className="w-12 h-12 mb-4 mx-auto" />
          <p className="mb-2 text-lg font-semibold">Sandbox Archived</p>
          <p className="text-sm text-gray-400">Logs are not available for archived sandboxes</p>
        </div>
      </div>
    )
  }

  if (sandboxState === SandboxState.DESTROYED) {
    return (
      <div className="flex flex-col items-center justify-center h-96 bg-black font-mono text-sm">
        <div className="text-center">
          <Trash2 className="w-12 h-12 mb-4 mx-auto" />
          <p className="mb-2 text-lg font-semibold">Sandbox Destroyed</p>
          <p className="text-sm text-gray-400">Logs are not available for destroyed sandboxes</p>
        </div>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-96 bg-black text-green-400 font-mono text-sm">
        <div className="flex items-center gap-2">
          <Loader2 className="w-4 h-4 animate-spin" />
          <span>Loading logs...</span>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center h-96 bg-black text-red-400 font-mono text-sm">
        <div className="text-center">
          <p className="mb-4">Failed to load entrypoint logs</p>
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetch()}
            disabled={isRefetching}
            className="bg-gray-800 border-gray-600 text-green-400 hover:bg-gray-700"
          >
            {isRefetching ? <Loader2 className="w-4 h-4 animate-spin" /> : <RefreshCw className="w-4 h-4" />}
            Retry
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-96 bg-black border border-gray-700 rounded-md">
      <div className="flex items-center justify-between p-2 bg-gray-800 border-b border-gray-700">
        <div className="flex items-center gap-2 text-green-400 font-mono text-xs">
          <div className="w-2 h-2 bg-green-400 rounded-full"></div>
          <span>Sandbox Logs</span>
        </div>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => refetch()}
          disabled={isRefetching}
          className="h-6 w-6 p-0 text-green-400 hover:bg-gray-700"
        >
          {isRefetching ? <Loader2 className="w-3 h-3 animate-spin" /> : <RefreshCw className="w-3 h-3" />}
        </Button>
      </div>
      <div className="flex-1 overflow-auto p-3">
        <pre className="text-green-400 font-mono text-sm leading-relaxed whitespace-pre-wrap">
          {logs || 'No logs available'}
        </pre>
      </div>
    </div>
  )
}

export default SandboxLogs
