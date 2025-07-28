/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import { OrganizationEmailsTableHeaderProps } from './types'
import { Input } from '../ui/input'
import { Button } from '../ui/button'
import { Plus, Search, Filter, Mail } from 'lucide-react'
import { DebouncedInput } from '../DebouncedInput'
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../ui/dropdown-menu'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '../ui/dialog'
import { Label } from '../ui/label'
import { ORGANIZATION_EMAILS_TABLE_CONSTANTS } from './constants'

export function OrganizationEmailsTableHeader({ table, onAddEmail }: OrganizationEmailsTableHeaderProps) {
  const [globalFilter, setGlobalFilter] = React.useState('')
  const [addEmailDialogOpen, setAddEmailDialogOpen] = React.useState(false)
  const [newEmail, setNewEmail] = React.useState('')
  const [isValidEmail, setIsValidEmail] = React.useState(true)

  React.useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        // Focus search input
        const searchInput = document.querySelector('[data-search-input]') as HTMLInputElement
        if (searchInput) {
          searchInput.focus()
        }
      }
    }

    document.addEventListener('keydown', down)
    return () => document.removeEventListener('keydown', down)
  }, [])

  const validateEmail = (email: string) => {
    return ORGANIZATION_EMAILS_TABLE_CONSTANTS.EMAIL_REGEX.test(email)
  }

  const handleEmailChange = (value: string) => {
    setNewEmail(value)
    setIsValidEmail(value === '' || validateEmail(value))
  }

  const handleAddEmail = () => {
    if (validateEmail(newEmail)) {
      onAddEmail(newEmail)
      setNewEmail('')
      setAddEmailDialogOpen(false)
      setIsValidEmail(true)
    }
  }

  const handleDialogOpenChange = (open: boolean) => {
    setAddEmailDialogOpen(open)
    if (!open) {
      setNewEmail('')
      setIsValidEmail(true)
    }
  }

  return (
    <div className="flex items-center justify-between pb-4">
      <div className="flex flex-1 items-center space-x-2">
        <div className="relative w-full max-w-sm">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <DebouncedInput
            placeholder="Search emails..."
            value={globalFilter ?? ''}
            onChange={(value) => {
              setGlobalFilter(String(value))
              table.setGlobalFilter(String(value))
            }}
            className="pl-8"
            data-search-input
          />
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" className="ml-auto h-8">
              <Filter className="mr-2 h-4 w-4" />
              View
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-[150px]">
            <DropdownMenuLabel>Toggle columns</DropdownMenuLabel>
            <DropdownMenuSeparator />
            {table
              .getAllColumns()
              .filter((column) => typeof column.accessorFn !== 'undefined' && column.getCanHide())
              .map((column) => {
                return (
                  <DropdownMenuCheckboxItem
                    key={column.id}
                    className="capitalize"
                    checked={column.getIsVisible()}
                    onCheckedChange={(value) => column.toggleVisibility(!!value)}
                  >
                    {column.id}
                  </DropdownMenuCheckboxItem>
                )
              })}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      <div className="ml-4">
        <Dialog open={addEmailDialogOpen} onOpenChange={handleDialogOpenChange}>
          <DialogTrigger asChild>
            <Button className="h-8">
              <Plus className="mr-2 h-4 w-4" />
              Add Email
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-[425px]">
            <DialogHeader>
              <DialogTitle>Add Organization Email</DialogTitle>
              <DialogDescription>
                Add a new email address to your organization. A verification email will be sent to this address.
              </DialogDescription>
            </DialogHeader>
            <div className="gap-4 py-4">
              <div className="space-y-3">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="Enter email address"
                  value={newEmail}
                  onChange={(e) => handleEmailChange(e.target.value)}
                  className={`${!isValidEmail ? 'border-red-500' : ''}`}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && validateEmail(newEmail)) {
                      handleAddEmail()
                    }
                  }}
                />
              </div>
              {!isValidEmail && newEmail && (
                <p className="text-sm text-red-500 mt-2">Please enter a valid email address</p>
              )}
            </div>
            <DialogFooter>
              <Button
                type="submit"
                onClick={handleAddEmail}
                disabled={!validateEmail(newEmail)}
                className="w-full sm:w-auto"
              >
                <Mail className="mr-2 h-4 w-4" />
                Add Email
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </div>
  )
}
