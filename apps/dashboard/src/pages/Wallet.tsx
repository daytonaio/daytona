'use client'

/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { AutomaticTopUp } from '@/billing-api/types/OrganizationWallet'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useEffect, useMemo, useState } from 'react'
import { useCallback } from 'react'
import { Card } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { CreditCard, Info, Loader2, TrendingUp, DollarSign, RefreshCw } from 'lucide-react'
import { toast } from 'sonner'
import { Input } from '@/components/ui/input'
import { useApi } from '@/hooks/useApi'
import { useAuth } from 'react-oidc-context'
import type { OrganizationEmail } from '@/billing-api'
import { OrganizationEmailsTable } from '@/components/OrganizationEmails'
import { Progress } from '@/components/ui/progress'
import { useBilling } from '@/hooks/useBilling'

const formatAmount = (amount: number) => {
  return Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount / 100)
}

const CURRENCY_SYMBOL = '$'

const AutomaticTopUpSettings = ({
  wallet,
  automaticTopUp,
  setAutomaticTopUp,
  onSave,
  loading,
}: {
  wallet: any
  automaticTopUp: AutomaticTopUp | undefined
  setAutomaticTopUp: (config: AutomaticTopUp) => void
  onSave: () => void
  loading: boolean
}) => {
  const [isEnabled, setIsEnabled] = useState(false)
  const [selectedPreset, setSelectedPreset] = useState<string>('moderate')
  const [customThreshold, setCustomThreshold] = useState(25)
  const [customAmount, setCustomAmount] = useState(100)

  const presets = {
    conservative: {
      threshold: 25,
      amount: 60,
      label: 'Conservative',
      description: `Add ${CURRENCY_SYMBOL}60 when below ${CURRENCY_SYMBOL}25`,
    },
    moderate: {
      threshold: 50,
      amount: 100,
      label: 'Moderate',
      description: `Add ${CURRENCY_SYMBOL}100 when below ${CURRENCY_SYMBOL}50`,
    },
    aggressive: {
      threshold: 100,
      amount: 250,
      label: 'Aggressive',
      description: `Add ${CURRENCY_SYMBOL}250 when below ${CURRENCY_SYMBOL}100`,
    },
    custom: {
      threshold: customThreshold,
      amount: customAmount,
      label: 'Custom',
      description: `Add ${CURRENCY_SYMBOL}${customAmount} when below ${CURRENCY_SYMBOL}${customThreshold}`,
    },
  }

  useEffect(() => {
    if (wallet?.automaticTopUp) {
      setIsEnabled(true)
      const matchingPreset = Object.entries(presets).find(
        ([_, preset]) =>
          preset.threshold === wallet.automaticTopUp.thresholdAmount &&
          preset.amount === wallet.automaticTopUp.targetAmount,
      )
      setSelectedPreset(matchingPreset ? matchingPreset[0] : 'moderate')
    }
  }, [wallet])

  const handlePresetChange = (presetKey: string) => {
    setSelectedPreset(presetKey)
    const preset = presets[presetKey as keyof typeof presets]
    setAutomaticTopUp({
      thresholdAmount: preset.threshold,
      targetAmount: preset.amount,
    })
  }

  const handleToggle = (enabled: boolean) => {
    setIsEnabled(enabled)
    if (!enabled) {
      setAutomaticTopUp({ thresholdAmount: 0, targetAmount: 0 })
    } else {
      const preset = presets[selectedPreset as keyof typeof presets]
      setAutomaticTopUp({
        thresholdAmount: preset.threshold,
        targetAmount: preset.amount,
      })
    }
  }

  const handleCustomThresholdChange = (value: string) => {
    const numericValue = value.replace(/[^0-9]/g, '')
    const numValue = Number.parseInt(numericValue) || 0
    setCustomThreshold(numValue)
    if (selectedPreset === 'custom') {
      setAutomaticTopUp({
        thresholdAmount: numValue,
        targetAmount: customAmount,
      })
    }
  }

  const handleCustomAmountChange = (value: string) => {
    const numericValue = value.replace(/[^0-9]/g, '')
    const numValue = Number.parseInt(numericValue) || 0
    setCustomAmount(numValue)
    if (selectedPreset === 'custom') {
      setAutomaticTopUp({
        thresholdAmount: customThreshold,
        targetAmount: numValue,
      })
    }
  }

  const currentPreset = presets[selectedPreset as keyof typeof presets]
  const estimatedDays =
    wallet && currentPreset
      ? Math.floor(currentPreset.threshold / ((wallet.balanceCents - wallet.ongoingBalanceCents) / 30))
      : 0

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => handleToggle(!isEnabled)}
            className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 ${
              isEnabled ? 'bg-primary' : 'bg-muted'
            }`}
          >
            <span
              className={`inline-block h-3 w-3 transform rounded-full transition-transform ${
                isEnabled ? 'translate-x-5 bg-primary-foreground' : 'translate-x-1 bg-white'
              }`}
            />
          </button>
          <span className="text-sm font-medium">Enable Automatic Top-up</span>
        </div>
        {isEnabled && (
          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <RefreshCw size={12} strokeWidth={1.5} />
            <span>Active</span>
          </div>
        )}
      </div>

      {isEnabled && (
        <div className="space-y-4 pl-6 border-l border-primary/20 bg-muted/20 p-4 rounded-r-lg">
          <div className="space-y-3">
            {Object.entries(presets).map(([key, preset]) => (
              <div key={key} className="space-y-1">
                <label className="flex items-center gap-3 cursor-pointer p-2 rounded-lg hover:bg-muted/50 transition-colors">
                  <div className="relative">
                    <input
                      type="radio"
                      name="preset"
                      value={key}
                      checked={selectedPreset === key}
                      onChange={() => handlePresetChange(key)}
                      className="sr-only"
                    />
                    <div
                      className={`w-4 h-4 rounded-full border-2 transition-colors ${
                        selectedPreset === key ? 'border-primary bg-primary' : 'border-muted-foreground bg-transparent'
                      }`}
                    >
                      {selectedPreset === key && (
                        <div className="w-2 h-2 bg-white rounded-full absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2" />
                      )}
                    </div>
                  </div>
                  <div className="flex-1">
                    <div className="font-medium text-sm">{preset.label}</div>
                    <div className="text-xs text-muted-foreground">{preset.description}</div>
                  </div>
                </label>
                {key === 'custom' && selectedPreset === 'custom' && (
                  <div className="ml-7 mt-2 space-y-2">
                    <div className="grid grid-cols-2 gap-2">
                      <div>
                        <label className="text-xs text-muted-foreground">Minimum Balance</label>
                        <div className="flex items-center">
                          <span className="text-sm mr-1">{CURRENCY_SYMBOL}</span>
                          <Input
                            type="text"
                            value={customThreshold}
                            onChange={(e) => handleCustomThresholdChange(e.target.value)}
                            className="h-8 text-sm"
                          />
                        </div>
                      </div>
                      <div>
                        <label className="text-xs text-muted-foreground">Top-up Amount</label>
                        <div className="flex items-center">
                          <span className="text-sm mr-1">{CURRENCY_SYMBOL}</span>
                          <Input
                            type="text"
                            value={customAmount}
                            onChange={(e) => handleCustomAmountChange(e.target.value)}
                            className="h-8 text-sm"
                          />
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>

          {estimatedDays > 0 && (
            <div className="text-xs text-muted-foreground bg-muted/50 p-3 rounded-lg border border-muted">
              <Info size={12} strokeWidth={1.5} className="inline mr-1" />
              Based on your usage, this will trigger approximately every {estimatedDays} days
            </div>
          )}

          <Button variant="outline" size="sm" className="w-full bg-transparent" onClick={onSave} disabled={loading}>
            {loading ? <Loader2 size={16} strokeWidth={1.5} className="mr-2 animate-spin" /> : null}
            Save Auto Top-up Settings
          </Button>

          <div className="mt-6 pt-4 border-t border-muted">
            <h4 className="text-sm font-medium mb-3">Recent Automatic Top-ups</h4>
            <div className="space-y-2">
              <div className="flex items-center justify-between text-xs p-2 bg-muted/30 rounded-lg">
                <div className="flex items-center gap-2">
                  <RefreshCw size={12} strokeWidth={1.5} className="text-green-500" />
                  <span>Dec 15, 2024</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-muted-foreground">Added</span>
                  <span className="font-medium">{CURRENCY_SYMBOL}100</span>
                </div>
              </div>
              <div className="flex items-center justify-between text-xs p-2 bg-muted/30 rounded-lg">
                <div className="flex items-center gap-2">
                  <RefreshCw size={12} strokeWidth={1.5} className="text-green-500" />
                  <span>Dec 8, 2024</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-muted-foreground">Added</span>
                  <span className="font-medium">{CURRENCY_SYMBOL}100</span>
                </div>
              </div>
              <div className="flex items-center justify-between text-xs p-2 bg-muted/30 rounded-lg">
                <div className="flex items-center gap-2">
                  <RefreshCw size={12} strokeWidth={1.5} className="text-green-500" />
                  <span>Nov 30, 2024</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-muted-foreground">Added</span>
                  <span className="font-medium">{CURRENCY_SYMBOL}100</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

const SmartTopUpSuggestions = () => {
  const [selectedAmount, setSelectedAmount] = useState(100)
  const [isCustom, setIsCustom] = useState(false)
  const [customAmount, setCustomAmount] = useState('')

  const predefinedAmounts = [30, 60, 100, 250]

  const handleAmountSelect = (amount: number) => {
    setSelectedAmount(amount)
    setIsCustom(false)
    setCustomAmount('')
  }

  const handleCustomSelect = () => {
    setIsCustom(true)
    setCustomAmount(selectedAmount.toString())
  }

  const handleCustomAmountChange = (value: string) => {
    const numericValue = value.replace(/[^0-9.]/g, '')
    setCustomAmount(numericValue)
    const parsed = Number.parseFloat(numericValue)
    if (!isNaN(parsed)) {
      setSelectedAmount(parsed)
    }
  }

  const displayAmount = isCustom ? customAmount || '0' : selectedAmount.toString()

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-center h-16">
        {isCustom ? (
          <input
            type="text"
            value={`${CURRENCY_SYMBOL} ${customAmount}`}
            onChange={(e) => handleCustomAmountChange(e.target.value.replace(`${CURRENCY_SYMBOL} `, ''))}
            className="text-4xl font-bold bg-transparent border-none outline-none text-center text-white"
            style={{
              color: '#ffffff !important',
              width: `${Math.max(4, displayAmount.length + 2)}ch`,
            }}
            autoFocus
          />
        ) : (
          <div className="text-4xl font-bold text-white">
            {CURRENCY_SYMBOL} {selectedAmount}
          </div>
        )}
      </div>

      <div className="grid grid-cols-2 gap-2">
        {predefinedAmounts.map((amount) => (
          <Button
            key={amount}
            variant={selectedAmount === amount && !isCustom ? 'default' : 'outline'}
            size="sm"
            onClick={() => handleAmountSelect(amount)}
            className={`transition-all ${
              selectedAmount === amount && !isCustom
                ? 'bg-primary text-primary-foreground border-primary shadow-md'
                : 'bg-transparent hover:bg-muted'
            }`}
          >
            {CURRENCY_SYMBOL}
            {amount}
          </Button>
        ))}
      </div>

      <Button
        variant={isCustom ? 'default' : 'outline'}
        size="sm"
        onClick={handleCustomSelect}
        className={`w-full transition-all ${
          isCustom ? 'bg-primary text-primary-foreground border-primary shadow-md' : 'bg-transparent hover:bg-muted'
        }`}
      >
        Custom
      </Button>

      <div className="flex gap-2">
        <Button size="sm" className="w-full">
          Continue to Payment
        </Button>
      </div>
    </div>
  )
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
      await billingApi.redeemCoupon(selectedOrganization.id, couponCode)
      setRedeemCouponSuccess('Coupon redeemed successfully')
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
    if (wallet && !automaticTopUp) {
      setAutomaticTopUp({
        thresholdAmount: wallet.automaticTopUp?.thresholdAmount ?? 50,
        targetAmount: wallet.automaticTopUp?.targetAmount ?? 100,
      })
    }
  }, [wallet, automaticTopUp])

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Wallet</h1>
      </div>

      <Card className="p-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="space-y-2">
            {walletLoading || !wallet ? (
              <Skeleton className="h-16 w-full" />
            ) : (
              <>
                <div className="flex items-center gap-2">
                  <DollarSign size={16} strokeWidth={1.5} className="text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">Current Balance</span>
                </div>
                <div className="text-3xl font-bold">{formatAmount(wallet.ongoingBalanceCents)}</div>
              </>
            )}
          </div>

          <div className="space-y-2">
            {walletLoading || !wallet ? (
              <Skeleton className="h-16 w-full" />
            ) : (
              <>
                <div className="flex items-center gap-2">
                  <TrendingUp size={16} strokeWidth={1.5} className="text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">This Month</span>
                </div>
                <div className="text-3xl font-bold">
                  {formatAmount(wallet.balanceCents - wallet.ongoingBalanceCents)}
                </div>
              </>
            )}
          </div>

          <div className="space-y-2">
            {walletLoading || !wallet ? (
              <Skeleton className="h-16 w-full" />
            ) : (
              <>
                <div className="flex items-center gap-2">
                  <RefreshCw size={16} strokeWidth={1.5} className="text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">Estimated Runway</span>
                </div>
                <div className="text-3xl font-bold">
                  {wallet.ongoingBalanceCents > 0
                    ? `${Math.floor(wallet.ongoingBalanceCents / ((wallet.balanceCents - wallet.ongoingBalanceCents) / 30))} days`
                    : '∞'}
                </div>
              </>
            )}
          </div>
        </div>

        {!walletLoading && wallet && wallet.balanceCents - wallet.ongoingBalanceCents > 0 && (
          <div className="mt-6 space-y-2">
            <div className="flex justify-between text-sm text-muted-foreground">
              <span>Balance Health</span>
              <span>{Math.round((wallet.ongoingBalanceCents / wallet.balanceCents) * 100)}%</span>
            </div>
            <Progress value={(wallet.ongoingBalanceCents / wallet.balanceCents) * 100} className="h-2" />
          </div>
        )}
      </Card>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {wallet?.creditCardConnected && user?.profile.email_verified && (
          <Card className="p-4">
            <div className="space-y-4">
              <h3 className="text-lg font-medium">Top up</h3>
              <SmartTopUpSuggestions />
              <Button variant="outline" size="sm" onClick={handleUpdatePaymentMethod} className="w-full bg-transparent">
                <CreditCard size={16} strokeWidth={1.5} className="mr-2" />
                Update Payment Method
              </Button>
            </div>
          </Card>
        )}

        {!wallet?.creditCardConnected && user?.profile.email_verified && (
          <Card className="p-4">
            <div className="space-y-4">
              <h3 className="text-lg font-medium">Get Started</h3>
              <p className="text-sm text-muted-foreground">
                Connect your payment method to start using Daytona with automatic billing.
              </p>
              <Button variant="outline" size="sm" onClick={handleUpdatePaymentMethod} className="w-full bg-transparent">
                <CreditCard size={16} strokeWidth={1.5} className="mr-2" />
                Connect Card
              </Button>
            </div>
          </Card>
        )}

        {wallet?.creditCardConnected && (
          <Card className="p-4">
            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <RefreshCw size={16} strokeWidth={1.5} className="text-muted-foreground" />
                <h3 className="text-lg font-medium">Automatic Top-up</h3>
              </div>
              <p className="text-xs text-muted-foreground">
                Never run out of credits. We'll automatically add funds when your balance gets low.
              </p>

              {walletLoading || !wallet ? (
                <Skeleton className="h-32 w-full" />
              ) : (
                <AutomaticTopUpSettings
                  wallet={wallet}
                  automaticTopUp={automaticTopUp}
                  setAutomaticTopUp={setAutomaticTopUp}
                  onSave={handleSetAutomaticTopUp}
                  loading={automaticTopUpLoading}
                />
              )}
            </div>
          </Card>
        )}

        {user?.profile.email_verified && !!wallet && (
          <Card className="p-4">
            <div className="space-y-4">
              <h3 className="text-lg font-medium">Redeem Coupon</h3>

              <div className="space-y-3">
                <Input
                  placeholder="Enter coupon code"
                  value={couponCode}
                  onChange={(e) => setCouponCode(e.target.value)}
                  className="text-sm"
                />
                <Button
                  variant="outline"
                  size="sm"
                  disabled={redeemCouponLoading || !couponCode}
                  className="w-full bg-transparent"
                  onClick={handleRedeemCoupon}
                >
                  {redeemCouponLoading ? <Loader2 size={16} strokeWidth={1.5} className="mr-2 animate-spin" /> : null}
                  Redeem Coupon
                </Button>

                {redeemCouponError && (
                  <div className="text-sm text-destructive p-2 bg-destructive/10 rounded-md border border-destructive/20">
                    {redeemCouponError}
                  </div>
                )}
                {redeemCouponSuccess && (
                  <div className="text-sm text-green-600 dark:text-green-400 p-2 bg-green-50 dark:bg-green-950/20 rounded-md border border-green-200 dark:border-green-800">
                    {redeemCouponSuccess}
                  </div>
                )}
              </div>
            </div>
          </Card>
        )}
      </div>

      {!user?.profile.email_verified && wallet && (
        <Card className="p-4 border-amber-200 dark:border-amber-800 bg-amber-50 dark:bg-amber-950/20">
          <div className="text-sm text-amber-800 dark:text-amber-200">
            {wallet?.balanceCents > 0 ? (
              <>Please verify your email address to continue. A verification email was sent to you.</>
            ) : (
              <>
                Verify your email address to receive {CURRENCY_SYMBOL}100 of credits. A verification email was sent to
                you.
              </>
            )}
          </div>
        </Card>
      )}

      <Card className="p-4">
        <div className="space-y-4">
          <div>
            <h3 className="text-lg font-medium">Billing Emails</h3>
            <p className="text-sm text-muted-foreground mt-1">
              Manage billing emails for your organization which receive important billing notifications such as invoices
              and credit depletion notices.
            </p>
          </div>
          <OrganizationEmailsTable
            data={organizationEmails}
            loading={organizationEmailsLoading}
            handleDelete={handleDeleteEmail}
            handleResendVerification={handleResendVerification}
            handleAddEmail={handleAddEmail}
          />
        </div>
      </Card>
    </div>
  )
}

export default Wallet
