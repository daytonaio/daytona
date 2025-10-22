/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState, useRef, useEffect } from 'react'
import { useSandboxLogs } from '@/hooks/useSandboxLogs'
import { Loader2, RefreshCw, Archive, Trash2, Clock } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { SandboxState } from '@daytonaio/api-client'

interface SandboxLogsProps {
  sandboxId: string
  sandboxState?: SandboxState
}

// Function to parse logs with timestamps and format them for display
const parseLogsWithTimestamps = (logs: string, showTimestamps: boolean): React.ReactElement[] => {
  if (!logs) return []

  const lines = logs.split('\n')
  const parsedLines: React.ReactElement[] = []

  lines.forEach((line, index) => {
    if (line.trim() === '') {
      parsedLines.push(<br key={index} />)
      return
    }

    // Check if line starts with a timestamp (Docker format: YYYY-MM-DDTHH:MM:SS.sssssssssZ)
    const timestampRegex = /^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)\s+(.*)$/
    const match = line.match(timestampRegex)

    if (match) {
      const [, timestamp, content] = match
      if (showTimestamps) {
        // Parse UTC timestamp and format it in browser's timezone
        const date = new Date(timestamp)
        const formattedTimestamp = date
          .toLocaleString('en-US', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
            hour12: false,
          })
          .replace(',', '')
        parsedLines.push(
          <div key={index} className="flex items-baseline gap-4">
            <span className="text-gray-500 text-xs font-mono flex-shrink-0">{formattedTimestamp}</span>
            <span className="text-green-400 font-mono text-sm leading-relaxed whitespace-pre-wrap">{content}</span>
          </div>,
        )
      } else {
        // Show only content without timestamp
        parsedLines.push(
          <div key={index} className="text-green-400 font-mono text-sm leading-relaxed whitespace-pre-wrap">
            {content}
          </div>,
        )
      }
    } else {
      // Line without timestamp
      parsedLines.push(
        <div key={index} className="text-green-400 font-mono text-sm leading-relaxed whitespace-pre-wrap">
          {line}
        </div>,
      )
    }
  })

  return parsedLines
}

const SandboxLogs: React.FC<SandboxLogsProps> = ({ sandboxId, sandboxState }) => {
  const { data: logs, isLoading, error, refetch, isRefetching } = useSandboxLogs(sandboxId)
  const [showTimestamps, setShowTimestamps] = useState(true)
  const scrollContainerRef = useRef<HTMLDivElement>(null)

  const handleTimestampToggle = (checked: boolean) => {
    setShowTimestamps(checked)
    refetch() // Refresh logs when toggling timestamp display
  }

  // Auto-scroll to bottom when logs change
  useEffect(() => {
    if (scrollContainerRef.current && logs) {
      scrollContainerRef.current.scrollTop = scrollContainerRef.current.scrollHeight
    }
  }, [logs, showTimestamps])

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
          <span>Sandbox Entrypoint Logs</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex items-center gap-1">
            <Clock className="w-3 h-3 text-gray-400" />
            <Switch
              checked={showTimestamps}
              onCheckedChange={handleTimestampToggle}
              className="data-[state=checked]:bg-green-400 scale-75"
            />
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
      </div>
      <div ref={scrollContainerRef} className="flex-1 overflow-auto p-3">
        <div className="space-y-1">
          {logs ? (
            parseLogsWithTimestamps(logs, showTimestamps)
          ) : (
            <span className="text-green-400 font-mono text-sm">No logs available</span>
          )}
        </div>
      </div>
    </div>
  )
}

export default SandboxLogs
