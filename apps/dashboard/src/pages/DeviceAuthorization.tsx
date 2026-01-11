/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useEffect, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Button } from '../components/ui/button'
import { Input } from '../components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'
import { Alert, AlertDescription } from '../components/ui/alert'
import { Loader2, CheckCircle2, XCircle } from 'lucide-react'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'

interface DeviceAuthInfo {
  client_id: string
  scope: string
  expires_at: string
}

const DeviceAuthorization = () => {
  const { user, isAuthenticated, isLoading } = useAuth()
  const { selectedOrganization } = useSelectedOrganization()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [userCode, setUserCode] = useState(searchParams.get('user_code') || '')
  const [authInfo, setAuthInfo] = useState<DeviceAuthInfo | null>(null)
  const [loading, setLoading] = useState(false)
  const [verifying, setVerifying] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  // Auto-verify if user code is provided in URL
  useEffect(() => {
    if (userCode && isAuthenticated && !authInfo && !verifying) {
      handleVerifyCode()
    }
  }, [userCode, isAuthenticated, authInfo])

  const handleVerifyCode = async () => {
    if (!userCode) {
      setError('Please enter a user code')
      return
    }

    setVerifying(true)
    setError(null)

    try {
      const response = await fetch(`/api/device/info?user_code=${userCode}`, {
        headers: {
          Authorization: `Bearer ${user?.access_token}`,
        },
      })

      const data = await response.json()

      if (data.error) {
        setError(data.message || 'Invalid or expired user code')
        setAuthInfo(null)
      } else {
        setAuthInfo(data)
      }
    } catch (err) {
      setError('Failed to verify code. Please try again.')
      setAuthInfo(null)
    } finally {
      setVerifying(false)
    }
  }

  const handleApprove = async () => {
    setLoading(true)
    setError(null)

    try {
      const response = await fetch('/api/device/approve', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${user?.access_token}`,
        },
        body: JSON.stringify({
          user_code: userCode,
          organization_id: selectedOrganization?.id || user?.profile?.sub,
        }),
      })

      const data = await response.json()

      if (response.ok && data.success) {
        setSuccess(true)
      } else {
        setError(data.message || 'Failed to approve authorization')
      }
    } catch (err) {
      setError('Failed to approve authorization. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  const handleDeny = async () => {
    setLoading(true)
    setError(null)

    try {
      const response = await fetch('/api/device/deny', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${user?.access_token}`,
        },
        body: JSON.stringify({
          user_code: userCode,
        }),
      })

      const data = await response.json()

      if (response.ok && data.success) {
        navigate('/')
      } else {
        setError(data.message || 'Failed to deny authorization')
      }
    } catch (err) {
      setError('Failed to deny authorization. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="w-16 h-16 animate-spin mx-auto mb-4 text-blue-600" />
          <p className="text-gray-600">Loading...</p>
        </div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Authentication Required</CardTitle>
            <CardDescription>Please log in to authorize device access</CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={() => navigate('/login')} className="w-full">
              Log In
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (success) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Card className="w-full max-w-md">
          <CardHeader>
            <div className="flex items-center justify-center mb-4">
              <CheckCircle2 className="w-16 h-16 text-green-500" />
            </div>
            <CardTitle className="text-center">Authorization Successful!</CardTitle>
            <CardDescription className="text-center">
              You can now close this window and return to your terminal
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={() => navigate('/')} variant="outline" className="w-full">
              Return to Dashboard
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Device Authorization</CardTitle>
          <CardDescription>Authorize Daytona CLI to access your account</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {!authInfo ? (
            <>
              <div className="space-y-2">
                <label htmlFor="user-code" className="text-sm font-medium">
                  Enter the code from your terminal
                </label>
                <Input
                  id="user-code"
                  type="text"
                  placeholder="XXXX-XXXX"
                  value={userCode}
                  onChange={(e) => setUserCode(e.target.value.toUpperCase())}
                  className="font-mono text-center text-lg"
                  maxLength={9}
                />
              </div>

              {error && (
                <Alert variant="destructive">
                  <XCircle className="h-4 w-4" />
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}

              <Button onClick={handleVerifyCode} disabled={verifying || !userCode} className="w-full">
                {verifying ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Verifying...
                  </>
                ) : (
                  'Continue'
                )}
              </Button>
            </>
          ) : (
            <>
              <div className="space-y-3 p-4 bg-gray-50 rounded-lg">
                <div>
                  <p className="text-sm text-gray-600">Client</p>
                  <p className="font-medium">{authInfo.client_id}</p>
                </div>
                {authInfo.scope && (
                  <div>
                    <p className="text-sm text-gray-600">Requested Access</p>
                    <p className="font-medium">{authInfo.scope}</p>
                  </div>
                )}
                <div>
                  <p className="text-sm text-gray-600">Expires</p>
                  <p className="font-medium">{new Date(authInfo.expires_at).toLocaleString()}</p>
                </div>
              </div>

              {error && (
                <Alert variant="destructive">
                  <XCircle className="h-4 w-4" />
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}

              <div className="flex gap-2">
                <Button onClick={handleApprove} disabled={loading} className="flex-1">
                  {loading ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Approving...
                    </>
                  ) : (
                    'Approve'
                  )}
                </Button>
                <Button onClick={handleDeny} disabled={loading} variant="outline" className="flex-1">
                  Deny
                </Button>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

export default DeviceAuthorization
