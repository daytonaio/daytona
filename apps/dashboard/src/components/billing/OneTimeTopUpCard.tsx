/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { InputGroup, InputGroupAddon, InputGroupInput, InputGroupText } from '@/components/ui/input-group'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { useTopUpWalletMutation } from '@/hooks/mutations/useTopUpWalletMutation'
import { usePaymentMethodsQuery } from '@/hooks/queries/usePaymentMethodsQuery'
import { CreditCardIcon } from 'lucide-react'
import { useCallback, useState } from 'react'
import { NumericFormat } from 'react-number-format'
import { toast } from 'sonner'

const TOP_UP_PRESETS = [25, 500, 1000, 2000]

interface OneTimeTopUpCardProps {
  organizationId: string
}

export function OneTimeTopUpCard({ organizationId }: OneTimeTopUpCardProps) {
  const [oneTimeTopUpAmount, setOneTimeTopUpAmount] = useState<number | undefined>(undefined)
  const [selectedPreset, setSelectedPreset] = useState<number | null>(null)
  const paymentMethodsQuery = usePaymentMethodsQuery({ organizationId })
  const topUpWalletMutation = useTopUpWalletMutation()

  const paymentMethods = paymentMethodsQuery.data
  const paymentMethodsLoading = paymentMethodsQuery.isLoading
  const hasNoPaymentMethod = (paymentMethods?.length ?? 0) === 0
  const showMissingPaymentMethodTopUpMessage = hasNoPaymentMethod
  const amount = selectedPreset ?? oneTimeTopUpAmount
  const topUpEnabled = Boolean(
    !paymentMethodsLoading && !hasNoPaymentMethod && !topUpWalletMutation.isPending && amount,
  )

  const handleTopUpWallet = useCallback(async () => {
    if (!amount) {
      return
    }

    const newWindow = window.open('', '_blank')
    try {
      const result = await topUpWalletMutation.mutateAsync({
        organizationId,
        amountCents: amount * 100,
      })
      if (newWindow) {
        newWindow.location.href = result.url ?? ''
      }
    } catch (error) {
      newWindow?.close()
      toast.error('Failed to initiate top-up', {
        description: String(error),
      })
    }
  }, [organizationId, amount, topUpWalletMutation])

  return (
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
            <div className="grid grid-cols-1 xxs:grid-cols-4 overflow-hidden rounded-md border border-input">
              {TOP_UP_PRESETS.map((preset) => (
                <Button
                  key={preset}
                  type="button"
                  variant={selectedPreset === preset ? 'default' : 'ghost'}
                  size="default"
                  className="flex h-9 min-w-0 rounded-none border-t border-border px-2 text-[13px] first:border-t-0 xxs:border-l xxs:border-t-0 xxs:first:border-l-0"
                  onClick={() => {
                    setSelectedPreset((currentPreset) => (currentPreset === preset ? null : preset))
                    setOneTimeTopUpAmount(undefined)
                  }}
                >
                  <span className="font-semibold">${preset.toLocaleString()}</span>
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
        {paymentMethodsLoading ? (
          <Skeleton className="h-4 w-64 max-w-full" />
        ) : showMissingPaymentMethodTopUpMessage ? (
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <CreditCardIcon className="w-4 h-4 shrink-0" />
            <span>Add a payment method to top up.</span>
          </div>
        ) : (
          <div className="text-sm text-muted-foreground">You will be redirected to Stripe to complete the payment.</div>
        )}
        <Button onClick={handleTopUpWallet} disabled={!topUpEnabled} size="sm">
          {topUpWalletMutation.isPending && <Spinner />}
          Top up
        </Button>
      </CardFooter>
    </Card>
  )
}
