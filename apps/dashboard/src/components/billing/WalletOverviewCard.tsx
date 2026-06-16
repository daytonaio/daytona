/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Spinner } from '@/components/ui/spinner'
import { useRedeemCouponMutation } from '@/hooks/mutations/useRedeemCouponMutation'
import { formatAmount } from '@/lib/utils'
import type { OrganizationWallet } from '@daytona/billing-api-client'
import type { User } from 'oidc-client-ts'
import { useCallback, useState } from 'react'

interface WalletOverviewCardProps {
  organizationId?: string
  wallet: OrganizationWallet
  isPostPaid: boolean
  user?: User | null
}

export function WalletOverviewCard({ organizationId, wallet, isPostPaid, user }: WalletOverviewCardProps) {
  const [couponCode, setCouponCode] = useState('')
  const [redeemCouponError, setRedeemCouponError] = useState<string | null>(null)
  const [redeemCouponSuccess, setRedeemCouponSuccess] = useState<string | null>(null)
  const redeemCouponMutation = useRedeemCouponMutation()

  const handleRedeemCoupon = useCallback(async () => {
    if (!organizationId || !couponCode) {
      return
    }

    setRedeemCouponError(null)
    setRedeemCouponSuccess(null)

    try {
      const message = await redeemCouponMutation.mutateAsync({
        organizationId,
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
  }, [organizationId, couponCode, redeemCouponMutation])

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          Overview
          {isPostPaid && <Badge variant="secondary">Post-paid</Badge>}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex items-start sm:flex-row flex-col gap-4 sm:items-end justify-between">
          <div className="flex gap-4 sm:gap-12 sm:flex-row flex-col">
            <div className="flex flex-col gap-1">
              <div>Current balance</div>
              <div className="text-xl text-foreground font-semibold">
                {formatAmount(wallet.ongoingBalanceCents ?? 0)}
              </div>
            </div>
            <div className="flex flex-col gap-1">
              <div>Spent this month</div>
              <div className="text-xl font-semibold">
                {formatAmount((wallet.balanceCents ?? 0) - (wallet.ongoingBalanceCents ?? 0))}
              </div>
            </div>
          </div>
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
                onChange={(event) => setCouponCode(event.target.value)}
              />
              <Button variant="secondary" onClick={handleRedeemCoupon} disabled={redeemCouponMutation.isPending}>
                {redeemCouponMutation.isPending && <Spinner />}
                Redeem
              </Button>
            </div>
          </div>
        </CardContent>
      )}
    </Card>
  )
}
