/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Invoice } from '@/billing-api/types/Invoice'
import { AutomaticTopUp } from '@/billing-api/types/OrganizationWallet'
import { InvoicesTable } from '@/components/Invoices'
import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { InputGroup, InputGroupAddon, InputGroupInput, InputGroupText } from '@/components/ui/input-group'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { useCreateInvoicePaymentUrlMutation } from '@/hooks/mutations/useCreateInvoicePaymentUrlMutation'
import { useRedeemCouponMutation } from '@/hooks/mutations/useRedeemCouponMutation'
import { useSetAutomaticTopUpMutation } from '@/hooks/mutations/useSetAutomaticTopUpMutation'
import { useTopUpWalletMutation } from '@/hooks/mutations/useTopUpWalletMutation'
import { useVoidInvoiceMutation } from '@/hooks/mutations/useVoidInvoiceMutation'
import {
  useFetchOwnerCheckoutUrlQuery,
  useIsOwnerCheckoutUrlFetching,
  useOwnerBillingPortalUrlQuery,
  useOwnerInvoicesQuery,
  useOwnerWalletQuery,
} from '@/hooks/queries/billingQueries'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { formatAmount } from '@/lib/utils'
import { ArrowUpRight, CheckCircleIcon, InfoIcon, SparklesIcon, TriangleAlertIcon } from 'lucide-react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { NumericFormat } from 'react-number-format'
import { useAuth } from 'react-oidc-context'
import { toast } from 'sonner'

const DEFAULT_PAGE_SIZE = 10

