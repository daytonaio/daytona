/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useState } from 'react'
import { SshAccessDto } from '@daytonaio/api-client'
import { CheckIcon, CopyIcon, InfoIcon } from 'lucide-react'
import { AnimatePresence, motion } from 'motion/react'
import { Field, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { InputGroup, InputGroupButton, InputGroupInput } from '@/components/ui/input-group'
import { Spinner } from '@/components/ui/spinner'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useCreateSshAccessMutation } from '@/hooks/mutations/useCreateSshAccessMutation'
import { useCopyToClipboard } from '@/hooks/useCopyToClipboard'
import { handleApiError } from '@/lib/error-handling'

interface CreateSshAccessDialogProps {
  sandboxId: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

const MotionCopyIcon = motion(CopyIcon)
const MotionCheckIcon = motion(CheckIcon)

const iconProps = {
  initial: { opacity: 0, y: 5 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -5 },
  transition: { duration: 0.1 },
}

export function CreateSshAccessDialog({ sandboxId, open, onOpenChange }: CreateSshAccessDialogProps) {
  const [expiryMinutes, setExpiryMinutes] = useState(60)
  const [sshAccess, setSshAccess] = useState<SshAccessDto | null>(null)
  const createMutation = useCreateSshAccessMutation()

  const handleOpenChange = (isOpen: boolean) => {
    onOpenChange(isOpen)
    if (!isOpen) {
      setSshAccess(null)
      setExpiryMinutes(60)
      createMutation.reset()
    }
  }

  const handleCreate = async () => {
    try {
      const result = await createMutation.mutateAsync({ sandboxId, expiresInMinutes: expiryMinutes })
      setSshAccess(result)
    } catch (error) {
      handleApiError(error, 'Failed to create SSH access')
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{sshAccess ? 'SSH Access Created' : 'Create SSH Access'}</DialogTitle>
          <DialogDescription>
            {sshAccess ? 'Your SSH access has been created successfully.' : 'Set the expiration time for SSH access.'}
          </DialogDescription>
        </DialogHeader>
        {sshAccess ? (
          <SshAccessCreated sshAccess={sshAccess} />
        ) : (
          <Field>
            <FieldLabel htmlFor="ssh-expiry">Expiry (minutes)</FieldLabel>
            <Input
              id="ssh-expiry"
              type="number"
              min={1}
              max={1440}
              value={expiryMinutes}
              onChange={(e) => setExpiryMinutes(Number(e.target.value))}
            />
          </Field>
        )}
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="secondary">Close</Button>
          </DialogClose>
          {!sshAccess && (
            <Button onClick={handleCreate} disabled={createMutation.isPending}>
              {createMutation.isPending && <Spinner />}
              Create
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function SshAccessCreated({ sshAccess }: { sshAccess: SshAccessDto }) {
  const [copiedCommand, copyCommand] = useCopyToClipboard()

  return (
    <div className="space-y-4">
      <Alert variant="warning">
        <InfoIcon />
        <AlertDescription>Store the token safely — you won't be able to view it again.</AlertDescription>
      </Alert>
      <Field>
        <FieldLabel htmlFor="ssh-command">SSH Command</FieldLabel>
        <InputGroup className="pr-1">
          <InputGroupInput id="ssh-command" value={sshAccess.sshCommand} readOnly />
          <InputGroupButton variant="ghost" size="icon-xs" onClick={() => copyCommand(sshAccess.sshCommand)}>
            <AnimatePresence initial={false} mode="wait">
              {copiedCommand ? (
                <MotionCheckIcon className="size-4" key="copied" {...iconProps} />
              ) : (
                <MotionCopyIcon className="size-4" key="copy" {...iconProps} />
              )}
            </AnimatePresence>
          </InputGroupButton>
        </InputGroup>
      </Field>
    </div>
  )
}
