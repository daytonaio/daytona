/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useApi } from '@/hooks/useApi'
import { useOrganizations } from '@/hooks/useOrganizations'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import type { Organization } from '@daytona/api-client'
import React, { useState } from 'react'
import { toast } from 'sonner'

type HeaderEntry = { key: string; value: string }

const headersFromOrganization = (organization: Organization | null | undefined): HeaderEntry[] => {
  const headers = organization?.otelConfig?.headers
  if (!headers) {
    return []
  }
  return Object.entries(headers).map(([key, value]) => ({ key, value: value as string }))
}

const endpointFromOrganization = (organization: Organization | null | undefined): string =>
  organization?.otelConfig?.endpoint ?? ''

export const OtelConfigCard: React.FC = () => {
  const { organizationsApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()
  const { refreshOrganizations } = useOrganizations()

  const [endpoint, setEndpoint] = useState(() => endpointFromOrganization(selectedOrganization))
  const [headers, setHeaders] = useState<HeaderEntry[]>(() => headersFromOrganization(selectedOrganization))
  const [newHeader, setNewHeader] = useState<HeaderEntry>({ key: '', value: '' })
  const [saving, setSaving] = useState(false)

  const hasOtelEnabled = !!selectedOrganization?.otelConfig

  const handleSave = async () => {
    if (!selectedOrganization) {
      return
    }

    setSaving(true)
    try {
      const allHeaders = [...headers, newHeader]
        .filter(({ key, value }) => key.trim() && value.trim())
        .reduce(
          (acc, { key, value }) => {
            acc[key] = value
            return acc
          },
          {} as Record<string, string>,
        )

      await organizationsApi.updateOrganizationOtelConfig(selectedOrganization.id, {
        endpoint,
        headers: allHeaders,
      })
      await refreshOrganizations(selectedOrganization.id)
      setNewHeader({ key: '', value: '' })
      toast.success('OpenTelemetry configuration saved')
    } catch (error) {
      handleApiError(error, 'Failed to save OpenTelemetry configuration')
    } finally {
      setSaving(false)
    }
  }

  const handleDisable = async () => {
    if (!selectedOrganization) {
      return
    }

    setSaving(true)
    try {
      await organizationsApi.deleteOrganizationOtelConfig(selectedOrganization.id)
      await refreshOrganizations(selectedOrganization.id)
      setEndpoint('')
      setHeaders([])
      setNewHeader({ key: '', value: '' })
      toast.success('OpenTelemetry configuration disabled')
    } catch (error) {
      handleApiError(error, 'Failed to disable OpenTelemetry configuration')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Card>
      <CardHeader className="p-4">
        <CardTitle>OpenTelemetry</CardTitle>
      </CardHeader>
      <CardContent className="border-t border-border">
        <form
          id="otel-config-form"
          className="space-y-6"
          onSubmit={async (e) => {
            e.preventDefault()
            await handleSave()
          }}
        >
          <div className="space-y-2">
            <Label htmlFor="otel-endpoint">OTLP Endpoint</Label>
            <Input
              id="otel-endpoint"
              placeholder="https://otel-collector.example.com:4318"
              value={endpoint}
              onChange={(e) => setEndpoint(e.target.value)}
            />
            <p className="text-sm text-muted-foreground">The OpenTelemetry collector endpoint URL.</p>
          </div>

          <div className="space-y-2">
            <Label>Headers</Label>
            <div className="space-y-2">
              {headers.map(({ key, value }, index) => (
                <div key={index} className="flex items-center gap-2">
                  <HeaderInput
                    headerKey={key}
                    headerValue={value}
                    onChangeKey={(e) => {
                      const next = [...headers]
                      next[index].key = e.target.value
                      setHeaders(next)
                    }}
                    onChangeValue={(e) => {
                      const next = [...headers]
                      next[index].value = e.target.value
                      setHeaders(next)
                    }}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => setHeaders(headers.filter((_, i) => i !== index))}
                  >
                    Remove
                  </Button>
                </div>
              ))}

              <div className="flex items-center gap-2">
                <HeaderInput
                  headerKey={newHeader.key}
                  headerValue={newHeader.value}
                  onChangeKey={(e) => setNewHeader({ ...newHeader, key: e.target.value })}
                  onChangeValue={(e) => setNewHeader({ ...newHeader, value: e.target.value })}
                  onAdd={() => {
                    if (newHeader.key.trim() && newHeader.value.trim()) {
                      setHeaders([...headers, newHeader])
                      setNewHeader({ key: '', value: '' })
                    }
                  }}
                />
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    if (newHeader.key.trim() && newHeader.value.trim()) {
                      setHeaders([...headers, newHeader])
                      setNewHeader({ key: '', value: '' })
                    }
                  }}
                  disabled={!newHeader.key.trim() || !newHeader.value.trim()}
                >
                  +
                </Button>
              </div>
            </div>
            <p className="text-sm text-muted-foreground">
              Optional headers to send with OTLP requests (e.g., authentication tokens). Existing values are stored
              encrypted and shown as <code>******</code>.
            </p>
          </div>

          <div className="flex items-center gap-2">
            <Button type="submit" form="otel-config-form" disabled={saving || !endpoint.trim()}>
              {saving ? 'Saving...' : 'Save'}
            </Button>
            {hasOtelEnabled && (
              <Button type="button" variant="outline" onClick={handleDisable} disabled={saving}>
                Disable
              </Button>
            )}
          </div>
        </form>
      </CardContent>
    </Card>
  )
}

const HeaderInput = ({
  headerKey,
  headerValue,
  onChangeKey,
  onChangeValue,
  onAdd,
}: {
  headerKey: string
  headerValue: string
  onChangeKey: (e: React.ChangeEvent<HTMLInputElement>) => void
  onChangeValue: (e: React.ChangeEvent<HTMLInputElement>) => void
  onAdd?: () => void
}) => (
  <>
    <Input placeholder="Header key" className="flex-1" value={headerKey} onChange={onChangeKey} />
    <Input
      placeholder="Header value"
      className="flex-1"
      value={headerValue}
      onChange={onChangeValue}
      onKeyDown={(e) => {
        if (e.key === 'Enter' && headerKey.trim() && headerValue.trim()) {
          e.preventDefault()
          onAdd?.()
        }
      }}
    />
  </>
)
