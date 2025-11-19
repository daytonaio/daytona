/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationEmail } from '@/billing-api'
import { AutomaticTopUp } from '@/billing-api/types/OrganizationWallet'
import { OrganizationEmailsTable } from '@/components/OrganizationEmails'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { InputGroup, InputGroupAddon, InputGroupInput, InputGroupText } from '@/components/ui/input-group'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { useApi } from '@/hooks/useApi'
import { useBilling } from '@/hooks/useBilling'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { CheckCircleIcon, CreditCardIcon, InfoIcon, Loader, TriangleAlertIcon } from 'lucide-react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { NumericFormat } from 'react-number-format'
import { useAuth } from 'react-oidc-context'
import { toast } from 'sonner'

const formatAmount = (amount: number) => {
  return Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount / 100)
}

const Wallet = () => {
  const { selectedOrganization } = useSelectedOrganization()
  const { billingApi } = useApi()
  const { user } = useAuth()
  const { wallet, walletLoading, billingPortalUrl, billingPortalUrlLoading, refreshWallet } = useBilling()
  const [automaticTopUp, setAutomaticTopUp] = useState<AutomaticTopUp | undefined>(undefined)
  const [automaticTopUpLoading, setAutomaticTopUpLoading] = useState(false)
  const [redeemCouponLoading, setRedeemCouponLoading] = useState(false)
  const [couponCode, setCouponCode] = useState<string>('')
  const [redeemCouponError, setRedeemCouponError] = useState<string | null>(null)
  const [redeemCouponSuccess, setRedeemCouponSuccess] = useState<string | null>(null)
  const [organizationEmails, setOrganizationEmails] = useState<OrganizationEmail[]>([])
  const [organizationEmailsLoading, setOrganizationEmailsLoading] = useState(true)

  useEffect(() => {
    if (wallet?.automaticTopUp) {
      setAutomaticTopUp(wallet.automaticTopUp)
    }
  }, [wallet])

  const handleUpdatePaymentMethod = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    try {
      const data = await billingApi.getOrganizationCheckoutUrl(selectedOrganization.id)
      window.open(data, '_blank')
    } catch (error) {
      console.error('Failed to fetch checkout url:', error)
    }
  }, [billingApi, selectedOrganization])

  const handleSetAutomaticTopUp = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }

    setAutomaticTopUpLoading(true)
    try {
      await billingApi.setAutomaticTopUp(selectedOrganization.id, automaticTopUp)
      toast.success('Automatic top up set successfully')
      refreshWallet()
    } catch (error) {
      console.error('Failed to set automatic top up:', error)
      toast.error('Failed to set automatic top up', {
        description: String(error),
      })
    } finally {
      setAutomaticTopUpLoading(false)
    }
  }, [billingApi, selectedOrganization, automaticTopUp, refreshWallet])

  const handleRedeemCoupon = useCallback(async () => {
    if (!selectedOrganization || !couponCode) {
      return
    }

    if (redeemCouponLoading) {
      return
    }

    setRedeemCouponLoading(true)
    setRedeemCouponError(null)
    setRedeemCouponSuccess(null)
    try {
      setRedeemCouponSuccess(await billingApi.redeemCoupon(selectedOrganization.id, couponCode))
      setTimeout(() => {
        setRedeemCouponSuccess(null)
      }, 3000)
      setCouponCode('')
      refreshWallet()
    } catch (error) {
      console.error('Failed to redeem coupon:', error)
      setRedeemCouponError(String(error))
    } finally {
      setRedeemCouponLoading(false)
    }
  }, [billingApi, selectedOrganization, couponCode, refreshWallet, redeemCouponLoading])

  const saveAutomaticTopUpDisabled = useMemo(() => {
    if (automaticTopUpLoading) {
      return true
    }

    if (automaticTopUp?.thresholdAmount !== wallet?.automaticTopUp?.thresholdAmount) {
      if (!wallet?.automaticTopUp) {
        if ((automaticTopUp?.thresholdAmount || 0) !== 0) {
          return false
        }
      } else {
        return false
      }
    }

    if (automaticTopUp?.targetAmount !== wallet?.automaticTopUp?.targetAmount) {
      if (!wallet?.automaticTopUp) {
        if ((automaticTopUp?.targetAmount || 0) !== 0) {
          return false
        }
      } else {
        return false
      }
    }

    return true
  }, [automaticTopUpLoading, wallet, automaticTopUp])

  const fetchOrganizationEmails = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    setOrganizationEmailsLoading(true)
    try {
      const data = await billingApi.listOrganizationEmails(selectedOrganization.id)
      setOrganizationEmails(data)
    } catch (error) {
      console.error('Failed to fetch organization emails:', error)
    } finally {
      setOrganizationEmailsLoading(false)
    }
  }, [billingApi, selectedOrganization])

  const handleDeleteEmail = useCallback(
    async (email: string) => {
      if (!selectedOrganization) {
        return
      }
      try {
        await billingApi.deleteOrganizationEmail(selectedOrganization.id, email)
        toast.success('Email deleted successfully')
        fetchOrganizationEmails()
      } catch (error) {
        console.error('Failed to delete email:', error)
        toast.error('Failed to delete email', {
          description: String(error),
        })
      }
    },
    [billingApi, selectedOrganization, fetchOrganizationEmails],
  )

  const handleResendVerification = useCallback(
    async (email: string) => {
      if (!selectedOrganization) {
        return
      }
      try {
        await billingApi.resendOrganizationEmailVerification(selectedOrganization.id, email)
        toast.success('Verification email sent successfully')
      } catch (error) {
        console.error('Failed to resend verification email:', error)
        toast.error('Failed to resend verification email', {
          description: String(error),
        })
      }
    },
    [billingApi, selectedOrganization],
  )

  const handleAddEmail = useCallback(
    async (email: string) => {
      if (!selectedOrganization) {
        return
      }
      try {
        await billingApi.addOrganizationEmail(selectedOrganization.id, email)
        toast.success('Email added successfully. A verification email has been sent.')
        fetchOrganizationEmails()
      } catch (error) {
        console.error('Failed to add email:', error)
        toast.error('Failed to add email', {
          description: String(error),
        })
      }
    },
    [billingApi, selectedOrganization, fetchOrganizationEmails],
  )

  useEffect(() => {
    fetchOrganizationEmails()
  }, [fetchOrganizationEmails])

  return (
    <div className="p-6 max-w-3xl">
      <h1 className="text-2xl font-bold mb-3">Wallet</h1>
      {!user?.profile.email_verified && (
        <Alert variant="destructive">
          <TriangleAlertIcon />
          <AlertTitle>Verify your email</AlertTitle>
          <AlertDescription>
            {wallet?.balanceCents && wallet.balanceCents > 0 ? (
              <>
                Please verify your email address to complete your account setup.
                <br />A verification email was sent to you.
              </>
            ) : (
              <>
                Verify your email address to recieve $100 of credits.
                <br />A verification email was sent to you.
              </>
            )}
          </AlertDescription>
        </Alert>
      )}

      {!wallet?.creditCardConnected && user?.profile.email_verified && (
        <Alert variant="warning">
          <CreditCardIcon />
          <AlertTitle> Credit card not connected</AlertTitle>
          <AlertDescription>
            {selectedOrganization?.personal ? (
              <>Connect a credit card to receive an additional $100 of credits.</>
            ) : (
              <>Please connect your credit card to your account to continue using our service.</>
            )}
          </AlertDescription>
        </Alert>
      )}

      <div className="flex flex-col gap-6 mt-4">
        <Card className="h-full">
          <CardHeader>
            <CardTitle>Overview</CardTitle>
          </CardHeader>
          <CardContent className="">
            {walletLoading || !wallet ? (
              <Skeleton className="max-w-sm h-10" />
            ) : (
              <div className="flex items-start sm:flex-row flex-col gap-4 sm:items-center justify-between">
                <div className="flex gap-4 sm:gap-12 items-end sm:flex-row flex-col">
                  <div className="flex flex-col gap-1">
                    <div className="">Current balance</div>
                    <div className="text-xl text-foreground font-semibold">
                      {formatAmount(wallet.ongoingBalanceCents)}
                    </div>
                  </div>
                  <div className="flex flex-col gap-1">
                    <div className="">Spent this month</div>
                    <div className="text-xl font-semibold">
                      {formatAmount(wallet.balanceCents - wallet.ongoingBalanceCents)}
                    </div>
                  </div>
                </div>
                {wallet.creditCardConnected && billingPortalUrl && (
                  <Button variant="default" className="">
                    Top-up
                  </Button>
                )}
              </div>
            )}
          </CardContent>
          <CardContent className="border-t border-border">
            {walletLoading || !wallet || billingPortalUrlLoading || !billingPortalUrl ? (
              <Skeleton className="max-w-sm h-10" />
            ) : (
              <div className="flex gap-4 items-center justify-between">
                <div className="flex flex-col gap-1 items-start">
                  <div className="text-sm font-medium">Payment method</div>
                  {!wallet.creditCardConnected ? (
                    <div className="text-sm text-muted-foreground">Payment method not connected</div>
                  ) : (
                    <div className="text-sm text-muted-foreground flex items-center gap-2">
                      <CheckCircleIcon className="w-4 h-4 shrink-0" /> Credit card connected
                    </div>
                  )}
                </div>
                {!wallet.creditCardConnected ? (
                  <Button variant="default" onClick={handleUpdatePaymentMethod}>
                    Connect
                  </Button>
                ) : (
                  <Button variant="secondary" onClick={handleUpdatePaymentMethod}>
                    Update
                  </Button>
                )}
              </div>
            )}
          </CardContent>

          {user?.profile.email_verified && !!wallet && (
            <CardContent className="border-t border-border">
              <div className="flex gap-4 md:items-center justify-between md:flex-row flex-col">
                <div className="flex flex-col gap-1 items-start flex-1">
                  <div className="text-sm font-medium">Redeem coupon</div>
                  {redeemCouponError ? (
                    <div className="text-sm text-destructive">{redeemCouponError}</div>
                  ) : redeemCouponSuccess ? (
                    <div className="text-sm text-success">{redeemCouponSuccess}</div>
                  ) : (
                    <div className="text-sm text-muted-foreground">Enter a coupon code to redeem your credits.</div>
                  )}
                </div>

                <div className="flex gap-2 items-center">
                  <Input
                    placeholder="Enter coupon code"
                    value={couponCode}
                    onChange={(e) => setCouponCode(e.target.value)}
                  />
                  <Button variant="secondary" className="min-w-[4.5rem]" onClick={handleRedeemCoupon}>
                    Redeem
                  </Button>
                </div>
              </div>
            </CardContent>
          )}
        </Card>
        {wallet?.creditCardConnected && (
          <Card className="w-full">
            <CardHeader>
              <CardTitle>Automatic top-up</CardTitle>
              <CardDescription>
                Set automatic top-up rules for your wallet.
                <br />
                The target amount must be at least $10 higher than the threshold amount.
              </CardDescription>
            </CardHeader>
            <CardContent>
              {walletLoading || !wallet ? (
                <Skeleton className="max-w-sm h-10" />
              ) : (
                <div className="flex sm:flex-row flex-col gap-6">
                  <div className="flex flex-col gap-2 flex-1">
                    <Label htmlFor="thresholdAmount">When balance is below</Label>
                    <InputGroup>
                      <InputGroupAddon>
                        <InputGroupText>$</InputGroupText>
                      </InputGroupAddon>
                      <NumericFormat
                        customInput={InputGroupInput}
                        placeholder="0.00"
                        id="thresholdAmount"
                        inputMode="decimal"
                        thousandSeparator
                        decimalScale={2}
                        value={automaticTopUp?.thresholdAmount ?? ''}
                        onValueChange={({ floatValue }) => {
                          const value = floatValue ?? 0

                          let targetAmount = automaticTopUp?.targetAmount ?? 0
                          if (value > targetAmount - 10) {
                            targetAmount = value + 10
                          }

                          setAutomaticTopUp({
                            thresholdAmount: value,
                            targetAmount,
                          })
                        }}
                      />
                      <InputGroupAddon align="inline-end">
                        <InputGroupText>USD</InputGroupText>
                      </InputGroupAddon>
                    </InputGroup>
                  </div>

                  <div className="flex flex-col gap-2 flex-1">
                    <Label htmlFor="targetAmount">Bring balance to</Label>
                    <InputGroup>
                      <InputGroupAddon>
                        <InputGroupText>$</InputGroupText>
                      </InputGroupAddon>
                      <NumericFormat
                        placeholder="0.00"
                        customInput={InputGroupInput}
                        id="targetAmount"
                        inputMode="decimal"
                        thousandSeparator
                        decimalScale={2}
                        value={automaticTopUp?.targetAmount ?? ''}
                        onValueChange={({ floatValue }) => {
                          const thresholdAmount = automaticTopUp?.thresholdAmount ?? 0
                          setAutomaticTopUp({
                            thresholdAmount,
                            targetAmount: floatValue ?? 0,
                          })
                        }}
                        onBlur={() => {
                          const thresholdAmount = automaticTopUp?.thresholdAmount ?? 0
                          const currentTarget = automaticTopUp?.targetAmount ?? 0

                          if (currentTarget < thresholdAmount) {
                            setAutomaticTopUp({
                              thresholdAmount,
                              targetAmount: thresholdAmount,
                            })
                          }
                        }}
                      />
                      <InputGroupAddon align="inline-end">
                        <InputGroupText>USD</InputGroupText>
                      </InputGroupAddon>
                    </InputGroup>
                  </div>
                </div>
              )}
            </CardContent>
            <CardFooter className="flex justify-between gap-2">
              <div className="flex items-center gap-2 text-muted-foreground">
                <InfoIcon className="w-4 h-4 shrink-0" />{' '}
                <span className="text-sm ">Setting both values to 0 will disable automatic top-ups.</span>
              </div>
              <div className="flex gap-2 items-center ml-auto">
                <Button
                  onClick={handleSetAutomaticTopUp}
                  disabled={saveAutomaticTopUpDisabled || walletLoading || !wallet}
                  className="min-w-[4.5rem]"
                >
                  {automaticTopUpLoading ? <Loader className="w-4 h-4 animate-spin" /> : 'Save'}
                </Button>
              </div>
            </CardFooter>
          </Card>
        )}
      </div>

      {/* Organization Emails Section */}
      <div className="mt-8">
        <Card>
          <CardHeader>
            <CardTitle>Billing emails</CardTitle>
            <CardDescription>
              Manage billing emails for your organization which recieve important billing notifications such as invoices
              and credit depletion notices.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <OrganizationEmailsTable
              data={organizationEmails}
              loading={organizationEmailsLoading}
              handleDelete={handleDeleteEmail}
              handleResendVerification={handleResendVerification}
              handleAddEmail={handleAddEmail}
            />
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

export default Wallet
