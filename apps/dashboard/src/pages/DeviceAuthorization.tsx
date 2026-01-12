/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useCallback, useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useAuth } from 'react-oidc-context'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Loader2, CheckCircle2, XCircle, Terminal, Shield } from 'lucide-react'
import { useConfig } from '@/hooks/useConfig'
import { toast } from 'sonner'
import LoadingFallback from '@/components/LoadingFallback'

interface DeviceStatus {
  user_code: string
  client_id: string
  scope: string
  status: string
  expires_in: number
}

interface Organization {
  id: string
  name: string
  personal: boolean
}

const DeviceAuthorization: React.FC = () => {
  const [searchParams] = useSearchParams()
  const { apiUrl } = useConfig()
  const { isAuthenticated, isLoading: authLoading, user, signinRedirect } = useAuth()

  const [userCode, setUserCode] = useState(searchParams.get('user_code') || '')
  const [deviceStatus, setDeviceStatus] = useState<DeviceStatus | null>(null)
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [result, setResult] = useState<'approved' | 'denied' | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [organizations, setOrganizations] = useState<Organization[]>([])
  const [selectedOrgId, setSelectedOrgId] = useState<string>('')

  const fetchDeviceStatus = useCallback(
    async (code: string) => {
      if (!code.trim()) return

      setLoading(true)
      setError(null)
      setDeviceStatus(null)

      try {
        const response = await fetch(`${apiUrl}/auth/device/status?user_code=${encodeURIComponent(code.trim())}`)
        if (!response.ok) {
          const data = await response.json()
          throw new Error(data.message || 'Invalid or expired code')
        }
        const data: DeviceStatus = await response.json()
        setDeviceStatus(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to verify code')
      } finally {
        setLoading(false)
      }
    },
    [apiUrl],
  )

  const fetchOrganizations = useCallback(async () => {
    if (!user?.access_token) return

    try {
      const response = await fetch(`${apiUrl}/organizations`, {
        headers: {
          Authorization: `Bearer ${user.access_token}`,
        },
      })
      if (response.ok) {
        const data: Organization[] = await response.json()
        setOrganizations(data)
        // Default to personal organization or first one
        const personalOrg = data.find((org) => org.personal)
        setSelectedOrgId(personalOrg?.id || data[0]?.id || '')
      }
    } catch {
      // Silently fail, user can still proceed
    }
  }, [apiUrl, user?.access_token])

  useEffect(() => {
    // Auto-fetch status if user_code is in URL
    if (searchParams.get('user_code')) {
      fetchDeviceStatus(searchParams.get('user_code')!)
    }
  }, [searchParams, fetchDeviceStatus])

  useEffect(() => {
    if (isAuthenticated && user) {
      fetchOrganizations()
    }
  }, [isAuthenticated, user, fetchOrganizations])

  const handleSubmitCode = (e: React.FormEvent) => {
    e.preventDefault()
    fetchDeviceStatus(userCode)
  }

  const handleApprove = async () => {
    if (!deviceStatus || !user?.access_token || !selectedOrgId) return

    setSubmitting(true)
    setError(null)

    try {
      const response = await fetch(`${apiUrl}/auth/device/approve`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${user.access_token}`,
          'X-Daytona-Organization-ID': selectedOrgId,
        },
        body: JSON.stringify({
          user_code: deviceStatus.user_code,
          action: 'approve',
          organization_id: selectedOrgId,
        }),
      })

      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.message || 'Failed to approve authorization')
      }

      setResult('approved')
      toast.success('Device authorization approved!')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to approve authorization')
    } finally {
      setSubmitting(false)
    }
  }

  const handleDeny = async () => {
    if (!deviceStatus || !user?.access_token) return

    setSubmitting(true)
    setError(null)

    try {
      const response = await fetch(`${apiUrl}/auth/device/approve`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${user.access_token}`,
          'X-Daytona-Organization-ID': selectedOrgId,
        },
        body: JSON.stringify({
          user_code: deviceStatus.user_code,
          action: 'deny',
        }),
      })

      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.message || 'Failed to deny authorization')
      }

      setResult('denied')
      toast.info('Device authorization denied')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to deny authorization')
    } finally {
      setSubmitting(false)
    }
  }

  const handleLogin = () => {
    signinRedirect({
      state: {
        returnTo: `/device${userCode ? `?user_code=${userCode}` : ''}`,
      },
    })
  }

  if (authLoading) {
    return <LoadingFallback />
  }

  // Show success/denied result
  if (result) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background p-4">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            {result === 'approved' ? (
              <>
                <div className="mx-auto mb-4 w-12 h-12 rounded-full bg-green-100 dark:bg-green-900 flex items-center justify-center">
                  <CheckCircle2 className="h-6 w-6 text-green-600 dark:text-green-400" />
                </div>
                <CardTitle>Authorization Approved</CardTitle>
                <CardDescription>You can now close this window and return to your terminal.</CardDescription>
              </>
            ) : (
              <>
                <div className="mx-auto mb-4 w-12 h-12 rounded-full bg-red-100 dark:bg-red-900 flex items-center justify-center">
                  <XCircle className="h-6 w-6 text-red-600 dark:text-red-400" />
                </div>
                <CardTitle>Authorization Denied</CardTitle>
                <CardDescription>The CLI authentication request has been denied.</CardDescription>
              </>
            )}
          </CardHeader>
        </Card>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center">
            <Terminal className="h-6 w-6 text-primary" />
          </div>
          <CardTitle>CLI Device Authorization</CardTitle>
          <CardDescription>Authorize Daytona CLI to access your account</CardDescription>
        </CardHeader>

        <CardContent className="space-y-4">
          {error && (
            <Alert variant="destructive">
              <XCircle className="h-4 w-4" />
              <AlertTitle>Error</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {!deviceStatus ? (
            <form onSubmit={handleSubmitCode} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="userCode">Enter the code from your terminal</Label>
                <Input
                  id="userCode"
                  placeholder="XXXX-XXXX"
                  value={userCode}
                  onChange={(e) => setUserCode(e.target.value.toUpperCase())}
                  className="text-center text-lg font-mono tracking-widest"
                  maxLength={9}
                  autoComplete="off"
                  autoFocus
                />
              </div>
              <Button type="submit" className="w-full" disabled={loading || !userCode.trim()}>
                {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Verify Code
              </Button>
            </form>
          ) : (
            <div className="space-y-4">
              <Alert>
                <Shield className="h-4 w-4" />
                <AlertTitle>Authorization Request</AlertTitle>
                <AlertDescription>
                  <strong>{deviceStatus.client_id}</strong> is requesting access to your Daytona account.
                </AlertDescription>
              </Alert>

              <div className="rounded-lg bg-muted p-4 space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Code:</span>
                  <span className="font-mono font-medium">{deviceStatus.user_code}</span>
                </div>
                {deviceStatus.scope && (
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Permissions:</span>
                    <span className="font-medium">Full access</span>
                  </div>
                )}
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Expires in:</span>
                  <span className="font-medium">{Math.floor(deviceStatus.expires_in / 60)} minutes</span>
                </div>
              </div>

              {!isAuthenticated ? (
                <div className="space-y-3">
                  <p className="text-sm text-muted-foreground text-center">
                    Please sign in to approve or deny this request.
                  </p>
                  <Button onClick={handleLogin} className="w-full">
                    Sign In to Continue
                  </Button>
                </div>
              ) : (
                <>
                  {organizations.length > 1 && (
                    <div className="space-y-2">
                      <Label htmlFor="organization">Organization</Label>
                      <Select value={selectedOrgId} onValueChange={setSelectedOrgId}>
                        <SelectTrigger id="organization">
                          <SelectValue placeholder="Select organization" />
                        </SelectTrigger>
                        <SelectContent>
                          {organizations.map((org) => (
                            <SelectItem key={org.id} value={org.id}>
                              {org.name} {org.personal && '(Personal)'}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}
                </>
              )}
            </div>
          )}
        </CardContent>

        {deviceStatus && isAuthenticated && (
          <CardFooter className="flex gap-3">
            <Button variant="outline" className="flex-1" onClick={handleDeny} disabled={submitting}>
              {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Deny
            </Button>
            <Button className="flex-1" onClick={handleApprove} disabled={submitting || !selectedOrgId}>
              {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Approve
            </Button>
          </CardFooter>
        )}
      </Card>
    </div>
  )
}

export default DeviceAuthorization
