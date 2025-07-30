/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AutomaticTopUp, OrganizationWallet } from '@/billing-api/types/OrganizationWallet'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useEffect, useMemo, useState } from 'react'
import { useCallback } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { ArrowUpRight, CreditCard, Info, Loader2 } from 'lucide-react'
import { Slider } from '@/components/ui/slider'
import { toast } from 'sonner'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Tooltip } from '@/components/Tooltip'
import { useApi } from '@/hooks/useApi'
import { useAuth } from 'react-oidc-context'
import { OrganizationEmail } from '@/billing-api'
import { OrganizationEmailsTable } from '@/components/OrganizationEmails'
import EmailVerify from './EmailVerify'

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
  const [wallet, setWallet] = useState<OrganizationWallet | null>(null)
  const [walletLoading, setWalletLoading] = useState(true)
  const [billingPortalUrl, setBillingPortalUrl] = useState<string | null>(null)
  const [billingPortalUrlLoading, setBillingPortalUrlLoading] = useState(true)
  const [automaticTopUp, setAutomaticTopUp] = useState<AutomaticTopUp | undefined>(undefined)
  const [automaticTopUpLoading, setAutomaticTopUpLoading] = useState(false)
  const [redeemCouponLoading, setRedeemCouponLoading] = useState(false)
  const [couponCode, setCouponCode] = useState<string>('')
  const [redeemCouponError, setRedeemCouponError] = useState<string | null>(null)
  const [redeemCouponSuccess, setRedeemCouponSuccess] = useState<string | null>(null)
  const [organizationEmails, setOrganizationEmails] = useState<OrganizationEmail[]>([])
  const [organizationEmailsLoading, setOrganizationEmailsLoading] = useState(true)

  const fetchWallet = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    setWalletLoading(true)
    try {
      const data = await billingApi.getOrganizationWallet(selectedOrganization.id)
      setWallet(data)
      setAutomaticTopUp(data.automaticTopUp)
    } catch (error) {
      console.error('Failed to fetch wallet data:', error)
    } finally {
      setWalletLoading(false)
    }
  }, [billingApi, selectedOrganization])

  const fetchBillingPortalUrl = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    setBillingPortalUrlLoading(true)
    try {
      const data = await billingApi.getOrganizationBillingPortalUrl(selectedOrganization.id)
      setBillingPortalUrl(data)
    } catch (error) {
      console.error('Failed to fetch billing portal url:', error)
    } finally {
      setBillingPortalUrlLoading(false)
    }
  }, [billingApi, selectedOrganization])

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
      fetchWallet()
    } catch (error) {
      console.error('Failed to set automatic top up:', error)
      toast.error('Failed to set automatic top up', {
        description: String(error),
      })
    } finally {
      setAutomaticTopUpLoading(false)
    }
  }, [billingApi, selectedOrganization, automaticTopUp, fetchWallet])

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
      await billingApi.redeemCoupon(selectedOrganization.id, couponCode)
      setRedeemCouponSuccess('Coupon redeemed successfully')
      setTimeout(() => {
        setRedeemCouponSuccess(null)
      }, 3000)
      setCouponCode('')
      fetchWallet()
    } catch (error) {
      console.error('Failed to redeem coupon:', error)
      setRedeemCouponError(String(error))
    } finally {
      setRedeemCouponLoading(false)
    }
  }, [billingApi, selectedOrganization, couponCode, fetchWallet, redeemCouponLoading])

  const saveAutomaticTopUpDisabled = useMemo(() => {
    if (automaticTopUpLoading) {
      return true
    }

    if (automaticTopUp?.thresholdAmount !== wallet?.automaticTopUp?.thresholdAmount) {
      if (!wallet?.automaticTopUp) {
        if (automaticTopUp?.thresholdAmount !== 0) {
          return false
        }
      } else {
        return false
      }
    }

    if (automaticTopUp?.targetAmount !== wallet?.automaticTopUp?.targetAmount) {
      if (!wallet?.automaticTopUp) {
        if (automaticTopUp?.targetAmount !== 0) {
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

  useEffect(() => {
    fetchWallet()
  }, [fetchWallet])

  useEffect(() => {
    fetchBillingPortalUrl()
  }, [fetchBillingPortalUrl])

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold">Wallet</h1>
      <div className="flex gap-4">
        <Card className="my-4 h-full">
          <CardHeader>
            <CardTitle>Wallet</CardTitle>
          </CardHeader>
          <CardContent className="p-6 w-full h-full">
            {walletLoading || !wallet ? (
              <Skeleton className="max-w-sm h-10" />
            ) : (
              <CardDescription className="flex flex-col justify-between h-full">
                <div>
                  <div className="text-2xl font-bold">Balance: {formatAmount(wallet.ongoingBalanceCents)}</div>
                  <div className="text-xl font-bold my-2">
                    Spent this month: {formatAmount(wallet.balanceCents - wallet.ongoingBalanceCents)}
                  </div>
                </div>
                {!user?.profile.email_verified && (
                  <div className="text-sm text-red-500 max-w-sm">
                    {wallet?.balanceCents > 0 ? (
                      <>
                        Please verify your email address to continue.
                        <br />A verification email was sent to you.
                      </>
                    ) : (
                      <>
                        Verify your email address to recieve $100 of credits.
                        <br />A verification email was sent to you.
                      </>
                    )}
                  </div>
                )}
                {!wallet.creditCardConnected && user?.profile.email_verified && (
                  <div className="text-sm text-red-500">
                    {selectedOrganization?.personal ? (
                      <>Connect a credit card to receive an additional $100 of credits.</>
                    ) : (
                      <>Please connect your credit card to your account to continue using our service.</>
                    )}
                    <div className="mt-2">
                      <Button variant="secondary" size="icon" className="w-44" onClick={handleUpdatePaymentMethod}>
                        <CreditCard className="w-20 h-20" />
                        Connect
                      </Button>
                    </div>
                  </div>
                )}
                {wallet.creditCardConnected &&
                  user?.profile.email_verified &&
                  (billingPortalUrlLoading || !billingPortalUrl ? (
                    <Skeleton className="max-w-sm h-10" />
                  ) : (
                    <div className="mt-2 grid grid-cols-1 gap-2">
                      <a href={billingPortalUrl ?? ''} target="_blank" rel="noopener noreferrer">
                        <Button variant="secondary" size="icon" className="w-full">
                          <ArrowUpRight className="w-20 h-20" />
                          Top up
                        </Button>
                      </a>
                      <Button
                        variant="secondary"
                        size="icon"
                        className="w-full p-2"
                        onClick={handleUpdatePaymentMethod}
                      >
                        <CreditCard className="w-20 h-20" />
                        Update Payment Method
                      </Button>
                    </div>
                  ))}
              </CardDescription>
            )}
          </CardContent>
        </Card>
        {wallet?.creditCardConnected && (
          <Card className="my-4 w-full max-w-sm">
            <CardHeader>
              <Tooltip
                label={
                  <CardTitle className="flex items-center gap-2">
                    <Info className="w-4 h-4" />
                    Automatic Top up
                  </CardTitle>
                }
                side="bottom"
                content={
                  <div className="flex flex-col gap-2 max-w-sm">
                    <div>
                      <strong>Threshold</strong> is the amount of credit you want to have in your account before they
                      are automatically topped up.
                    </div>
                    <div>
                      <strong>Target</strong> is the amount of credit you want to have in your account after they are
                      automatically topped up. The target must always be greater than the threshold by{' '}
                      <strong>at least $10</strong>.
                    </div>
                    <div>Setting both values to 0 will disable automatic top-ups.</div>
                  </div>
                }
              />
            </CardHeader>
            <CardContent className="p-6 w-full">
              {walletLoading || !wallet ? (
                <Skeleton className="max-w-sm h-10" />
              ) : (
                <CardDescription>
                  <div className="flex flex-col gap-6">
                    <div className="flex justify-between items-end">
                      <Label>Threshold ($)</Label>
                      <Input
                        type="number"
                        className="w-24"
                        value={automaticTopUp?.thresholdAmount ?? 0}
                        onChange={(e) => {
                          let targetAmount = automaticTopUp?.targetAmount ?? 0
                          if (Number(e.target.value) > targetAmount - 10) {
                            targetAmount = Number(e.target.value) + 10
                          }

                          setAutomaticTopUp({
                            thresholdAmount: Number(e.target.value),
                            targetAmount,
                          })
                        }}
                      />
                    </div>
                    <Slider
                      defaultValue={[wallet.automaticTopUp?.thresholdAmount ?? 0]}
                      max={1000}
                      min={0}
                      step={0.5}
                      className="mb-4"
                      value={automaticTopUp?.thresholdAmount ? [automaticTopUp.thresholdAmount] : undefined}
                      onValueChange={(value) => {
                        let targetAmount = automaticTopUp?.targetAmount ?? 0
                        if (value[0] > targetAmount - 10) {
                          targetAmount = value[0] + 10
                        }

                        setAutomaticTopUp({
                          thresholdAmount: value[0],
                          targetAmount,
                        })
                      }}
                    />
                    <div className="flex justify-between items-end">
                      <Label>Target ($)</Label>
                      <Input
                        type="number"
                        className="w-24"
                        value={automaticTopUp?.targetAmount ?? 0}
                        onBlur={(e) => {
                          const thresholdAmount = automaticTopUp?.thresholdAmount ?? 0
                          if (Number(e.target.value) < thresholdAmount) {
                            setAutomaticTopUp({
                              thresholdAmount,
                              targetAmount: thresholdAmount,
                            })
                          }
                        }}
                        onChange={(e) => {
                          const thresholdAmount = automaticTopUp?.thresholdAmount ?? 0
                          setAutomaticTopUp({
                            thresholdAmount,
                            targetAmount: Number(e.target.value),
                          })
                        }}
                      />
                    </div>
                    <Slider
                      defaultValue={[wallet.automaticTopUp?.targetAmount ?? 0]}
                      max={1000}
                      min={0}
                      step={0.5}
                      value={automaticTopUp?.targetAmount ? [automaticTopUp.targetAmount] : undefined}
                      onValueChange={(value) => {
                        const thresholdAmount = automaticTopUp?.thresholdAmount ?? 0
                        if (value[0] <= 10 && value[0] < thresholdAmount) {
                          return
                        }

                        if (value[0] > 10 && value[0] < thresholdAmount + 10) {
                          return
                        }

                        setAutomaticTopUp({
                          thresholdAmount,
                          targetAmount: value[0],
                        })
                      }}
                    />
                    <div>
                      <Button
                        variant="secondary"
                        size="icon"
                        className="w-44 mt-4"
                        onClick={handleSetAutomaticTopUp}
                        disabled={saveAutomaticTopUpDisabled}
                      >
                        {automaticTopUpLoading ? <Loader2 className="w-20 h-20 animate-spin" /> : 'Save'}
                      </Button>
                    </div>
                  </div>
                </CardDescription>
              )}
            </CardContent>
          </Card>
        )}
        {user?.profile.email_verified && !!wallet && (
          <Card className="my-4 w-full max-w-sm h-full">
            <CardHeader>
              <CardTitle>Redeem Coupon</CardTitle>
            </CardHeader>
            <CardContent className="p-6 w-full">
              <Input placeholder="Coupon Code" value={couponCode} onChange={(e) => setCouponCode(e.target.value)} />
              <Button
                variant="secondary"
                disabled={redeemCouponLoading || !couponCode}
                className="mt-4 w-20"
                onClick={handleRedeemCoupon}
              >
                {redeemCouponLoading ? <Loader2 className="w-20 h-20 animate-spin" /> : 'Redeem'}
              </Button>
              {redeemCouponError && <div className="text-red-500 mt-4">{redeemCouponError}</div>}
              {redeemCouponSuccess && <div className="text-green-500 mt-4">{redeemCouponSuccess}</div>}
            </CardContent>
          </Card>
        )}
      </div>

      {/* Organization Emails Section */}
      <div className="mt-8">
        <Card>
          <CardHeader>
            <CardTitle>Billing Emails</CardTitle>
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
