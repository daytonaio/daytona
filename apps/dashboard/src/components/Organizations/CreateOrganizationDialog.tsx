/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useState } from 'react'
import { Check, Copy } from 'lucide-react'
import { Organization } from '@daytonaio/api-client'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Link } from 'react-router-dom'
import { Label } from '@/components/ui/label'
import { RoutePath } from '@/enums/RoutePath'

interface CreateOrganizationDialogProps {
  open: boolean
  billingApiUrl?: string
  onOpenChange: (open: boolean) => void
  onCreateOrganization: (name: string) => Promise<Organization | null>
}

export const CreateOrganizationDialog: React.FC<CreateOrganizationDialogProps> = ({
  open,
  billingApiUrl,
  onOpenChange,
  onCreateOrganization,
}) => {
  const [name, setName] = useState('')
  const [loading, setLoading] = useState(false)
  const [createdOrg, setCreatedOrg] = useState<Organization | null>(null)
  const [copied, setCopied] = useState<string | null>(null)

  const handleCreateOrganization = async () => {
    setLoading(true)
    const org = await onCreateOrganization(name)
    if (org) {
      setCreatedOrg(org)
      setName('')
    }
    setLoading(false)
  }

  const copyToClipboard = async (text: string, label: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(label)
      setTimeout(() => setCopied(null), 2000)
    } catch (err) {
      console.error('Failed to copy text:', err)
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        onOpenChange(isOpen)
        if (!isOpen) {
          setName('')
          setCreatedOrg(null)
          setCopied(null)
        }
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{createdOrg ? 'New Organization' : 'Create New Organization'}</DialogTitle>
          <DialogDescription>
            {createdOrg
              ? 'You can switch between organizations in the top left corner of the sidebar.'
              : 'Create a new organization to share resources and collaborate with others.'}
          </DialogDescription>
        </DialogHeader>
        {createdOrg ? (
          <div className="space-y-6">
            <div className="space-y-3">
              <Label htmlFor="organization-id">Organization ID</Label>
              <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                <span className="overflow-x-auto pr-2 cursor-text select-all">{createdOrg.id}</span>
                {(copied === 'Organization ID' && <Check className="w-4 h-4" />) || (
                  <Copy
                    className="w-4 h-4 cursor-pointer"
                    onClick={() => copyToClipboard(createdOrg.id, 'Organization ID')}
                  />
                )}
              </div>
            </div>

            <div className="p-3 rounded-md bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400">
              <p className="font-medium">Your organization is created.</p>
              <p className="text-sm mt-1">
                {billingApiUrl ? (
                  <>
                    To get started, add a payment method on the{' '}
                    <Link
                      to={RoutePath.BILLING_WALLET}
                      className="text-blue-500 hover:underline"
                      onClick={(e) => {
                        onOpenChange(false)
                      }}
                    >
                      wallet page
                    </Link>
                    .
                  </>
                ) : (
                  <></>
                )}
              </p>
            </div>
          </div>
        ) : (
          <form
            id="create-organization-form"
            className="space-y-6 overflow-y-auto px-1 pb-1"
            onSubmit={async (e) => {
              e.preventDefault()
              await handleCreateOrganization()
            }}
          >
            <div className="space-y-3">
              <Label htmlFor="organization-name">Organization Name</Label>
              <Input id="organization-name" value={name} onChange={(e) => setName(e.target.value)} placeholder="Name" />
            </div>
          </form>
        )}
        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="secondary" disabled={loading}>
              {createdOrg ? 'Close' : 'Cancel'}
            </Button>
          </DialogClose>
          {!createdOrg &&
            (loading ? (
              <Button type="button" variant="default" disabled>
                Creating...
              </Button>
            ) : (
              <Button type="submit" form="create-organization-form" variant="default" disabled={!name.trim()}>
                Create
              </Button>
            ))}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
