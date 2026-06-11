/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { InputGroup, InputGroupAddon, InputGroupInput, InputGroupText } from '@/components/ui/input-group'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { useSetAutomaticTopUpMutation } from '@/hooks/mutations/useSetAutomaticTopUpMutation'
import { usePaymentMethodsQuery } from '@/hooks/queries/usePaymentMethodsQuery'
import type { AutomaticTopUp, OrganizationWallet } from '@daytona/billing-api-client'
import { InfoIcon } from 'lucide-react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { NumericFormat } from 'react-number-format'
import { toast } from 'sonner'

interface AutomaticTopUpCardProps {
  organizationId: string
  wallet: OrganizationWallet
}

export function AutomaticTopUpCard({ organizationId, wallet }: AutomaticTopUpCardProps) {
  const [automaticTopUp, setAutomaticTopUp] = useState<AutomaticTopUp | undefined>(undefined)
  const paymentMethodsQuery = usePaymentMethodsQuery({ organizationId })
  const setAutomaticTopUpMutation = useSetAutomaticTopUpMutation()
  const paymentMethods = paymentMethodsQuery.data
  const paymentMethodsLoading = paymentMethodsQuery.isLoading
  const hasNoPaymentMethod = (paymentMethods?.length ?? 0) === 0

  useEffect(() => {
    setAutomaticTopUp(wallet.automaticTopUp)
  }, [wallet.automaticTopUp])

  const automaticTopUpHasChanges = useMemo(() => {
    if (wallet.automaticTopUp?.disabled && (automaticTopUp?.thresholdAmount || 0) > 0) {
      return true
    }

    if (automaticTopUp?.thresholdAmount !== wallet.automaticTopUp?.thresholdAmount) {
      if (!wallet.automaticTopUp) {
        if ((automaticTopUp?.thresholdAmount || 0) !== 0) {
          return true
        }
      } else {
        return true
      }
    }

    if (automaticTopUp?.targetAmount !== wallet.automaticTopUp?.targetAmount) {
      if (!wallet.automaticTopUp) {
        if ((automaticTopUp?.targetAmount || 0) !== 0) {
          return true
        }
      } else {
        return true
      }
    }

    return false
  }, [wallet.automaticTopUp, automaticTopUp])

  const handleSetAutomaticTopUp = useCallback(async () => {
    try {
      await setAutomaticTopUpMutation.mutateAsync({
        organizationId,
        automaticTopUp,
      })
      toast.success('Automatic top up set successfully')
    } catch (error) {
      toast.error('Failed to set automatic top up', {
        description: String(error),
      })
    }
  }, [organizationId, automaticTopUp, setAutomaticTopUpMutation])

  const saveDisabled =
    !automaticTopUpHasChanges || setAutomaticTopUpMutation.isPending || paymentMethodsLoading || hasNoPaymentMethod

  return (
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
          <InfoIcon className="w-4 h-4 shrink-0" />
          <span className="text-sm">Setting both values to 0 will disable automatic top-ups.</span>
        </div>
        <div className="flex gap-2 items-center ml-auto">
          <Button onClick={handleSetAutomaticTopUp} disabled={saveDisabled}>
            {setAutomaticTopUpMutation.isPending && <Spinner />}
            Save
          </Button>
        </div>
      </CardFooter>
    </Card>
  )
}
