/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationTier, Tier, TierLimit } from '@/billing-api'
import { RoutePath } from '@/enums/RoutePath'
import { cn } from '@/lib/utils'
import { CheckIcon, CpuIcon, ExternalLinkIcon, HardDriveIcon, Loader2, MemoryStickIcon } from 'lucide-react'
import React, { ReactNode, useState } from 'react'
import { Link } from 'react-router-dom'
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from './ui/accordion'
import { Badge } from './ui/badge'
import { Button } from './ui/button'
import { Skeleton } from './ui/skeleton'

interface Props {
  emailVerified: boolean
  githubConnected: boolean
  organizationTier?: OrganizationTier | null
  creditCardConnected: boolean
  phoneVerified: boolean
  tierLoading: boolean
  tiers: Tier[]
  tierFeatures?: Record<number, ReactNode>
  onUpgrade: (tier: number) => Promise<void>
  onDowngrade: (tier: number) => Promise<void>
}

export function TierAccordion({
  emailVerified,
  githubConnected,
  organizationTier,
  creditCardConnected: creditCardLinked,
  phoneVerified,
  tierLoading,
  tiers,
  tierFeatures,
  onUpgrade,
  onDowngrade,
}: Props) {
  const nextTiers = tiers.filter((t) => t.tier > (organizationTier?.tier || 0)).map((t) => String(t.tier))

  return (
    <>
      <Accordion type="multiple" className="w-full" defaultValue={nextTiers}>
        {tiers.map((tier) => {
          const isCurrentTier = organizationTier?.tier === tier.tier
          const topUpChecked =
            !!organizationTier &&
            organizationTier.largestSuccessfulPaymentCents >= tier.minTopUpAmountCents &&
            (!tier.topUpIntervalDays ||
              (!!organizationTier.largestSuccessfulPaymentDate &&
                organizationTier.largestSuccessfulPaymentDate.getTime() >
                  Date.now() - 1000 * 60 * 60 * 24 * tier.topUpIntervalDays))

          const isGreaterThanCurrentTier = organizationTier && tier.tier > organizationTier.tier

          const features = tierFeatures ? tierFeatures[tier.tier] : null

          const canUpgrade = canUpgradeToTier(
            organizationTier ?? null,
            tier,
            creditCardLinked,
            githubConnected,
            phoneVerified,
          )

          return (
            <AccordionItem
              value={String(tier.tier)}
              key={tier.tier}
              className={cn('border-b-0 relative border-t border-border py-5 px-4')}
            >
              <div className="grid items-center gap-2 w-full mb-4">
                <div className="flex items-center gap-2 flex-wrap">
                  <span
                    className={cn('font-semibold text-lg', {
                      'text-green-500': isCurrentTier,
                    })}
                  >
                    Tier {tier.tier}
                  </span>
                  {isCurrentTier && (
                    <Badge variant="outline" className="font-mono uppercase">
                      Current
                    </Badge>
                  )}
                  {organizationTier && (
                    <div className="flex items-center gap-1 ml-auto">
                      <TierActionButton
                        tier={tier.tier}
                        currentTier={organizationTier?.tier ?? 0}
                        canUpgrade={canUpgrade}
                        tierLoading={tierLoading}
                        tierExpiresAt={organizationTier?.expiresAt}
                        onUpgrade={onUpgrade}
                        onDowngrade={onDowngrade}
                      />
                    </div>
                  )}
                </div>
              </div>
              <AccordionTrigger
                className={cn(
                  'w-full text-left hover:no-underline items-end p-0 text-muted-foreground hover:text-foreground',
                )}
              >
                <div className="flex items-start flex-col gap-2 mr-4">
                  <span className="text-xs uppercase font-mono text-muted-foreground">Resources:</span>
                  <TierLimitsIndicator limit={tier.tierLimit} />
                </div>

                <div className="ml-auto mr-2 text-sm">Details</div>
              </AccordionTrigger>
              <AccordionContent className="border-t border-border border-dashed pt-4 pb-0 mt-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-2 pr-8 gap-y-5">
                  {features && (
                    <div className="flex flex-col gap-2">
                      <span className="text-xs uppercase font-mono text-muted-foreground">Additional Features:</span>
                      {features}
                    </div>
                  )}
                  <div className="flex flex-col gap-2">
                    <span className="text-xs uppercase font-mono text-muted-foreground">Requirements:</span>
                    <AdditionalTierRequirements
                      tier={tier}
                      emailVerified={emailVerified}
                      creditCardLinked={creditCardLinked}
                      githubConnected={githubConnected}
                      phoneVerified={phoneVerified}
                      businessEmailVerified={organizationTier?.hasVerifiedBusinessEmail ?? false}
                    />
                    {!!tier.minTopUpAmountCents && (
                      <div className={cn(topUpChecked ? 'text-foreground' : undefined)}>
                        <TierRequirementItem
                          checked={topUpChecked}
                          label={`Top up ${getDollarAmount(tier.minTopUpAmountCents)}${tier.topUpIntervalDays ? ` (every ${tier.topUpIntervalDays} days)` : ' (one time)'}`}
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
                    <div className="">
                      {isGreaterThanCurrentTier && !canUpgrade && (
                        <div className="text-xs text-muted-foreground">
                          Please complete all requirements to upgrade.
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </AccordionContent>
            </AccordionItem>
          )
        })}
      </Accordion>
      <div className="justify-between grid grid-cols-2 items-center border-t border-border px-4 py-5 gap-2">
        <span
          className={cn({
            'text-green-500': organizationTier?.tier && organizationTier.tier >= 3,
          })}
        >
          Custom
        </span>

        <Button variant="secondary" asChild className="w-fit justify-self-end">
          <a href="mailto:sales@daytona.io?subject=Custom%20Tier%20Inquiry&body=Hi%20Daytona%20Team%2C%0A%0AI%27m%20interested%20in%20a%20custom%20plan%20and%20would%20like%20to%20learn%20more%20about%20your%20options.%0A%0AHere%27s%20some%20context%3A%0A%0A-%20Your%20use%20case%3A%20%0A-%20Current%20technology%3A%20%0A-%20Requirements%3A%20%0A-%20Typical%20sandbox%20size%3A%20%0A-%20Peak%20concurrent%20sandboxes%3A%20%0A%0AThanks.">
            Contact Sales
          </a>
        </Button>
        <span className="text-sm text-muted-foreground col-[2] row-[2] justify-self-end">
          Tier 3 or higher required.
        </span>
      </div>
    </>
  )
}

export function TierAccordionItemSkeleton() {
  return (
    <div className="w-full flex flex-col gap-4">
      <Skeleton className="h-4 w-1/2" />
      <Skeleton className="h-12 w-full" />
    </div>
  )
}

export function TierAccordionSkeleton() {
  return (
    <div className="w-full flex flex-col gap-5">
      <TierAccordionItemSkeleton />
      <TierAccordionItemSkeleton />
      <TierAccordionItemSkeleton />
      <TierAccordionItemSkeleton />
    </div>
  )
}

function TierLimitResource({ label, value, icon }: { label: string; value: number | string; icon: React.ReactNode }) {
  return (
    <div className="flex items-end font-mono gap-2 flex-wrap justify-end">
      <div className="flex items-center gap-1">
        {icon}
        <div className="text-sm text-muted-foreground">{label}</div>{' '}
      </div>
      <div className="text-sm text-foreground">{value}</div>
    </div>
  )
}

function TierLimitsIndicator({
  limit,
  className,
  children,
}: {
  limit: TierLimit
  className?: string
  children?: React.ReactNode
}) {
  return (
    <div
      className={cn(
        'flex sm:items-center text-sm text-muted-foreground font-mono sm:flex-row flex-col items-start sm:gap-5 gap-2',
        className,
      )}
    >
      <TierLimitResource label="vCPU" value={limit.concurrentCPU} icon={<CpuIcon strokeWidth={1.5} size={16} />} />
      <TierLimitResource
        label="RAM"
        value={`${limit.concurrentRAMGiB} GiB`}
        icon={<MemoryStickIcon strokeWidth={1.5} size={16} />}
      />
      <TierLimitResource
        label="DISK"
        value={`${limit.concurrentDiskGiB} GiB`}
        icon={<HardDriveIcon strokeWidth={1.5} size={16} />}
      />
      {children}
    </div>
  )
}

function RequirementIcon({ checked, label }: { checked: boolean; label: string }) {
  return (
    <div
      className={cn(
        'flex-shrink-0 w-4 h-4 rounded-full flex items-center justify-center border border-muted-foreground/50',
        {
          'bg-muted/50 text-foreground': checked,
          'text-transparent': !checked,
        },
      )}
    >
      <CheckIcon size={10} aria-label={label} />
    </div>
  )
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
      return (
        <Button variant="secondary" disabled>
          Current
        </Button>
      )
    }

    return (
      <div className="text-sm text-foreground">
        Tier expires on{' '}
        {tierExpiresAt.toLocaleDateString('en-US', {
          month: 'short',
          day: 'numeric',
        })}
        .
      </div>
    )
  }

  if (tier === currentTier + 1) {
    return (
      <Button
        variant="default"
        onClick={() => {
          setTierActionLoading(true)
          onUpgrade(tier).finally(() => setTierActionLoading(false))
        }}
        disabled={tierActionLoading || !canUpgrade}
      >
        {tierActionLoading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
        Upgrade
      </Button>
    )
  }

  if (tier === currentTier - 1) {
    return (
      <Button
        variant="secondary"
        onClick={() => {
          setTierActionLoading(true)
          onDowngrade(tier).finally(() => setTierActionLoading(false))
        }}
        disabled={tierActionLoading}
      >
        {tierActionLoading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
        Downgrade
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
    return <TierRequirementItem checked={props.emailVerified} label="Email verification" />
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

function TierRequirementItem({ checked, label, link, externalLink }: TierRequirementItemProps) {
  const content = (
    <span className="flex items-center gap-2">
      <RequirementIcon checked={checked} label={label} />
      <span
        className={cn({
          'text-muted-foreground line-through': checked,
          'text-foreground': !checked,
          'hover:underline': !checked && link,
        })}
      >
        {label}
      </span>
      {!checked && externalLink && (
        <ExternalLinkIcon size={16} className="inline align-text-bottom" aria-label={label} />
      )}
    </span>
  )

  if (!checked && link) {
    return (
      <div
        className={cn({
          'text-foreground': checked,
          'text-muted-foreground': !checked,
          'hover:underline': !checked && link,
        })}
      >
        <Link to={link}>{content}</Link>
      </div>
    )
  }

  return <div className={cn(checked ? 'text-foreground' : 'text-muted-foreground')}>{content}</div>
}
