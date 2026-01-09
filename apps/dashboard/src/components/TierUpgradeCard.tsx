/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OrganizationTier, Tier } from '@/billing-api'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { RoutePath } from '@/enums/RoutePath'
import { useDowngradeTierMutation } from '@/hooks/mutations/useDowngradeTierMutation'
import { useUpgradeTierMutation } from '@/hooks/mutations/useUpgradeTierMutation'
import { handleApiError } from '@/lib/error-handling'
import { cn } from '@/lib/utils'
import { Organization } from '@daytonaio/api-client/src'
import { CheckIcon, ExternalLinkIcon, Loader2 } from 'lucide-react'
import { useMemo } from 'react'
import { Link } from 'react-router-dom'
import { toast } from 'sonner'

interface Props {
  tiers: Tier[]
  organizationTier?: OrganizationTier | null
  organization: Organization
  requirementsState: {
    emailVerified: boolean
    creditCardLinked: boolean
    businessEmailVerified: boolean
    githubConnected: boolean
  }
}

export function TierUpgradeCard({ tiers, organizationTier, requirementsState, organization }: Props) {
  const { currentTier, previousTier, nextTier } = useMemo(() => {
    const targetTiers: { currentTier?: Tier; previousTier?: Tier; nextTier?: Tier } = {}
    for (const tier of tiers) {
      if (tier.tier === organizationTier?.tier) {
        targetTiers.currentTier = tier
      }
      if (tier.tier < (organizationTier?.tier || 0)) {
        targetTiers.previousTier = tier
      }
      if (tier.tier > (organizationTier?.tier || 0) && !targetTiers.nextTier) {
        targetTiers.nextTier = tier
      }
    }
    return targetTiers
  }, [tiers, organizationTier])

  const requirements = getTierRequirementItems(requirementsState, organizationTier, nextTier)

  const canUpgrade = requirements.length > 0 && requirements.every((requirement) => requirement.isChecked)

  const downgradeTier = useDowngradeTierMutation()
  const upgradeTier = useUpgradeTierMutation()

  const handleUpgradeTier = async (tier: number) => {
    if (!organization) {
      return
    }

    try {
      await upgradeTier.mutateAsync({ organizationId: organization.id, tier })
      toast.success('Tier upgraded successfully')
    } catch (error) {
      handleApiError(error, 'Failed to upgrade organization tier')
    }
  }

  const handleDowngradeTier = async (tier: number) => {
    if (!organization) {
      return
    }

    try {
      await downgradeTier.mutateAsync({ organizationId: organization.id, tier })
      toast.success('Tier downgraded successfully')
    } catch (error) {
      handleApiError(error, 'Failed to downgrade organization tier')
    }
  }

  return (
    <Card>
      <CardContent className="p-0">
        {nextTier && (
          <div className="grid sm:grid-cols-2 grid-cols-1">
            <div className="p-4 flex flex-col gap-1">
              <div className="text-lg font-medium">Upgrade to Tier {nextTier?.tier}</div>
              <div className="text-muted-foreground text-sm">
                Unlock more resources and higher rate limits by completing the verification steps.
              </div>
            </div>
            <div className="sm:border-l border-border p-4 flex flex-col gap-2">
              <div className="text-xs text-muted-foreground">Requirements</div>
              <ul>
                {requirements.map((requirement) => (
                  <li key={requirement.label}>
                    <TierRequirementItem
                      checked={requirement.isChecked}
                      label={requirement.label}
                      link={requirement.link}
                    />
                  </li>
                ))}
              </ul>
              {requirements.length && !canUpgrade && (
                <div className="text-xs text-muted-foreground">Please complete all requirements to upgrade.</div>
              )}
              <Button
                className="w-full mt-4"
                onClick={() => handleUpgradeTier(nextTier.tier)}
                disabled={!canUpgrade || upgradeTier.isPending}
              >
                {upgradeTier.isPending && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                Upgrade
              </Button>
            </div>
          </div>
        )}
        <div className="p-4 border-t border-border flex items-center justify-between gap-2">
          <div className="flex flex-col gap-1 text-sm">
            <div className="font-medium">Enterprise</div>
            <div className="text-muted-foreground">
              Contact sales at{' '}
              <a href="mailto:sales@daytona.io" className="hover:text-foreground underline">
                sales@daytona.io
              </a>
              .
            </div>
          </div>

          <Button variant={organizationTier?.tier && organizationTier.tier > 2 ? 'default' : 'secondary'} asChild>
            <a href="mailto:sales@daytona.io?subject=Custom%20Tier%20Inquiry&body=Hi%20Daytona%20Team%2C%0A%0AI%27m%20interested%20in%20a%20custom%20plan%20and%20would%20like%20to%20learn%20more%20about%20your%20options.%0A%0AHere%27s%20some%20context%3A%0A%0A-%20Your%20use%20case%3A%20%0A-%20Current%20technology%3A%20%0A-%20Requirements%3A%20%0A-%20Typical%20sandbox%20size%3A%20%0A-%20Peak%20concurrent%20sandboxes%3A%20%0A%0AThanks.">
              Contact Sales
            </a>
          </Button>
        </div>
        {organizationTier && (
          <div className="border-t border-border p-4 flex items-center justify-between gap-2">
            <div className="flex flex-col gap-1 text-sm">
              <div className="font-medium">Current Tier: {organizationTier?.tier}</div>
              <div className="text-muted-foreground empty:hidden flex flex-col">
                {organizationTier.expiresAt && (
                  <div>
                    Tier expires on{' '}
                    {organizationTier.expiresAt.toLocaleDateString('en-US', {
                      month: 'short',
                      day: 'numeric',
                    })}
                    .
                  </div>
                )}
                {currentTier && currentTier?.topUpIntervalDays > 0 && (
                  <div>
                    Automatically charged {getDollarAmount(currentTier.minTopUpAmountCents)} every{' '}
                    {currentTier.topUpIntervalDays} days.
                  </div>
                )}
              </div>
            </div>
            {previousTier && (
              <Button
                variant="outline"
                onClick={() => handleDowngradeTier(previousTier.tier)}
                disabled={downgradeTier.isPending}
              >
                {downgradeTier.isPending && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                Downgrade
              </Button>
            )}
          </div>
        )}
      </CardContent>
    </Card>
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

function checkTopUpRequirementStatus(currentTier: OrganizationTier, nextTier: Tier) {
  if (!currentTier) {
    return false
  }

  if (currentTier.largestSuccessfulPaymentCents < nextTier.minTopUpAmountCents) {
    return false
  }

  if (nextTier.topUpIntervalDays && currentTier.largestSuccessfulPaymentDate) {
    const diffTime = Math.abs(Date.now() - (currentTier.largestSuccessfulPaymentDate?.getTime() || 0))
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))

    return diffDays < nextTier.topUpIntervalDays
  }

  return true
}

function getTierRequirementItems(
  requirementsState: {
    emailVerified: boolean
    creditCardLinked: boolean
    githubConnected: boolean
    businessEmailVerified: boolean
  },
  currentTier?: OrganizationTier | null,
  tier?: Tier | null,
) {
  if (!tier || !currentTier) {
    return []
  }
  if (tier.tier < 1 || tier.tier > 4) {
    return []
  }

  const items = []

  if (tier.tier === 1) {
    items.push({
      label: 'Email verification',
      isChecked: requirementsState.emailVerified,
      link: RoutePath.ACCOUNT_SETTINGS,
    })
  }
  if (tier.tier === 2) {
    items.push(
      {
        label: 'Credit card linked',
        isChecked: requirementsState.creditCardLinked,
        link: RoutePath.BILLING_WALLET,
      },
      {
        label: 'GitHub connected',
        isChecked: requirementsState.githubConnected,
        link: RoutePath.ACCOUNT_SETTINGS,
      },
    )
  }
  if (tier.tier === 3) {
    items.push({
      label: 'Business email verified',
      isChecked: requirementsState.businessEmailVerified,
      link: RoutePath.BILLING_WALLET,
    })
  }

  if (tier.minTopUpAmountCents) {
    items.push({
      label: `Top up ${getDollarAmount(tier.minTopUpAmountCents)} (${tier.topUpIntervalDays ? `every ${tier.topUpIntervalDays} days` : 'one time'})`,
      isChecked: checkTopUpRequirementStatus(currentTier, tier),
      link: RoutePath.BILLING_WALLET,
    })
  }

  return items
}

interface TierRequirementItemProps {
  checked: boolean
  label: string
  link?: string
  externalLink?: boolean
}

function RequirementIcon({ checked, label }: { checked: boolean; label: string }) {
  return (
    <div
      className={cn(
        'flex-shrink-0 w-3.5 h-3.5 rounded-full flex items-center justify-center border border-muted-foreground/50',
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

function TierRequirementItem({ checked, label, link, externalLink }: TierRequirementItemProps) {
  const content = (
    <span className="flex items-center gap-2 text-sm">
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
