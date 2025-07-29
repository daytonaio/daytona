/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RoutePath } from '@/enums/RoutePath'
import { Button } from './ui/button'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from './ui/table'
import { PhoneCall, CheckCircle, Circle, Info, Loader2, ExternalLinkIcon } from 'lucide-react'
import { Link } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { Tooltip } from './Tooltip'
import { OrganizationTier, Tier } from '@/billing-api'
import { useState } from 'react'

type Props = {
  emailVerified: boolean
  githubConnected: boolean
  organizationTier: OrganizationTier | null
  creditCardConnected: boolean
  phoneVerified: boolean
  tierLoading: boolean
  tiers: Tier[]
  onUpgrade: (tier: number) => Promise<void>
  onDowngrade: (tier: number) => Promise<void>
}

export function TierTable({
  emailVerified,
  githubConnected,
  organizationTier,
  creditCardConnected: creditCardLinked,
  phoneVerified,
  tierLoading,
  tiers,
  onUpgrade,
  onDowngrade,
}: Props) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Tier</TableHead>
          <TableHead className="cursor-pointer">
            <Tooltip
              label={
                <div className="flex items-center gap-2 max-w-80">
                  <Info size={16} />
                  Available Compute (vCPU / RAM / Storage)
                </div>
              }
              content={
                <div className="max-w-80">
                  Total vCPU, RAM, and Storage available at any moment across all running sandboxes.
                  <br />
                  The number of concurrent sandboxes depends on how much compute each one uses.
                </div>
              }
            />
          </TableHead>
          <TableHead>Access Verification</TableHead>
          <TableHead></TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {tiers.map((tier) => {
          const topUpChecked =
            !!organizationTier &&
            organizationTier.largestSuccessfulPaymentCents >= tier.minTopUpAmountCents &&
            (!tier.topUpIntervalDays ||
              (!!organizationTier.largestSuccessfulPaymentDate &&
                organizationTier.largestSuccessfulPaymentDate.getTime() >
                  Date.now() - 1000 * 60 * 60 * 24 * tier.topUpIntervalDays))

          return (
            <TableRow key={tier.tier}>
              <TableCell>
                <b>Tier {tier.tier}</b>
              </TableCell>
              <TableCell>
                {tier.tierLimit.concurrentCPU} vCPU / {tier.tierLimit.concurrentRAMGiB} GiB /{' '}
                {tier.tierLimit.concurrentDiskGiB} GiB
              </TableCell>
              <TableCell>
                <div className="grid grid-cols-1 gap-0 gap-y-2 py-2 [&>*]:flex [&>*]:items-center [&>*]:flex-wrap">
                  <AdditionalTierRequirements
                    tier={tier}
                    emailVerified={emailVerified}
                    creditCardLinked={creditCardLinked}
                    githubConnected={githubConnected}
                    phoneVerified={phoneVerified}
                    businessEmailVerified={organizationTier?.hasVerifiedBusinessEmail ?? false}
                  />
                  {!!tier.minTopUpAmountCents && (
                    <div className={cn(topUpChecked ? 'text-green-500' : undefined)}>
                      <TierRequirementItem
                        checked={topUpChecked}
                        label={`Top up ${getDollarAmount(tier.minTopUpAmountCents)}${tier.topUpIntervalDays ? ` (every ${tier.topUpIntervalDays} days)` : ''}`}
                        link={RoutePath.BILLING_WALLET}
                      />
                      {!!tier.topUpIntervalDays && (
                        <div className="basis-full ml-6">
                          {organizationTier &&
                            organizationTier.largestSuccessfulPaymentDate &&
                            organizationTier.largestSuccessfulPaymentCents >= tier.minTopUpAmountCents &&
                            ` (latest top-up: ${organizationTier.largestSuccessfulPaymentDate.toLocaleDateString(
                              'en-US',
                              {
                                month: 'short',
                                day: 'numeric',
                              },
                            )})`}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              </TableCell>
              <TableCell className="text-center">
                <TierActionButton
                  tier={tier.tier}
                  currentTier={organizationTier?.tier ?? 0}
                  canUpgrade={canUpgradeToTier(
                    organizationTier,
                    tier,
                    creditCardLinked,
                    githubConnected,
                    phoneVerified,
                  )}
                  tierLoading={tierLoading}
                  tierExpiresAt={organizationTier?.expiresAt}
                  onUpgrade={onUpgrade}
                  onDowngrade={onDowngrade}
                />
              </TableCell>
            </TableRow>
          )
        })}
        <TableRow>
          <TableCell>
            <b>Custom</b>
          </TableCell>
          <TableCell>Custom</TableCell>
          <TableCell>
            <div className="grid gap-0 gap-y-4 py-2">
              <div>
                <PhoneCall size={18} className="inline align-text-bottom mr-2" aria-label="Contact sales" />
                Contact sales at sales@daytona.io
              </div>
            </div>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  )
}

function getIcon(checked: boolean, label: string) {
  if (checked) {
    return <CheckCircle size={18} className="inline align-text-bottom mr-2" aria-label={label} />
  }
  return <Circle size={18} className="inline align-text-bottom mr-2" aria-label={label} />
}

function getDollarAmount(cents: number) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(cents / 100)
}

function canUpgradeToTier(
  organizationTier: OrganizationTier | null,
  tier: Tier,
  creditCardLinked: boolean,
  githubConnected: boolean,
  phoneVerified: boolean,
) {
  if (!organizationTier || tier.tier <= 1) {
    return false
  }

  if (!organizationTier.largestSuccessfulPaymentDate) {
    return false
  }

  if (organizationTier.largestSuccessfulPaymentCents < tier.minTopUpAmountCents) {
    return false
  }

  if (tier.topUpIntervalDays) {
    const diffTime = Math.abs(Date.now() - organizationTier.largestSuccessfulPaymentDate.getTime())
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))

    return diffDays < tier.topUpIntervalDays
  }

  switch (tier.tier) {
    case 2:
      return creditCardLinked && githubConnected
    case 3:
      return organizationTier.hasVerifiedBusinessEmail && phoneVerified
  }

  return true
}

