/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useCallback, useEffect, useState } from 'react'
import { SshAccessDto } from '@daytonaio/api-client'
import { useForm } from '@tanstack/react-form'
import { CheckIcon, CopyIcon, InfoIcon } from 'lucide-react'
import { AnimatePresence, motion } from 'motion/react'
import { NumericFormat } from 'react-number-format'
import { z } from 'zod'
import { Field, FieldError, FieldLabel } from '@/components/ui/field'
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
  InputGroupText,
} from '@/components/ui/input-group'
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

const formSchema = z.object({
  expiryMinutes: z.number().int('Must be a whole number').min(1, 'Minimum 1 minute').max(1440, 'Maximum 1440 minutes'),
})

type FormValues = z.infer<typeof formSchema>

const defaultValues: FormValues = {
  expiryMinutes: 60,
}

export function CreateSshAccessDialog({ sandboxId, open, onOpenChange }: CreateSshAccessDialogProps) {
  const [sshAccess, setSshAccess] = useState<SshAccessDto | null>(null)
  const { reset: resetMutation, ...createMutation } = useCreateSshAccessMutation()

  const form = useForm({
    defaultValues,
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: async ({ value }) => {
      try {
        const result = await createMutation.mutateAsync({
          sandboxId,
          expiresInMinutes: value.expiryMinutes,
        })
        setSshAccess(result)
      } catch (error) {
        handleApiError(error, 'Failed to create SSH access')
      }
    },
  })

  const resetState = useCallback(() => {
    form.reset(defaultValues)
    resetMutation()
    setSshAccess(null)
  }, [form, resetMutation])

  useEffect(() => {
    if (open) {
      resetState()
    }
  }, [open, resetState])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
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
          <form
            id="create-ssh-form"
            onSubmit={(e) => {
              e.preventDefault()
              e.stopPropagation()
              form.handleSubmit()
            }}
          >
            <form.Field name="expiryMinutes">
              {(field) => {
                const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                return (
                  <Field data-invalid={isInvalid}>
                    <FieldLabel htmlFor={field.name}>Expiry</FieldLabel>
                    <InputGroup>
                      <NumericFormat
                        customInput={InputGroupInput}
                        aria-invalid={isInvalid}
                        id={field.name}
                        name={field.name}
                        inputMode="numeric"
                        allowNegative={false}
                        decimalScale={0}
                        value={field.state.value}
                        onBlur={field.handleBlur}
                        onValueChange={({ floatValue }) => field.handleChange(floatValue ?? 0)}
                      />
                      <InputGroupAddon align="inline-end">
                        <InputGroupText>min</InputGroupText>
                      </InputGroupAddon>
                    </InputGroup>
                    {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                      <FieldError errors={field.state.meta.errors} />
                    )}
                  </Field>
                )
              }}
            </form.Field>
          </form>
        )}
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="secondary">Close</Button>
          </DialogClose>
          {!sshAccess && (
            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
                <Button type="submit" form="create-ssh-form" disabled={!canSubmit || isSubmitting}>
                  {isSubmitting && <Spinner />}
                  Create
                </Button>
              )}
            />
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
