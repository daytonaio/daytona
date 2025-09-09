/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationTier, Tier, TierLimit } from '@/billing-api'
import { RoutePath } from '@/enums/RoutePath'
import { cn } from '@/lib/utils'
import { CheckIcon, ExternalLinkIcon, Info, Loader2, MinusIcon } from 'lucide-react'
import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Tooltip } from './Tooltip'
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from './ui/accordion'
import { Badge } from './ui/badge'
import { Button } from './ui/button'
import { Skeleton } from './ui/skeleton'

type Props = {
  emailVerified: boolean
  githubConnected: boolean
  organizationTier?: OrganizationTier | null
  creditCardConnected: boolean
  phoneVerified: boolean
  tierLoading: boolean
  tiers: Tier[]
  onUpgrade: (tier: number) => Promise<void>
  onDowngrade: (tier: number) => Promise<void>
}

export function TierAccordionSkeleton() {
  return (
    <div className="w-full flex flex-col gap-4">
      <Skeleton className="h-12 w-full" />
      <Skeleton className="h-12 w-full" />
      <Skeleton className="h-12 w-full" />
      <Skeleton className="h-12 w-full" />
      <Skeleton className="h-12 w-full" />
    </div>
  )
}

export function TierAccordion({
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
  const nextTier = useMemo(() => {
    return organizationTier?.tier ? organizationTier.tier + 1 : 1
  }, [organizationTier?.tier])

  return (
    <>
      <Accordion type="multiple" className="w-full" defaultValue={[String(nextTier)]}>
        {tiers.map((tier) => {
          const isCurrentTier = organizationTier?.tier === tier.tier
          const topUpChecked =
            !!organizationTier &&
            organizationTier.largestSuccessfulPaymentCents >= tier.minTopUpAmountCents &&
            (!tier.topUpIntervalDays ||
              (!!organizationTier.largestSuccessfulPaymentDate &&
                organizationTier.largestSuccessfulPaymentDate.getTime() >
                  Date.now() - 1000 * 60 * 60 * 24 * tier.topUpIntervalDays))

          const isLessThanCurrentTier = organizationTier && tier.tier < organizationTier.tier

          return (
            <AccordionItem value={String(tier.tier)} key={tier.tier}>
              <AccordionTrigger
                className={cn('w-full text-left hover:no-underline', {
                  '[&:not([data-state=open])]:opacity-40': isLessThanCurrentTier,
                })}
              >
                <div className="flex items-start md:items-center gap-2 w-full pr-3 md:flex-row flex-col">
                  <div>
                    <span
                      className={cn('font-mono uppercase', {
                        'text-green-500': isCurrentTier,
                      })}
                    >
                      Tier {tier.tier}
                    </span>{' '}
                    {isCurrentTier && (
                      <Badge variant="outline" className="font-mono uppercase">
                        Current
                      </Badge>
                    )}
                  </div>

                  <TierLimitsBadge limit={tier.tierLimit} className="md:ml-auto" />
                </div>
              </AccordionTrigger>
              <AccordionContent>
                <div className="flex items-end justify-between pr-8">
                  <div className="flex flex-col gap-2">
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
                  </div>

                  {organizationTier && (
                    <div className="flex items-center gap-2">
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
                    </div>
                  )}
                </div>
              </AccordionContent>
            </AccordionItem>
          )
        })}
      </Accordion>
      <div className="flex items-end justify-between pr-8 py-4">
        <span
          className={cn('font-mono uppercase', {
            'text-green-500': organizationTier?.tier && organizationTier.tier >= 3,
          })}
        >
          Custom
        </span>

        <Button variant="outline" asChild>
          <a href="mailto:sales@daytona.io?subject=Custom%20Tier%20Inquiry&body=Hi%20Daytona%20Team%2C%0A%0AI%27m%20interested%20in%20a%20custom%20plan%20and%20would%20like%20to%20learn%20more%20about%20your%20options.%0A%0AHere%27s%20some%20context%3A%0A%0A-%20Your%20use%20case%3A%20%0A-%20Current%20technology%3A%20%0A-%20Requirements%3A%20%0A-%20Typical%20sandbox%20size%3A%20%0A-%20Peak%20concurrent%20sandboxes%3A%20%0A%0AThanks.">
            Contact Sales
          </a>
        </Button>
      </div>
    </>
  )
}

function TierLimitResource({ label, value }: { label: string; value: number | string }) {
  return (
    <div className="flex items-center font-mono [&+&]:border-l border-border px-3 py-2 gap-2">
      <div className="text-sm text-muted-foreground">{label}</div>{' '}
      <div className="text-sm text-foreground">{value}</div>
    </div>
  )
}

function TierLimitsBadge({ limit, className }: { limit: TierLimit; className?: string }) {
  return (
    <div
      className={cn(
        'flex items-center text-sm text-muted-foreground rounded-md border border-border font-mono',
        className,
      )}
    >
      <TierLimitResource label="vCPU" value={limit.concurrentCPU} />
      <TierLimitResource label="RAM" value={`${limit.concurrentRAMGiB} GiB`} />
      <TierLimitResource label="DISK" value={`${limit.concurrentDiskGiB} GiB`} />
    </div>
  )
}

function getIcon(checked: boolean, label: string) {
  if (checked) {
    return <CheckIcon size={18} className="inline align-text-bottom mr-2" aria-label={label} />
  }
  return <MinusIcon size={18} className="inline align-text-bottom mr-2" aria-label={label} />
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
      <>
        <Tooltip
          label={
            <button>
              <Info size={16} className="text-muted-foreground" />
            </button>
          }
          content={<div className="max-w-80">Complete all requirements to upgrade.</div>}
        />

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
      </>
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
      <div
        className={cn('text-muted-foreground', {
          'text-foreground': props.emailVerified,
        })}
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
      {!checked && link && <ExternalLinkIcon size={16} className="inline align-text-bottom ml-1" aria-label={label} />}
    </>
  )

  if (!checked && link) {
    return (
      <div className={cn(checked ? 'text-foreground' : 'text-muted-foreground')}>
        <Link to={link}>{content}</Link>
      </div>
    )
  }

  return <div className={cn(checked ? 'text-foreground' : 'text-muted-foreground')}>{content}</div>
}
