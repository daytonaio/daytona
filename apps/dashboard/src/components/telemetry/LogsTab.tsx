/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState, useCallback } from 'react'
import { useSandboxLogs, LogsQueryParams } from '@/hooks/useSandboxLogs'
import { TimeRangeSelector } from './TimeRangeSelector'
import { SeverityBadge } from './SeverityBadge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ChevronLeft, ChevronRight, Search, FileText, RefreshCw, ChevronDown, ChevronUp } from 'lucide-react'
import { Spinner } from '@/components/ui/spinner'
import { format } from 'date-fns'
import { subHours } from 'date-fns'
import { LogEntry } from '@daytonaio/api-client'

interface LogsTabProps {
  sandboxId: string
}

const SEVERITY_OPTIONS = ['DEBUG', 'INFO', 'WARN', 'ERROR']

export const LogsTab: React.FC<LogsTabProps> = ({ sandboxId }) => {
  const [timeRange, setTimeRange] = useState(() => {
    const now = new Date()
    return { from: subHours(now, 1), to: now }
  })
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [searchInput, setSearchInput] = useState('')
  const [selectedSeverities, setSelectedSeverities] = useState<string[]>([])
  const [expandedRow, setExpandedRow] = useState<number | null>(null)
  const limit = 50

  const queryParams: LogsQueryParams = {
    from: timeRange.from,
    to: timeRange.to,
    page,
    limit,
    severities: selectedSeverities.length > 0 ? selectedSeverities : undefined,
    search: search || undefined,
  }

  const { data, isLoading, refetch } = useSandboxLogs(sandboxId, queryParams)

  const handleTimeRangeChange = useCallback((from: Date, to: Date) => {
    setTimeRange({ from, to })
    setPage(1)
  }, [])

  const handleSearch = () => {
    setSearch(searchInput)
    setPage(1)
  }

  const handleSeverityChange = (severity: string) => {
    setSelectedSeverities((prev) =>
      prev.includes(severity) ? prev.filter((s) => s !== severity) : [...prev, severity],
    )
    setPage(1)
  }

  const toggleRowExpansion = (index: number) => {
    setExpandedRow(expandedRow === index ? null : index)
  }

  const formatTimestamp = (timestamp: string) => {
    try {
      return format(new Date(timestamp), 'yyyy-MM-dd HH:mm:ss.SSS')
    } catch {
      return timestamp
    }
  }

  return (
    <div className="flex flex-col h-full gap-4 p-4">
      <div className="flex flex-wrap items-center gap-3">
        <TimeRangeSelector onChange={handleTimeRangeChange} className="w-auto" />

        <div className="flex items-center gap-2">
          <Input
            placeholder="Search logs..."
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
            className="w-48"
          />
          <Button variant="outline" size="icon" onClick={handleSearch}>
            <Search className="h-4 w-4" />
          </Button>
        </div>

        <Select
          value={selectedSeverities.length === 1 ? selectedSeverities[0] : ''}
          onValueChange={(value) => {
            if (value) {
              setSelectedSeverities([value])
            } else {
              setSelectedSeverities([])
            }
            setPage(1)
          }}
        >
          <SelectTrigger className="w-32">
            <SelectValue placeholder="Severity" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All</SelectItem>
            {SEVERITY_OPTIONS.map((sev) => (
              <SelectItem key={sev} value={sev}>
                {sev}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Button variant="outline" size="icon" onClick={() => refetch()}>
          <RefreshCw className="h-4 w-4" />
        </Button>
      </div>

      <ScrollArea className="flex-1 border rounded-md">
        {isLoading ? (
          <div className="flex items-center justify-center h-40">
            <Spinner className="w-6 h-6" />
          </div>
        ) : !data?.items?.length ? (
          <div className="flex flex-col items-center justify-center h-40 text-muted-foreground gap-2">
            <FileText className="w-8 h-8" />
            <span className="text-sm">No logs found</span>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-12"></TableHead>
                <TableHead className="w-48">Timestamp</TableHead>
                <TableHead className="w-24">Severity</TableHead>
                <TableHead>Message</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.items.map((log: LogEntry, index: number) => (
                <React.Fragment key={index}>
                  <TableRow className="cursor-pointer hover:bg-muted/50" onClick={() => toggleRowExpansion(index)}>
                    <TableCell>
                      {expandedRow === index ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
                    </TableCell>
                    <TableCell className="font-mono text-xs">{formatTimestamp(log.timestamp)}</TableCell>
                    <TableCell>
                      <SeverityBadge severity={log.severityText} />
                    </TableCell>
                    <TableCell className="max-w-md truncate font-mono text-xs">{log.body}</TableCell>
                  </TableRow>
                  {expandedRow === index && (
                    <TableRow>
                      <TableCell colSpan={4} className="bg-muted/30 p-4">
                        <div className="space-y-3">
                          <div>
                            <h4 className="text-sm font-medium mb-1">Full Message</h4>
                            <pre className="text-xs bg-background p-2 rounded overflow-x-auto whitespace-pre-wrap">
                              {log.body}
                            </pre>
                          </div>
                          {log.traceId && (
                            <div>
                              <h4 className="text-sm font-medium mb-1">Trace ID</h4>
                              <code className="text-xs bg-background p-1 rounded">{log.traceId}</code>
                            </div>
                          )}
                          {log.spanId && (
                            <div>
                              <h4 className="text-sm font-medium mb-1">Span ID</h4>
                              <code className="text-xs bg-background p-1 rounded">{log.spanId}</code>
                            </div>
                          )}
                          {Object.keys(log.logAttributes || {}).length > 0 && (
                            <div>
                              <h4 className="text-sm font-medium mb-1">Attributes</h4>
                              <pre className="text-xs bg-background p-2 rounded overflow-x-auto">
                                {JSON.stringify(log.logAttributes, null, 2)}
                              </pre>
                            </div>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  )}
                </React.Fragment>
              ))}
            </TableBody>
          </Table>
        )}
      </ScrollArea>

      {data && data.totalPages > 1 && (
        <div className="flex items-center justify-between">
          <span className="text-sm text-muted-foreground">
            Page {page} of {data.totalPages} ({data.total} total logs)
          </span>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
              <ChevronLeft className="h-4 w-4" />
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={page >= data.totalPages}
              onClick={() => setPage((p) => p + 1)}
            >
              Next
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