type TierActionButtonProps = {
  tier: number
  currentTier: number
  canUpgrade: boolean
  tierLoading: boolean
  tierExpiresAt?: Date
  onUpgrade: (tier: number) => Promise<void>
  onDowngrade: (tier: number) => Promise<void>
}

function TierActionButton({
  tier,
  currentTier,
  canUpgrade,
  tierLoading,
  tierExpiresAt,
  onUpgrade,
  onDowngrade,
}: TierActionButtonProps) {
  const [tierActionLoading, setTierActionLoading] = useState<boolean>(false)

  if (tierLoading) {
    return <div></div>
  }

  if (tier === currentTier) {
    if (!tierExpiresAt) {
      return <div></div>
    }

    return (
      <div>
        Expires on{' '}
        {tierExpiresAt.toLocaleDateString('en-US', {
          month: 'short',
          day: 'numeric',
        })}
      </div>
    )
  }

  if (tier === currentTier + 1) {
    return (
      <Button
        variant="outline"
        onClick={() => {
          setTierActionLoading(true)
          onUpgrade(tier).finally(() => setTierActionLoading(false))
        }}
        disabled={!canUpgrade || tierActionLoading}
      >
        {tierActionLoading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
        Upgrade to Tier {tier}
      </Button>
    )
  }

  if (tier === currentTier - 1) {
    return (
      <Button
        variant="outline"
        onClick={() => {
          setTierActionLoading(true)
          onDowngrade(tier).finally(() => setTierActionLoading(false))
        }}
        disabled={tierActionLoading}
      >
        {tierActionLoading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
        Downgrade to Tier {tier}
      </Button>
    )
  }

  return null
}

type AdditionalTierRequirementsProps = {
  tier: Tier
  emailVerified: boolean
  creditCardLinked: boolean
  githubConnected: boolean
  businessEmailVerified: boolean
  phoneVerified: boolean
}

function AdditionalTierRequirements({ tier, ...props }: AdditionalTierRequirementsProps) {
  if (tier.tier < 1 || tier.tier > 3) {
    return null
  }

  if (tier.tier === 1) {
    return (
      <div className={cn(props.emailVerified ? 'text-green-500' : undefined)}>
        {getIcon(props.emailVerified, 'Email verification')}Email verification
      </div>
    )
  }

  if (tier.tier === 2) {
    return (
      <>
        <TierRequirementItem
          checked={props.creditCardLinked}
          label="Credit card linked"
          link={RoutePath.BILLING_WALLET}
        />
        <TierRequirementItem
          checked={props.githubConnected}
          label="GitHub connected"
          link={RoutePath.ACCOUNT_SETTINGS}
        />
      </>
    )
  }

  if (tier.tier === 3) {
    return (
      <>
        <TierRequirementItem
          checked={props.businessEmailVerified}
          label="Business email verified"
          link={RoutePath.BILLING_WALLET}
        />
        <TierRequirementItem checked={props.phoneVerified} label="Phone verified" link={RoutePath.ACCOUNT_SETTINGS} />
      </>
    )
  }

  return null
}

interface TierRequirementItemProps {
  checked: boolean
  label: string
  link?: string
  externalLink?: boolean
}

function TierRequirementItem({ checked, label, link }: TierRequirementItemProps) {
  const content = (
    <>
      {getIcon(checked, label)}
      {label}
      {!checked && link && <ExternalLinkIcon size={16} className="inline align-text-bottom ml-1" aria-label={label} />}
    </>
  )

  if (!checked && link) {
    return (
      <div className={cn(checked ? 'text-green-500' : undefined)}>
        <Link to={link}>{content}</Link>
      </div>
    )
  }

  return <div className={cn(checked ? 'text-green-500' : undefined)}>{content}</div>
}
