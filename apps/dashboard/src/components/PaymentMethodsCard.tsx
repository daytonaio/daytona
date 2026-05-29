/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { useOrganizationBillingPortalUrlQuery } from '@/hooks/queries/useOrganizationBillingPortalUrlQuery'
import { usePaymentMethodsQuery } from '@/hooks/queries/usePaymentMethodsQuery'
import { useSetupCheckoutUrlQuery } from '@/hooks/queries/useSetupCheckoutUrlQuery'
import { PaymentMethod } from '@daytona/billing-api-client'
import { CreditCardIcon, PencilIcon, PlusIcon } from 'lucide-react'
import { useCallback } from 'react'
import { toast } from 'sonner'

interface PaymentMethodsCardProps {
  organizationId: string
  creditCardConnectedCreditsGranted?: boolean
}

export function PaymentMethodsCard({ organizationId, creditCardConnectedCreditsGranted }: PaymentMethodsCardProps) {
  const paymentMethodsQuery = usePaymentMethodsQuery({ organizationId })
  const portalUrlQuery = useOrganizationBillingPortalUrlQuery({ organizationId })
  const fetchSetupCheckoutUrl = useSetupCheckoutUrlQuery(organizationId)
  const methods = paymentMethodsQuery.data ?? []
  const portalUrl = portalUrlQuery.data
  const showAddButton = methods.length === 0 && !creditCardConnectedCreditsGranted

  const handleAddPaymentMethod = useCallback(async () => {
    const newWindow = window.open('', '_blank')
    // Sever the opener link so the checkout page can't reach back into this tab.
    if (newWindow) {
      newWindow.opener = null
    }
    try {
      const url = await fetchSetupCheckoutUrl()
      if (newWindow) {
        newWindow.location.href = url
      } else {
        // Popup was blocked — fall back to navigating the current tab.
        window.location.href = url
      }
    } catch (error) {
      newWindow?.close()
      toast.error('Failed to fetch payment method setup url', {
        description: String(error),
      })
    }
  }, [fetchSetupCheckoutUrl])

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4 space-y-0">
        <div className="flex flex-col gap-1.5">
          <CardTitle>Payment methods</CardTitle>
          <CardDescription>Cards on file. The default is used for automatic charges and invoices.</CardDescription>
        </div>
        {methods.length > 0 &&
          (portalUrl ? (
            <Button variant="secondary" size="sm" asChild>
              <a href={portalUrl} target="_blank" rel="noopener noreferrer">
                <PencilIcon />
                Edit
              </a>
            </Button>
          ) : (
            <Button variant="secondary" size="sm" disabled>
              <PencilIcon />
              Edit
            </Button>
          ))}
      </CardHeader>

      <CardContent className="border-t border-border p-0">
        {paymentMethodsQuery.isLoading ? (
          <div className="p-4 flex flex-col gap-3">
            {Array.from({ length: 2 }).map((_, i) => (
              <Skeleton key={i} className="h-8 w-full" />
            ))}
          </div>
        ) : methods.length === 0 ? (
          <div className="p-4 flex items-center justify-between">
            <p className="text-sm text-muted-foreground">No cards on file yet.</p>
            {showAddButton && (
              <Button variant="secondary" size="sm" onClick={handleAddPaymentMethod}>
                <PlusIcon />
                Add payment method
              </Button>
            )}
          </div>
        ) : (
          <ul className="divide-y divide-border">
            {methods.map((method, index) => (
              <li key={method.id ?? `method-${index}`} className="flex items-center justify-between gap-4 p-4">
                <PaymentMethodRow method={method} />
              </li>
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  )
}

function PaymentMethodRow({ method }: { method: PaymentMethod }) {
  return (
    <div className="flex items-center gap-3 min-w-0">
      <CreditCardIcon className="w-4 h-4 text-muted-foreground shrink-0" />
      <div className="flex items-center gap-2 flex-wrap min-w-0">
        <span className="text-sm font-medium capitalize truncate">{method.brand ?? 'Card'}</span>
        {method.last4 && <span className="text-sm text-muted-foreground truncate">•••• {method.last4}</span>}
        {method.expMonth && method.expYear && (
          <span className="text-xs text-muted-foreground">
            Exp {String(method.expMonth).padStart(2, '0')}/{String(method.expYear).slice(-2)}
          </span>
        )}
        {method.isDefault && (
          <Badge variant="secondary" className="uppercase text-[10px]">
            Default
          </Badge>
        )}
      </div>
    </div>
  )
}
