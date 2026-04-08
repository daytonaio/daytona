/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { Ref, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState } from 'react'
import { Region, CreateRunner, CreateRunnerResponse } from '@daytona/api-client'
import { useForm } from '@tanstack/react-form'
import { useMutation } from '@tanstack/react-query'
import { z } from 'zod'
import { AnimatePresence, motion } from 'framer-motion'
import { CheckIcon, CopyIcon, EyeIcon, EyeOffIcon, InfoIcon, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Field, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { InputGroup, InputGroupButton, InputGroupInput } from '@/components/ui/input-group'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Spinner } from '@/components/ui/spinner'
import { useCopyToClipboard } from '@/hooks/useCopyToClipboard'
import { getMaskedToken } from '@/lib/utils'

const formSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  regionId: z.string().min(1, 'Region is required'),
})

type FormValues = z.infer<typeof formSchema>

interface CreateRunnerSheetProps {
  regions: Region[]
  onCreateRunner: (data: CreateRunner) => Promise<CreateRunnerResponse | null>
  ref?: Ref<{ open: () => void }>
}

const buildDefaultValues = (regions: Region[]): FormValues => ({
  name: '',
  regionId: regions[0]?.id ?? '',
})

export const CreateRunnerSheet: React.FC<CreateRunnerSheetProps> = ({ regions, onCreateRunner, ref }) => {
  const [open, setOpen] = useState(false)
  const [createdRunner, setCreatedRunner] = useState<CreateRunnerResponse | null>(null)
  const formRef = useRef<HTMLFormElement>(null)

  useImperativeHandle(ref, () => ({
    open: () => setOpen(true),
  }))

  const defaultValues = useMemo(() => buildDefaultValues(regions), [regions])

  const createRunnerMutation = useMutation({
    mutationFn: async (value: FormValues) => {
      return onCreateRunner({
        name: value.name.trim(),
        regionId: value.regionId,
      })
    },
  })

  const form = useForm({
    defaultValues,
    validators: {
      onSubmit: formSchema,
    },
    onSubmitInvalid: () => {
      const formEl = formRef.current
      if (!formEl) return
      const invalidInput = formEl.querySelector('[aria-invalid="true"]') as HTMLElement | null
      if (invalidInput) {
        invalidInput.scrollIntoView({ behavior: 'smooth', block: 'center' })
        invalidInput.focus()
      }
    },
    onSubmit: async ({ value }) => {
      const runner = await createRunnerMutation.mutateAsync(value)
      if (!runner) {
        return
      }

      setCreatedRunner(runner)
      resetForm(buildDefaultValues(regions))
    },
  })
  const { reset: resetForm } = form

  useEffect(() => {
    if (!form.getFieldValue('regionId') && regions[0]?.id) {
      form.setFieldValue('regionId', regions[0].id)
    }
  }, [form, regions])

  const { reset: resetMutation } = createRunnerMutation

  const resetState = useCallback(() => {
    setCreatedRunner(null)
    resetForm(buildDefaultValues(regions))
    resetMutation()
  }, [resetForm, resetMutation, regions])

  useEffect(() => {
    if (open) {
      resetState()
    }
  }, [open, resetState])

  if (regions.length === 0) {
    return null
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="default" size="sm" className="w-auto px-4">
          <Plus className="w-4 h-4" />
          Create Runner
        </Button>
      </SheetTrigger>

      <SheetContent className="w-dvw sm:w-[500px] p-0 flex flex-col gap-0">
        <SheetHeader className="border-b border-border p-4 px-5 items-center flex text-left flex-row">
          <SheetTitle className="text-2xl">{createdRunner ? 'Runner Created' : 'Create New Runner'}</SheetTitle>
          <SheetDescription className="sr-only">
            {createdRunner
              ? 'Your runner token has been created successfully.'
              : 'Add configuration for a new runner in your selected region.'}
          </SheetDescription>
        </SheetHeader>

        <ScrollArea fade="mask" className="flex-1 min-h-0">
          <div className="p-5">
            {createdRunner ? (
              <CreatedRunnerDisplay createdRunner={createdRunner} />
            ) : (
              <form
                ref={formRef}
                id="create-runner-form"
                className="space-y-6"
                onSubmit={(e) => {
                  e.preventDefault()
                  e.stopPropagation()
                  form.handleSubmit()
                }}
              >
                <form.Field name="regionId">
                  {(field) => {
                    const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                    return (
                      <Field data-invalid={isInvalid}>
                        <FieldLabel htmlFor={field.name}>Region</FieldLabel>
                        <Select
                          value={field.state.value}
                          onValueChange={(value) => {
                            field.handleChange(value)
                          }}
                        >
                          <SelectTrigger className="h-8" id={field.name} aria-invalid={isInvalid}>
                            <SelectValue placeholder="Select a region" />
                          </SelectTrigger>
                          <SelectContent>
                            {regions.map((region) => (
                              <SelectItem key={region.id} value={region.id}>
                                {region.name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                          <FieldError errors={field.state.meta.errors} />
                        )}
                      </Field>
                    )
                  }}
                </form.Field>

                <form.Field name="name">
                  {(field) => {
                    const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                    return (
                      <Field data-invalid={isInvalid}>
                        <FieldLabel htmlFor={field.name}>Name</FieldLabel>
                        <Input
                          aria-invalid={isInvalid}
                          id={field.name}
                          name={field.name}
                          value={field.state.value}
                          onBlur={field.handleBlur}
                          onChange={(e) => field.handleChange(e.target.value)}
                          placeholder="runner-1"
                        />
                        {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                          <FieldError errors={field.state.meta.errors} />
                        )}
                      </Field>
                    )
                  }}
                </form.Field>
              </form>
            )}
          </div>
        </ScrollArea>

        <SheetFooter className="border-t border-border p-4 px-5">
          <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
            {createdRunner ? 'Close' : 'Cancel'}
          </Button>
          {!createdRunner && (
            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
                <Button type="submit" form="create-runner-form" variant="default" disabled={!canSubmit || isSubmitting}>
                  {isSubmitting && <Spinner />}
                  Create
                </Button>
              )}
            />
          )}
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}