const Wallet = () => {
  const { selectedOrganization } = useSelectedOrganization()
  const { user } = useAuth()
  const [automaticTopUp, setAutomaticTopUp] = useState<AutomaticTopUp | undefined>(undefined)
  const [couponCode, setCouponCode] = useState<string>('')
  const [redeemCouponError, setRedeemCouponError] = useState<string | null>(null)
  const [redeemCouponSuccess, setRedeemCouponSuccess] = useState<string | null>(null)
  const [oneTimeTopUpAmount, setOneTimeTopUpAmount] = useState<number | undefined>(undefined)
  const [selectedPreset, setSelectedPreset] = useState<number | null>(null)
  const [invoicesPagination, setInvoicesPagination] = useState({
    pageIndex: 0,
    pageSize: DEFAULT_PAGE_SIZE,
  })
  const walletQuery = useOwnerWalletQuery({ refetchOnMount: 'always' })
  const billingPortalUrlQuery = useOwnerBillingPortalUrlQuery()
  const invoicesQuery = useOwnerInvoicesQuery(invoicesPagination.pageIndex + 1, invoicesPagination.pageSize)

  const isCheckoutUrlLoading = useIsOwnerCheckoutUrlFetching()
  const fetchCheckoutUrl = useFetchOwnerCheckoutUrlQuery()
  const wallet = walletQuery.data
  const billingPortalUrl = billingPortalUrlQuery.data
  const setAutomaticTopUpMutation = useSetAutomaticTopUpMutation()
  const redeemCouponMutation = useRedeemCouponMutation()
  const topUpWalletMutation = useTopUpWalletMutation()
  const createInvoicePaymentUrlMutation = useCreateInvoicePaymentUrlMutation()
  const voidInvoiceMutation = useVoidInvoiceMutation()

  useEffect(() => {
    if (wallet?.automaticTopUp) {
      setAutomaticTopUp(wallet.automaticTopUp)
    }
  }, [wallet])

  const handleUpdatePaymentMethod = useCallback(async () => {
    try {
      const data = await fetchCheckoutUrl()
      window.open(data, '_blank')
    } catch (error) {
      toast.error('Failed to fetch checkout url', {
        description: String(error),
      })
    }
  }, [fetchCheckoutUrl])

  const handleSetAutomaticTopUp = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }

    try {
      await setAutomaticTopUpMutation.mutateAsync({
        organizationId: selectedOrganization.id,
        automaticTopUp,
      })
      toast.success('Automatic top up set successfully')
    } catch (error) {
      toast.error('Failed to set automatic top up', {
        description: String(error),
      })
    }
  }, [selectedOrganization, automaticTopUp, setAutomaticTopUpMutation])

  const handleRedeemCoupon = useCallback(async () => {
    if (!selectedOrganization || !couponCode) {
      return
    }

    setRedeemCouponError(null)
    setRedeemCouponSuccess(null)

    try {
      const message = await redeemCouponMutation.mutateAsync({
        organizationId: selectedOrganization.id,
        couponCode,
      })
      setRedeemCouponSuccess(message)
      setTimeout(() => {
        setRedeemCouponSuccess(null)
      }, 3000)
      setCouponCode('')
    } catch (error) {
      setRedeemCouponError(String(error))
      console.error('Failed to redeem coupon:', error)
    }
  }, [selectedOrganization, couponCode, redeemCouponMutation])

  const saveAutomaticTopUpDisabled = useMemo(() => {
    if (setAutomaticTopUpMutation.isPending) {
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
  }, [setAutomaticTopUpMutation.isPending, wallet, automaticTopUp])

  const handleTopUpWallet = useCallback(async () => {
    if (!selectedOrganization) {
      return
    }
    const amount = selectedPreset ?? oneTimeTopUpAmount
    if (!amount) {
      return
    }

    try {
      const result = await topUpWalletMutation.mutateAsync({
        organizationId: selectedOrganization.id,
        amountCents: amount * 100,
      })
      window.open(result.url, '_blank')
    } catch (error) {
      toast.error('Failed to initiate top-up', {
        description: String(error),
      })
    }
  }, [selectedOrganization, selectedPreset, oneTimeTopUpAmount, topUpWalletMutation])

  const handlePayInvoice = useCallback(
    async (invoice: Invoice) => {
      if (!selectedOrganization) {
        return
      }

      if (invoice.paymentStatus === 'pending' && invoice.totalDueAmountCents > 0) {
        try {
          const result = await createInvoicePaymentUrlMutation.mutateAsync({
            organizationId: selectedOrganization.id,
            invoiceId: invoice.id,
          })
          window.open(result.url, '_blank')
        } catch (error) {
          toast.error('Failed to open invoice', {
            description: String(error),
          })
        }
      }
    },
    [selectedOrganization, createInvoicePaymentUrlMutation],
  )

  const handleViewInvoice = useCallback(
    async (invoice: Invoice) => {
      if (!selectedOrganization) {
        return
      }

      window.open(invoice.fileUrl, '_blank')
    },
    [selectedOrganization],
  )

  const handleVoidInvoice = useCallback(
    async (invoice: Invoice) => {
      if (!selectedOrganization) {
        return
      }
      try {
        await voidInvoiceMutation.mutateAsync({
          organizationId: selectedOrganization.id,
          invoiceId: invoice.id,
        })
        toast.success('Invoice voided successfully')
      } catch (error) {
        toast.error('Failed to void invoice', {
          description: String(error),
        })
      }
    },
    [selectedOrganization, voidInvoiceMutation],
  )

  const isBillingLoading = walletQuery.isLoading && billingPortalUrlQuery.isLoading
  const topUpEnabled =
    wallet?.creditCardConnected && !topUpWalletMutation.isPending && (selectedPreset || oneTimeTopUpAmount)

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Wallet</PageTitle>
      </PageHeader>

      <PageContent>
        {isBillingLoading && (
          <div className="flex flex-col gap-6">
            <Card className="flex flex-col gap-4">
              <CardContent className="flex flex-col gap-4">
                <Skeleton className="h-5 w-full max-w-sm" />
                <div className="flex items-center gap-2">
                  <Skeleton className="h-10 flex-1" />
                  <Skeleton className="h-10 flex-1" />
                </div>
                <Skeleton className=" h-10" />
                <Skeleton className=" h-10" />
              </CardContent>
            </Card>
            <Card className="flex flex-col gap-4">
              <CardContent className="flex flex-col gap-4">
                <Skeleton className="h-5 w-full max-w-sm" />
                <div className="flex items-center gap-2">
                  <Skeleton className="h-10 flex-1" />
                  <Skeleton className="h-10 flex-1" />
                </div>
                <Skeleton className=" h-10" />
              </CardContent>
            </Card>
          </div>
        )}
        {wallet && (
          <>
            {user && (
              <>
                {!user.profile.email_verified && (
                  <Alert variant="info">
                    <TriangleAlertIcon />
                    <AlertTitle>Verify your email</AlertTitle>
                    <AlertDescription>
                      {wallet.balanceCents && wallet.balanceCents > 0 ? (
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
                {!wallet.creditCardConnected && user.profile.email_verified && selectedOrganization?.personal && (
                  <Alert variant="neutral">
                    <SparklesIcon />
                    <AlertDescription>Connect a credit card to receive an additional $100 of credits.</AlertDescription>
                  </Alert>
                )}
              </>
            )}
            {wallet.hasFailedOrPendingInvoice && (
              <Alert variant="destructive">
                <TriangleAlertIcon />
                <AlertTitle>Outstanding invoices</AlertTitle>
                <AlertDescription>
                  You have failed or pending invoices that need to be resolved before adding new funds. Please review
                  your invoices below and complete or void any outstanding payments.
                </AlertDescription>
              </Alert>
            )}

            <Card className="h-full">
              <CardHeader>
                <CardTitle>Overview</CardTitle>
              </CardHeader>
              <CardContent className="">
                <div className="flex items-start sm:flex-row flex-col gap-4 sm:items-end justify-between">
                  <div className="flex gap-4 sm:gap-12 sm:flex-row flex-col">
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
                  {billingPortalUrlQuery.isLoading ? (
                    <Skeleton className="h-8 w-[160px]" />
                  ) : billingPortalUrl ? (
                    <Button variant="link" asChild className="flex items-center gap-2 !px-0">
                      <a
                        href={`${billingPortalUrl}/customer-edit-information`}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        Update Billing Info
                        <ArrowUpRight />
                      </a>
                    </Button>
                  ) : null}
                </div>
              </CardContent>
              <CardContent className="border-t border-border">
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
                    <Button variant="default" onClick={handleUpdatePaymentMethod} disabled={isCheckoutUrlLoading}>
                      {isCheckoutUrlLoading && <Spinner />} Connect
                    </Button>
                  ) : (
                    <Button variant="secondary" onClick={handleUpdatePaymentMethod} disabled={isCheckoutUrlLoading}>
                      {isCheckoutUrlLoading && <Spinner />} Update
                    </Button>
                  )}
                </div>
              </CardContent>

              {user?.profile.email_verified && (
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
                      <Button
                        variant="secondary"
                        onClick={handleRedeemCoupon}
                        disabled={redeemCouponMutation.isPending}
                      >
                        {redeemCouponMutation.isPending && <Spinner />} Redeem
                      </Button>
                    </div>
                  </div>
                </CardContent>
              )}
            </Card>

            {wallet.creditCardConnected && (
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
                </CardContent>
                <CardFooter className="flex justify-between gap-2">
                  <div className="flex items-center gap-2 text-muted-foreground">
                    <InfoIcon className="w-4 h-4 shrink-0" />{' '}
                    <span className="text-sm ">Setting both values to 0 will disable automatic top-ups.</span>
                  </div>
                  <div className="flex gap-2 items-center ml-auto">
                    <Button
                      onClick={handleSetAutomaticTopUp}
                      disabled={saveAutomaticTopUpDisabled || walletQuery.isLoading || !wallet}
                    >
                      {setAutomaticTopUpMutation.isPending && <Spinner />} Save
                    </Button>
                  </div>
                </CardFooter>
              </Card>
            )}

            <Card className="w-full">
              <CardHeader>
                <CardTitle>One time top-up</CardTitle>
                <CardDescription>
                  Add funds to your wallet instantly. Select a preset amount or enter a custom value.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 gap-10 items-center lg:grid-cols-2">
                  <div className="flex flex-col gap-2">
                    <Label className="text-sm font-medium">Select amount</Label>
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                      {[25, 500, 1000, 2000].map((amount) => (
                        <Button
                          key={amount}
                          type="button"
                          variant={selectedPreset === amount ? 'default' : 'outline'}
                          size="default"
                          className="flex h-9"
                          onClick={() => {
                            setSelectedPreset(amount)
                            setOneTimeTopUpAmount(undefined)
                          }}
                        >
                          <span className="font-semibold">${amount.toLocaleString()}</span>
                        </Button>
                      ))}
                    </div>
                  </div>
                  <div className="flex items-center gap-3 lg:hidden">
                    <div className="flex-1 h-px bg-border" />
                    <span className="text-sm text-muted-foreground">or</span>
                    <div className="flex-1 h-px bg-border" />
                  </div>
                  <div className="flex flex-col gap-2">
                    <Label htmlFor="customTopUpAmount" className="text-sm font-medium">
                      Enter custom amount
                    </Label>
                    <InputGroup>
                      <InputGroupAddon>
                        <InputGroupText>$</InputGroupText>
                      </InputGroupAddon>
                      <NumericFormat
                        placeholder="0.00"
                        customInput={InputGroupInput}
                        id="customTopUpAmount"
                        inputMode="decimal"
                        thousandSeparator
                        decimalScale={2}
                        value={oneTimeTopUpAmount ?? ''}
                        onValueChange={({ floatValue }) => {
                          const value = floatValue ?? undefined
                          setOneTimeTopUpAmount(value)
                          setSelectedPreset(null)
                        }}
                        onFocus={() => {
                          setSelectedPreset(null)
                        }}
                      />
                      <InputGroupAddon align="inline-end">
                        <InputGroupText>USD</InputGroupText>
                      </InputGroupAddon>
                    </InputGroup>
                  </div>
                </div>
              </CardContent>
              <CardFooter className="flex justify-between gap-2">
                <div className="text-sm text-muted-foreground">
                  You will be redirected to Stripe to complete the payment.
                </div>
                <Button onClick={handleTopUpWallet} disabled={!topUpEnabled} size="sm">
                  {topUpWalletMutation.isPending && <Spinner />}
                  Top up
                </Button>
              </CardFooter>
            </Card>

            <Card className="w-full">
              <CardHeader>
                <CardTitle>Invoices</CardTitle>
                <CardDescription>
                  View and download your billing invoices. All invoices are automatically generated and sent to your
                  billing emails.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <InvoicesTable
                  data={invoicesQuery.data?.items ?? []}
                  pagination={invoicesPagination}
                  pageCount={invoicesQuery.data?.totalPages ?? 0}
                  totalItems={invoicesQuery.data?.totalItems ?? 0}
                  onPaginationChange={setInvoicesPagination}
                  loading={invoicesQuery.isLoading}
                  onViewInvoice={handleViewInvoice}
                  onVoidInvoice={handleVoidInvoice}
                  onPayInvoice={handlePayInvoice}
                />
              </CardContent>
            </Card>
          </>
        )}
      </PageContent>
    </PageLayout>
  )
}

export default Wallet
