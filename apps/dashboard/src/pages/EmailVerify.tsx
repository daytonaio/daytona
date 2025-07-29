/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Logo } from '@/assets/Logo'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle } from '@/components/ui/card'
import { RoutePath } from '@/enums/RoutePath'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

export default function EmailVerify() {
  const { organizationId, email, token } = useParams<{
    organizationId: string
    email: string
    token: string
  }>()
  const navigate = useNavigate()
  const [verificationStatus, setVerificationStatus] = useState<'loading' | 'success' | 'error'>('loading')
  const [errorMessage, setErrorMessage] = useState<string>('')
  const { onSelectOrganization } = useSelectedOrganization()
  const { billingApi } = useApi()

  useEffect(() => {
    const verifyEmail = async () => {
      if (!organizationId || !email || !token) {
        setVerificationStatus('error')
        setErrorMessage('Invalid verification link')
        return
      }

      try {
        await billingApi.verifyOrganizationEmail(organizationId, email, token)
        setVerificationStatus('success')
        onSelectOrganization(organizationId)
        setTimeout(() => {
          navigate(RoutePath.BILLING_WALLET)
        }, 1000)
      } catch (error) {
        setVerificationStatus('error')
        setErrorMessage('An error occurred while verifying your email')
      }
    }

    verifyEmail()
  }, [organizationId, email, token, billingApi, navigate, onSelectOrganization])

  return (
    <div className="flex items-center justify-center min-h-screen bg-background">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-4">
            <Logo />
          </div>
          {verificationStatus === 'loading' && (
            <>
              <CardTitle>Verifying Your Email</CardTitle>
              <p className="text-muted-foreground">Please wait while we verify your email address...</p>
            </>
          )}
          {verificationStatus === 'success' && (
            <>
              <CardTitle className="text-green-600">Email Verified Successfully!</CardTitle>
              <p className="text-muted-foreground">
                Your email has been verified. You will be redirected to the wallet page shortly.
              </p>
            </>
          )}
          {verificationStatus === 'error' && (
            <>
              <CardTitle className="text-red-600">Verification Failed</CardTitle>
              <p className="text-muted-foreground">{errorMessage}</p>
              <Button onClick={() => navigate(RoutePath.BILLING_WALLET)} className="mt-4">
                Go to Wallet
              </Button>
            </>
          )}
        </CardHeader>
      </Card>
    </div>
  )
}