const MotionCopyIcon = motion(CopyIcon)
const MotionCheckIcon = motion(CheckIcon)

const iconProps = {
  initial: { opacity: 0, y: 5 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -5 },
  transition: { duration: 0.1 },
}

function CreatedRunnerDisplay({ createdRunner }: { createdRunner: CreateRunnerResponse }) {
  const [copiedToken, copyToken] = useCopyToClipboard()
  const [tokenRevealed, setTokenRevealed] = useState(false)

  return (
    <div className="space-y-6">
      <Alert variant="warning">
        <InfoIcon />
        <AlertDescription>You can only view this token once. Store it safely.</AlertDescription>
      </Alert>
      <FieldGroup className="gap-4">
        <Field>
          <FieldLabel htmlFor="runner-token">Runner Token</FieldLabel>

          <InputGroup className="pr-1 flex-1">
            <InputGroupInput
              id="runner-token"
              value={tokenRevealed ? createdRunner.apiKey : getMaskedToken(createdRunner.apiKey)}
              readOnly
            />
            <InputGroupButton
              variant="ghost"
              size="icon-xs"
              aria-label={tokenRevealed ? 'Hide runner token' : 'Show runner token'}
              aria-pressed={tokenRevealed}
              onClick={() => setTokenRevealed(!tokenRevealed)}
            >
              {tokenRevealed ? <EyeOffIcon className="h-4 w-4" /> : <EyeIcon className="h-4 w-4" />}
            </InputGroupButton>
            <InputGroupButton
              variant="ghost"
              size="icon-xs"
              aria-label="Copy runner token"
              onClick={() => copyToken(createdRunner.apiKey)}
            >
              <AnimatePresence initial={false} mode="wait">
                {copiedToken ? (
                  <MotionCheckIcon className="h-4 w-4" key="copied" {...iconProps} />
                ) : (
                  <MotionCopyIcon className="h-4 w-4" key="copy" {...iconProps} />
                )}
              </AnimatePresence>
            </InputGroupButton>
          </InputGroup>
        </Field>
      </FieldGroup>
    </div>
  )
}
