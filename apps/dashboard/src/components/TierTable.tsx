'use client'

/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RoutePath } from '@/enums/RoutePath'
import { Button } from './ui/button'
import { PhoneCall, CheckCircle, Circle, Loader2, ExternalLinkIcon, Cpu, HardDrive, MemoryStick } from 'lucide-react'
import { Link } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { Tooltip } from './Tooltip'
import type { OrganizationTier, Tier } from '@/billing-api'
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
  creditCardConnected,
  phoneVerified,
  tierLoading,
  tiers,
  onUpgrade,
  onDowngrade,
}: Props) {
  return (
    <div className="space-y-8">
      {/* Progressive tier layout */}
      <div className="relative">
        <div className="absolute left-1 top-3 bottom-8 w-px bg-gray-200 dark:bg-gray-700" />

        <div className="space-y-4">
          {tiers.map((tier, index) => {
            const topUpChecked =
              !!organizationTier &&
              organizationTier.largestSuccessfulPaymentCents >= tier.minTopUpAmountCents &&
              (!tier.topUpIntervalDays ||
                (!!organizationTier.largestSuccessfulPaymentDate &&
                  organizationTier.largestSuccessfulPaymentDate.getTime() >
                    Date.now() - 1000 * 60 * 60 * 24 * tier.topUpIntervalDays))

            const isCurrentTier = tier.tier === (organizationTier?.tier ?? 0)
            const isInactiveTier = tier.tier !== (organizationTier?.tier ?? 0)

            const canUpgrade = canUpgradeToTier(
              organizationTier,
              tier,
              creditCardConnected,
              githubConnected,
              phoneVerified,
            )
            const missingRequirements = getMissingRequirements(
              tier,
              emailVerified,
              creditCardConnected,
              githubConnected,
              phoneVerified,
              organizationTier?.hasVerifiedBusinessEmail ?? false,
              topUpChecked,
            )

            return (
              <div key={tier.tier} className="relative flex items-start gap-8">
                <div className="relative z-10 flex-shrink-0 mt-2">
                  <div
                    className={cn(
                      'w-2 h-2 rounded-full',
                      isCurrentTier ? 'bg-gray-600 dark:bg-gray-300' : 'bg-gray-300 dark:bg-gray-600',
                    )}
                  />
                </div>

                <div className="flex-1 min-w-0">
                  <div
                    className={cn(
                      'py-2',
                      isInactiveTier && 'opacity-60',
                      isCurrentTier &&
                        'bg-gray-50/10 dark:bg-gray-800/10 rounded-lg px-4 py-4 border border-gray-200/20 dark:border-gray-700/20',
                    )}
                  >
                    <div className="flex items-start justify-between mb-6">
                      <div>
                        <div className="flex items-center gap-3 mb-1">
                          <h3 className="text-lg font-semibold">Tier {tier.tier}</h3>
                        </div>
                        <div className="flex items-center gap-4 text-xs font-medium text-foreground bg-gray-100/50 dark:bg-gray-800/50 rounded-md px-3 py-2">
                          <div className="flex items-center gap-1.5">
                            <Cpu size={14} className="text-muted-foreground/60" />
                            <span>{tier.tierLimit.concurrentCPU} vCPU</span>
                          </div>
                          <div className="flex items-center gap-1.5">
                            <MemoryStick size={14} className="text-muted-foreground/60" />
                            <span>{tier.tierLimit.concurrentRAMGiB} GiB</span>
                          </div>
                          <div className="flex items-center gap-1.5">
                            <HardDrive size={14} className="text-muted-foreground/60" />
                            <span>{tier.tierLimit.concurrentDiskGiB} GiB</span>
                          </div>
                        </div>
                      </div>

                      <div className="flex-shrink-0">
                        <TierActionButton
                          tier={tier.tier}
                          currentTier={organizationTier?.tier ?? 0}
                          canUpgrade={canUpgrade}
                          tierLoading={tierLoading}
                          tierExpiresAt={organizationTier?.expiresAt}
                          missingRequirements={missingRequirements}
                          onUpgrade={onUpgrade}
                          onDowngrade={onDowngrade}
                        />
                      </div>
                    </div>

                    <div className="space-y-2">
                      <div className="text-xs font-medium text-muted-foreground mb-2">Requirements</div>
                      <div className="space-y-1.5 mt-3">
                        <AdditionalTierRequirements
                          tier={tier}
                          emailVerified={emailVerified}
                          creditCardLinked={creditCardConnected}
                          githubConnected={githubConnected}
                          phoneVerified={phoneVerified}
                          businessEmailVerified={organizationTier?.hasVerifiedBusinessEmail ?? false}
                        />
                        {!!tier.minTopUpAmountCents && (
                          <div className="text-sm text-muted-foreground">
                            <TierRequirementItem
                              checked={topUpChecked}
                              label={`Top up ${getDollarAmount(tier.minTopUpAmountCents)}${tier.topUpIntervalDays ? ` (every ${tier.topUpIntervalDays} days)` : ' (one time)'}`}
                              link={RoutePath.BILLING_WALLET}
                            />
                            {!!tier.topUpIntervalDays && (
                              <div className="ml-5 mt-1 text-xs text-muted-foreground">
                                {organizationTier &&
                                  organizationTier.largestSuccessfulPaymentDate &&
                                  organizationTier.largestSuccessfulPaymentCents >= tier.minTopUpAmountCents &&
                                  `Latest: ${organizationTier.largestSuccessfulPaymentDate.toLocaleDateString('en-US', {
                                    month: 'short',
                                    day: 'numeric',
                                  })}`}
                              </div>
                            )}
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            )
          })}

          {/* Custom tier */}
          <div className="relative flex items-start gap-8 opacity-40">
            <div className="relative z-10 flex-shrink-0 mt-2">
              <div className="w-2 h-2 rounded-full bg-gray-300 dark:bg-gray-600" />
            </div>

            <div className="flex-1 min-w-0 py-2">
              <div className="flex items-start justify-between mb-3">
                <div>
                  <h3 className="text-lg font-medium mb-1">Custom</h3>
                  <div className="text-sm text-muted-foreground">Custom limits based on your needs</div>
                </div>
                <Button variant="outline" size="sm">
                  <a href="mailto:sales@daytona.io?subject=Custom%20Tier%20Inquiry&body=Hi%20Daytona%20Team%2C%0A%0AI%27m%20interested%20in%20a%20custom%20plan%20and%20would%20like%20to%20learn%20more%20about%20your%20options.%0A%0AHere%27s%20some%20context%3A%0A%0A-%20Your%20use%20case%3A%20%0A-%20Current%20technology%3A%20%0A-%20Requirements%3A%20%0A-%20Typical%20sandbox%20size%3A%20%0A-%20Peak%20concurrent%20sandboxes%3A%20%0A%0AThanks.">
                    Contact Sales
                  </a>
                </Button>
              </div>

              <div className="space-y-2">
                <div className="text-xs font-medium text-muted-foreground mb-2">Requirements</div>
                <div className="space-y-1.5">
                  <div className="text-sm text-muted-foreground">
                    {getIcon(!!organizationTier?.tier && organizationTier.tier >= 3, 'At least Tier 3')}At least Tier 3
                  </div>
                  <div className="text-sm text-muted-foreground">
                    <PhoneCall size={14} className="inline align-text-bottom mr-2" aria-label="Contact sales" />
                    Contact sales at sales@daytona.io
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function getIcon(checked: boolean, label: string) {
  if (checked) {
    return (
      <CheckCircle
        size={14}
        className="inline align-text-bottom mr-2 text-green-600 dark:text-green-500"
        aria-label={label}
      />
    )
  }
  return <Circle size={14} className="inline align-text-bottom mr-2 text-muted-foreground" aria-label={label} />
}

function canUpgradeToTier(
  organizationTier: OrganizationTier | null,
  tier: Tier,
  creditCardConnected: boolean,
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
      return creditCardConnected && githubConnected
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
  missingRequirements: string[]
  onUpgrade: (tier: number) => Promise<void>
  onDowngrade: (tier: number) => Promise<void>
}

function TierActionButton({
  tier,
  currentTier,
  canUpgrade,
  tierLoading,
  tierExpiresAt,
  missingRequirements,
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
      <div className="text-sm text-muted-foreground">
        Expires on{' '}
        {tierExpiresAt.toLocaleDateString('en-US', {
          month: 'short',
          day: 'numeric',
        })}
      </div>
    )
  }

  if (tier === currentTier + 1) {
    const upgradeButton = (
      <Button
        variant="outline"
        size="sm"
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

    if (!canUpgrade) {
      return (
        <Tooltip
          label={<div>{upgradeButton}</div>}
          content={
            <div className="max-w-60">
              <div className="font-medium mb-1 text-red-400 text-xs">Requirements not met:</div>
              <div className="space-y-0.5">
                {missingRequirements.map((req, idx) => (
                  <div key={idx} className="text-xs flex items-center gap-1.5">
                    <Circle size={10} className="flex-shrink-0 text-red-400" />
                    {req}
                  </div>
                ))}
              </div>
            </div>
          }
        />
      )
    }

    return upgradeButton
  }

  if (tier === currentTier - 1) {
    return (
      <Button
        variant="outline"
        size="sm"
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
      <div
        className={cn('text-sm', props.emailVerified ? 'text-green-600 dark:text-green-500' : 'text-muted-foreground')}
      >
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
      {!checked && link && <ExternalLinkIcon size={12} className="inline align-text-bottom ml-1" aria-label={label} />}
    </>
  )

  if (!checked && link) {
    return (
      <div className="text-sm text-muted-foreground">
        <Link to={link}>{content}</Link>
      </div>
    )
  }

  return (
    <div className={cn('text-sm', checked ? 'text-green-600 dark:text-green-500' : 'text-muted-foreground')}>
      {content}
    </div>
  )
}

function getMissingRequirements(
  tier: Tier,
  emailVerified: boolean,
  creditCardConnected: boolean,
  githubConnected: boolean,
  phoneVerified: boolean,
  businessEmailVerified: boolean,
  topUpChecked: boolean,
): string[] {
  const missing: string[] = []

  if (tier.tier === 1 && !emailVerified) {
    missing.push('Email verification')
  }

  if (tier.tier === 2) {
    if (!creditCardConnected) missing.push('Credit card linked')
    if (!githubConnected) missing.push('GitHub connected')
  }

  if (tier.tier === 3) {
    if (!businessEmailVerified) missing.push('Business email verified')
    if (!phoneVerified) missing.push('Phone verified')
  }

  if (tier.minTopUpAmountCents && !topUpChecked) {
    missing.push(`Top up ${getDollarAmount(tier.minTopUpAmountCents)}`)
  }

  return missing
}

function getDollarAmount(cents: number) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(cents / 100)
}
