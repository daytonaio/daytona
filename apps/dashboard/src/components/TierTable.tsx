/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RoutePath } from '@/enums/RoutePath'
import { Button } from './ui/button'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from './ui/table'
import { Check, CreditCard, Package, Github, PhoneCall, FileText } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

export function TierTable() {
  const navigate = useNavigate()

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Tier</TableHead>
          <TableHead>Concurrent Resources (vCPU / RAM / Disk)</TableHead>
          <TableHead>Per-sandbox Resources (vCPU / RAM / Disk)</TableHead>
          <TableHead>Number of Volumes</TableHead>
          {/* <TableHead>Persistence</TableHead> */}
          <TableHead>Access Verification</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell>
            <b>Tier 1</b>
          </TableCell>
          <TableCell>10 vCPU / 10 GiB / 30 GiB</TableCell>
          <TableCell>Up to 2 vCPU / 2 GiB / 3 GiB</TableCell>
          <TableCell>3</TableCell>
          {/* <TableCell>
            <X size={18} className="inline align-text-bottom text-destructive" aria-label="No" /> No
          </TableCell> */}
          <TableCell>
            <Check size={18} className="inline align-text-bottom text-success" aria-label="Email verification" /> Email
            verification
          </TableCell>
        </TableRow>
        <TableRow>
          <TableCell>
            <b>Tier 2</b>
          </TableCell>
          <TableCell>100 vCPU / 200 GiB / 300 GiB</TableCell>
          <TableCell>Up to 4 vCPU / 8 GiB / 5 GiB</TableCell>
          <TableCell>5</TableCell>
          {/* <TableCell>
            <Check size={18} className="inline align-text-bottom text-success" aria-label="Yes" /> Yes
          </TableCell> */}
          <TableCell>
            <div className="flex items-center gap-12">
              <div>
                <CreditCard size={18} className="inline align-text-bottom" aria-label="Credit card linked" /> Credit
                card linked
                <br />
                <Package size={18} className="inline align-text-bottom" aria-label="Top-Up Wallet" /> Top-Up Wallet
                <br />
                <Github size={18} className="inline align-text-bottom" aria-label="GitHub connected" /> GitHub connected
              </div>
              <div>
                <Button
                  variant="outline"
                  onClick={() => {
                    navigate(RoutePath.BILLING)
                  }}
                >
                  Go to Billing
                </Button>
              </div>
            </div>
          </TableCell>
        </TableRow>
        {/* <TableRow>
          <TableCell>
            <b>Tier 3</b>
          </TableCell>
          <TableCell>500 vCPU / 1000 GiB / 1500 GiB</TableCell>
          <TableCell>Up to 8 vCPU / 18 GiB / 10 GiB</TableCell>
          <TableCell>10</TableCell>
          <TableCell>
            <Check size={18} className="inline align-text-bottom text-success" aria-label="Yes" /> Yes
          </TableCell>
          <TableCell>
            <Check size={18} className="inline align-text-bottom text-success" aria-label="Spent $1,000" />{' '}
            <b>Spent $1,000</b> (total)
            <br />
            <Phone size={18} className="inline align-text-bottom" aria-label="Phone number linked (2FA)" />{' '}
            <b>Phone number linked (2FA)</b>
            <br />
            <Shield size={18} className="inline align-text-bottom" aria-label="Auto abuse checks (IP, proxies)" /> Auto
            abuse checks (IP, proxies)
          </TableCell>
        </TableRow> */}
        <TableRow>
          <TableCell>
            <b>Custom</b>
          </TableCell>
          <TableCell>Custom</TableCell>
          <TableCell>Custom</TableCell>
          <TableCell>Custom</TableCell>
          {/* <TableCell>
            <Check size={18} className="inline align-text-bottom text-success" aria-label="Yes" /> Yes
          </TableCell> */}
          <TableCell>
            <PhoneCall size={18} className="inline align-text-bottom" aria-label="Contact sales" />{' '}
            <b>Contact sales at sales@daytona.io</b>
            <br />
            <FileText
              size={18}
              className="inline align-text-bottom"
              aria-label="Custom agreement & verification"
            />{' '}
            <b>Custom agreement & verification</b>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  )
}
