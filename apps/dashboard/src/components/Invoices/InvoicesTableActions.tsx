/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MoreHorizontalIcon } from 'lucide-react'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '../ui/alert-dialog'
import { Button } from '../ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../ui/dropdown-menu'
import { InvoicesTableActionsProps } from './types'

export function InvoicesTableActions({ invoice, onView, onVoid, onPay }: InvoicesTableActionsProps) {
  const handleDownload = () => {
    if (invoice.fileUrl) {
      window.open(invoice.fileUrl, '_blank')
    }
  }

  return (
    <div className="flex items-center gap-1">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
            <MoreHorizontalIcon className="h-4 w-4" aria-label="Open menu" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          {onView && (
            <DropdownMenuItem className="cursor-pointer" onSelect={() => onView?.(invoice)}>
              View
            </DropdownMenuItem>
          )}
          {onPay && (
            <DropdownMenuItem className="cursor-pointer" onSelect={() => onPay?.(invoice)}>
              Pay
            </DropdownMenuItem>
          )}
          {invoice.fileUrl && (
            <DropdownMenuItem className="cursor-pointer" onSelect={handleDownload}>
              Download
            </DropdownMenuItem>
          )}
          {Boolean(
            invoice.status === 'finalized' && ['pending', 'failed'].includes(invoice.paymentStatus) && onVoid,
          ) && (
            <>
              <DropdownMenuSeparator />
              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <DropdownMenuItem
                    className="cursor-pointer"
                    onSelect={(e) => e.preventDefault()}
                    variant="destructive"
                  >
                    Void
                  </DropdownMenuItem>
                </AlertDialogTrigger>
                <AlertDialogContent className="sm:max-w-md">
                  <AlertDialogHeader>
                    <AlertDialogTitle>Void Invoice</AlertDialogTitle>
                    <AlertDialogDescription>
                      Are you sure you want to void the invoice <span className="font-bold">{invoice.number}</span>?
                      <br />
                      This action cannot be undone.
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction onClick={() => onVoid?.(invoice)} variant="destructive">
                      Void
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </>
          )}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
