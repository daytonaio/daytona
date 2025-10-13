/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState } from 'react'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { handleApiError } from '@/lib/error-handling'
import { toast } from 'sonner'

const Experimental: React.FC = () => {
  const { organizationsApi } = useApi()
  const [loadingCreate, setLoadingCreate] = useState(false)
  const { selectedOrganization } = useSelectedOrganization()
  const [headers, setHeaders] = useState<{ key: string; value: string }[]>(() => {
    const config = selectedOrganization?.experimentalConfig as Record<string, any> | undefined
    if (!config || typeof config !== 'object' || Array.isArray(config)) {
      return []
    }

    return config['otel']?.headers
      ? Object.entries(config['otel']?.headers).map(([key, value]) => ({ key, value: value as string }))
      : []
  })
  const [newHeader, setNewHeader] = useState<{ key: string; value: string }>({ key: '', value: '' })
  const [endpoint, setEndpoint] = useState(() => {
    const config = selectedOrganization?.experimentalConfig as Record<string, any> | undefined
    if (!config || typeof config !== 'object' || Array.isArray(config)) {
      return ''
    }

    return config['otel']?.endpoint || ''
  })
  const hasOtelEnabled =
    selectedOrganization?.experimentalConfig && !!(selectedOrganization.experimentalConfig as Record<string, any>).otel

  const handleSaveConfig = async () => {
    if (!selectedOrganization) {
      return
    }

    setLoadingCreate(true)
    try {
      const experimentalConfig = {
        ...selectedOrganization.experimentalConfig,
        otel: {
          endpoint,
          headers: [...headers, newHeader]
            .filter(({ key, value }) => key.trim() && value.trim())
            .reduce(
              (acc, { key, value }) => {
                acc[key] = value
                return acc
              },
              {} as { [key: string]: string },
            ),
        },
      }

      await organizationsApi.updateExperimentalConfig(selectedOrganization.id, experimentalConfig)
      toast.success('Experimental configuration saved successfully')
      setTimeout(() => window.location.reload(), 500)
    } catch (error) {
      handleApiError(error, 'Failed to save experimental configuration')
    } finally {
      setLoadingCreate(false)
    }
  }

  const handleDisableOtel = async () => {
    if (!selectedOrganization) {
      return
    }

    setLoadingCreate(true)
    try {
      const experimentalConfig: any = {
        ...selectedOrganization.experimentalConfig,
      }
      delete experimentalConfig['otel']

      await organizationsApi.updateExperimentalConfig(selectedOrganization.id, experimentalConfig)
      setEndpoint('')
      setHeaders([])
    } catch (error) {
      handleApiError(error, 'Failed to disable OpenTelemetry configuration')
    } finally {
      setLoadingCreate(false)
    }
  }

  if (!selectedOrganization) {
    return <div className="p-6">No organization selected.</div>
  }

  return (
    <div className="px-6 py-2">
      <div className="mb-2 h-12">
        <h1 className="text-2xl font-medium">Experimental Features</h1>

        <Card className="mt-4">
          <CardHeader>OpenTelemetry (OTEL) Configuration</CardHeader>
          <CardContent>
            <form
              id="experimental-features-form"
              className="space-y-6 overflow-y-auto px-1 pb-1"
              onSubmit={async (e) => {
                e.preventDefault()
                await handleSaveConfig()
              }}
            >
              <div className="space-y-3">
                <Label htmlFor="otel-endpoint">Endpoint</Label>
                <Input
                  id="otel-endpoint"
                  placeholder="https://otel-collector.example.com:4318"
                  value={endpoint}
                  onChange={(e) => setEndpoint(e.target.value)}
                />
                <p className="text-sm text-muted-foreground mt-1 pl-1">The OpenTelemetry collector endpoint URL.</p>
              </div>

              <div className="space-y-3 max-w-2xl">
                <div className="flex items-center justify-between">
                  <Label>Headers</Label>
                </div>

                <div className="space-y-2">
                  {headers.map(({ key, value }, index) => (
                    <div key={index} className="flex items-center gap-2">
                      <HeaderInput
                        headerKey={key}
                        headerValue={value}
                        onChangeKey={(e) => {
                          const newHeaders = [...headers]
                          newHeaders[index].key = e.target.value
                          setHeaders(newHeaders)
                        }}
                        onChangeValue={(e) => {
                          const newHeaders = [...headers]
                          newHeaders[index].value = e.target.value
                          setHeaders(newHeaders)
                        }}
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          const newHeaders = headers.filter((_, i) => i !== index)
                          setHeaders(newHeaders)
                        }}
                      >
                        Remove
                      </Button>
                    </div>
                  ))}

                  <div className="flex items-center gap-2">
                    <HeaderInput
                      headerKey={newHeader.key}
                      headerValue={newHeader.value}
                      onChangeKey={(e) => {
                        setNewHeader({ ...newHeader, key: e.target.value })
                      }}
                      onChangeValue={(e) => {
                        setNewHeader({ ...newHeader, value: e.target.value })
                      }}
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

                <p className="text-sm text-muted-foreground mt-1 pl-1">
                  Optional headers to send with OTEL requests (e.g., authentication tokens).
                </p>
              </div>

              <div className="flex items-center gap-2">
                <Button type="submit" form="experimental-features-form" disabled={loadingCreate || !endpoint.trim()}>
                  {loadingCreate ? 'Saving...' : 'Save'}
                </Button>
                {hasOtelEnabled && (
                  <Button type="button" variant="outline" onClick={handleDisableOtel}>
                    Disable
                  </Button>
                )}
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
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
          onAdd?.()
        }
      }}
    />
  </>
)

export default Experimental
