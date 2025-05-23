/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RoutePath } from '@/enums/RoutePath'
import { Button } from './ui/button'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from './ui/table'
import { PhoneCall, CheckCircle, Circle, Info } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { Tooltip } from './Tooltip'

type Props = {
  emailVerified: boolean
  githubConnected: boolean
  walletToppedUp: boolean
  creditCardConnected: boolean
}

export function TierTable({
  emailVerified,
  githubConnected,
  walletToppedUp,
  creditCardConnected: creditCardLinked,
}: Props) {
  const navigate = useNavigate()

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
                  Available Compute (vCPU / RAM / Disk)
                </div>
              }
              content={
                <div className="max-w-80">
                  Total vCPU, RAM, and Disk available at any moment across all running sandboxes.
                  <br />
                  The number of concurrent sandboxes depends on how much compute each one uses.
                </div>
              }
            />
          </TableHead>
          <TableHead>Access Verification</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell>
            <b>Tier 1</b>
          </TableCell>
          <TableCell>10 vCPU / 10 GiB / 30 GiB</TableCell>
          <TableCell className={cn(emailVerified ? 'text-green-500' : undefined)}>
            {getIcon(emailVerified, 'Email verification')} Email verification
          </TableCell>
        </TableRow>
        <TableRow>
          <TableCell>
            <b>Tier 2</b>
          </TableCell>
          <TableCell>100 vCPU / 200 GiB / 300 GiB</TableCell>
          <TableCell>
            <div className="grid grid-cols-2 gap-0 gap-y-4 py-2">
              <div className={cn(creditCardLinked ? 'text-green-500' : undefined)}>
                {getIcon(creditCardLinked, 'Credit card linked')} Credit card linked
              </div>
              <div className="row-span-2 border-b pb-2">
                <Button
                  variant="outline"
                  className="ml-4"
                  onClick={() => {
                    navigate(RoutePath.BILLING)
                  }}
                >
                  Go to Billing
                </Button>
              </div>
              <div className={cn(walletToppedUp ? 'text-green-500' : undefined, 'border-b', 'pb-2')}>
                {getIcon(walletToppedUp, 'Top-Up $10')} Top-Up $10
              </div>
              <div className={cn(githubConnected ? 'text-green-500' : undefined, 'content-center')}>
                {getIcon(githubConnected, 'GitHub connected')} GitHub connected
              </div>
              <div>
                <Button
                  variant="outline"
                  className="ml-4"
                  onClick={() => {
                    navigate(RoutePath.LINKED_ACCOUNTS)
                  }}
                >
                  Linked Accounts
                </Button>
              </div>
            </div>
          </TableCell>
        </TableRow>
        <TableRow>
          <TableCell>
            <b>Custom</b>
          </TableCell>
          <TableCell>Custom</TableCell>
          <TableCell>
            <div className="grid gap-0 gap-y-4 py-2">
              <div>
                <PhoneCall size={18} className="inline align-text-bottom" aria-label="Contact sales" />{' '}
                <b>Contact sales at sales@daytona.io</b>
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
    return <CheckCircle size={18} className="inline align-text-bottom" aria-label={label} />
  }
  return <Circle size={18} className="inline align-text-bottom" aria-label={label} />
}
