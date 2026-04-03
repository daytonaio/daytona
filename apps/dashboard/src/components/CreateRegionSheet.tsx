/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { Ref, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState } from 'react'
import { CreateRegion, CreateRegionResponse } from '@daytona/api-client'
import { useForm } from '@tanstack/react-form'
import { useMutation } from '@tanstack/react-query'
import { z } from 'zod'
import { CheckIcon, CopyIcon, EyeIcon, EyeOffIcon, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Field, FieldDescription, FieldError, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { InputGroup, InputGroupButton, InputGroupInput } from '@/components/ui/input-group'
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

const REGION_NAME_REGEX = /^[a-zA-Z0-9._-]+$/

const optionalUrlSchema = z
  .string()
  .trim()
  .refine((value) => value.length === 0 || z.string().url().safeParse(value).success, 'Must be a valid URL')

const formSchema = z.object({
  name: z
    .string()
    .min(1, 'Region name is required')
    .refine(
      (value) => REGION_NAME_REGEX.test(value),
      'Only letters, numbers, underscores, periods, and hyphens are allowed',
    ),
  proxyUrl: optionalUrlSchema,
  sshGatewayUrl: optionalUrlSchema,
  snapshotManagerUrl: optionalUrlSchema,
})

type FormValues = z.infer<typeof formSchema>

const defaultValues: FormValues = {
  name: '',
  proxyUrl: '',
  sshGatewayUrl: '',
  snapshotManagerUrl: '',
}

interface CreateRegionSheetProps {
  onCreateRegion: (data: CreateRegion) => Promise<CreateRegionResponse | null>
  writePermitted: boolean
  loadingData: boolean
  ref?: Ref<{ open: () => void }>
}

const hasRegionCredentials = (region: CreateRegionResponse | null) => {
  if (!region) return false

  return Boolean(
    region.proxyApiKey || region.sshGatewayApiKey || region.snapshotManagerUsername || region.snapshotManagerPassword,
  )
}

export const CreateRegionSheet: React.FC<CreateRegionSheetProps> = ({
  onCreateRegion,
  writePermitted,
  loadingData,
  ref,
}) => {
  const [open, setOpen] = useState(false)
  const [createdRegion, setCreatedRegion] = useState<CreateRegionResponse | null>(null)
  const [isProxyApiKeyRevealed, setIsProxyApiKeyRevealed] = useState(false)
  const [isSshGatewayApiKeyRevealed, setIsSshGatewayApiKeyRevealed] = useState(false)
  const [isSnapshotManagerPasswordRevealed, setIsSnapshotManagerPasswordRevealed] = useState(false)
  const formRef = useRef<HTMLFormElement>(null)

  useImperativeHandle(ref, () => ({
    open: () => setOpen(true),
  }))

  const createRegionMutation = useMutation({
    mutationFn: async (value: FormValues) => {
      const createRegionData: CreateRegion = {
        name: value.name.trim(),
        proxyUrl: value.proxyUrl.trim() || null,
        sshGatewayUrl: value.sshGatewayUrl.trim() || null,
        snapshotManagerUrl: value.snapshotManagerUrl.trim() || null,
      }

      return onCreateRegion(createRegionData)
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
      const region = await createRegionMutation.mutateAsync(value)

      if (!region) {
        return
      }

      if (!hasRegionCredentials(region)) {
        setOpen(false)
        setCreatedRegion(null)
      } else {
        setCreatedRegion(region)
      }

      resetForm(defaultValues)
    },
  })
  const { reset: resetForm } = form

  const { reset: resetMutation } = createRegionMutation

  const resetState = useCallback(() => {
    setCreatedRegion(null)
    setIsProxyApiKeyRevealed(false)
    setIsSshGatewayApiKeyRevealed(false)
    setIsSnapshotManagerPasswordRevealed(false)
    resetForm(defaultValues)
    resetMutation()
  }, [resetForm, resetMutation])

  useEffect(() => {
    if (open) {
      resetState()
    }
  }, [open, resetState])

  const [copiedText, copyToClipboard] = useCopyToClipboard()

  const showCredentials = useMemo(() => hasRegionCredentials(createdRegion), [createdRegion])

  if (!writePermitted) {
    return null
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="default" size="sm" disabled={loadingData} className="w-auto px-4">
          <Plus className="w-4 h-4" />
          Create Region
        </Button>
      </SheetTrigger>

      <SheetContent className="w-dvw sm:w-[500px] p-0 flex flex-col gap-0">
        <SheetHeader className="border-b border-border p-4 px-5 items-center flex text-left flex-row">
          <SheetTitle className="text-2xl">{createdRegion ? 'New Region Created' : 'Create New Region'}</SheetTitle>
          <SheetDescription className="sr-only">
            {!createdRegion
              ? 'Add a new region for grouping runners and sandboxes.'
              : "Save these credentials securely. You won't be able to see them again."}
          </SheetDescription>
        </SheetHeader>

        <ScrollArea fade="mask" className="flex-1 min-h-0">
          <div className="p-5">
            {showCredentials ? (
              <div className="space-y-6">
                {createdRegion?.proxyApiKey && (
                  <Field>
                    <FieldLabel htmlFor="proxy-api-key">Proxy API Key</FieldLabel>
                    <InputGroup className="pr-1 flex-1">
                      <InputGroupInput
                        id="proxy-api-key"
                        value={
                          isProxyApiKeyRevealed ? createdRegion.proxyApiKey : getMaskedToken(createdRegion.proxyApiKey)
                        }
                        readOnly
                      />
                      <InputGroupButton
                        variant="ghost"
                        size="icon-xs"
                        aria-label={isProxyApiKeyRevealed ? 'Hide proxy API key' : 'Show proxy API key'}
                        aria-pressed={isProxyApiKeyRevealed}
                        onClick={() => setIsProxyApiKeyRevealed((revealed) => !revealed)}
                      >
                        {isProxyApiKeyRevealed ? <EyeOffIcon className="h-4 w-4" /> : <EyeIcon className="h-4 w-4" />}
                      </InputGroupButton>
                      <InputGroupButton
                        variant="ghost"
                        size="icon-xs"
                        aria-label="Copy proxy API key"
                        onClick={() => copyToClipboard(createdRegion.proxyApiKey ?? '')}
                      >
                        {copiedText === createdRegion.proxyApiKey ? (
                          <CheckIcon className="h-4 w-4" />
                        ) : (
                          <CopyIcon className="h-4 w-4" />
                        )}
                      </InputGroupButton>
                    </InputGroup>
                  </Field>
                )}

                {createdRegion?.sshGatewayApiKey && (
                  <Field>
                    <FieldLabel htmlFor="ssh-gateway-api-key">SSH Gateway API Key</FieldLabel>
                    <InputGroup className="pr-1 flex-1">
                      <InputGroupInput
                        id="ssh-gateway-api-key"
                        value={
                          isSshGatewayApiKeyRevealed
                            ? createdRegion.sshGatewayApiKey
                            : getMaskedToken(createdRegion.sshGatewayApiKey)
                        }
                        readOnly
                      />
                      <InputGroupButton
                        variant="ghost"
                        size="icon-xs"
                        aria-label={
                          isSshGatewayApiKeyRevealed ? 'Hide SSH gateway API key' : 'Show SSH gateway API key'
                        }
                        aria-pressed={isSshGatewayApiKeyRevealed}
                        onClick={() => setIsSshGatewayApiKeyRevealed((revealed) => !revealed)}
                      >
                        {isSshGatewayApiKeyRevealed ? (
                          <EyeOffIcon className="h-4 w-4" />
                        ) : (
                          <EyeIcon className="h-4 w-4" />
                        )}
                      </InputGroupButton>
                      <InputGroupButton
                        variant="ghost"
                        size="icon-xs"
                        aria-label="Copy SSH gateway API key"
                        onClick={() => copyToClipboard(createdRegion.sshGatewayApiKey ?? '')}
                      >
                        {copiedText === createdRegion.sshGatewayApiKey ? (
                          <CheckIcon className="h-4 w-4" />
                        ) : (
                          <CopyIcon className="h-4 w-4" />
                        )}
                      </InputGroupButton>
                    </InputGroup>
                  </Field>
                )}

                {createdRegion?.snapshotManagerUsername && (
                  <Field>
                    <FieldLabel htmlFor="snapshot-manager-username">Snapshot manager username</FieldLabel>
                    <InputGroup className="pr-1 flex-1">
                      <InputGroupInput
                        id="snapshot-manager-username"
                        value={createdRegion.snapshotManagerUsername}
                        readOnly
                      />
                      <InputGroupButton
                        variant="ghost"
                        size="icon-xs"
                        aria-label="Copy snapshot manager username"
                        onClick={() => copyToClipboard(createdRegion.snapshotManagerUsername ?? '')}
                      >
                        {copiedText === createdRegion.snapshotManagerUsername ? (
                          <CheckIcon className="h-4 w-4" />
                        ) : (
                          <CopyIcon className="h-4 w-4" />
                        )}
                      </InputGroupButton>
                    </InputGroup>
                  </Field>
                )}

                {createdRegion?.snapshotManagerPassword && (
                  <Field>
                    <FieldLabel htmlFor="snapshot-manager-password">Snapshot manager password</FieldLabel>
                    <InputGroup className="pr-1 flex-1">
                      <InputGroupInput
                        id="snapshot-manager-password"
                        value={
                          isSnapshotManagerPasswordRevealed
                            ? createdRegion.snapshotManagerPassword
                            : getMaskedToken(createdRegion.snapshotManagerPassword)
                        }
                        readOnly
                      />
                      <InputGroupButton
                        variant="ghost"
                        size="icon-xs"
                        aria-label={
                          isSnapshotManagerPasswordRevealed
                            ? 'Hide snapshot manager password'
                            : 'Show snapshot manager password'
                        }
                        aria-pressed={isSnapshotManagerPasswordRevealed}
                        onClick={() => setIsSnapshotManagerPasswordRevealed((revealed) => !revealed)}
                      >
                        {isSnapshotManagerPasswordRevealed ? (
                          <EyeOffIcon className="h-4 w-4" />
                        ) : (
                          <EyeIcon className="h-4 w-4" />
                        )}
                      </InputGroupButton>
                      <InputGroupButton
                        variant="ghost"
                        size="icon-xs"
                        aria-label="Copy snapshot manager password"
                        onClick={() => copyToClipboard(createdRegion.snapshotManagerPassword ?? '')}
                      >
                        {copiedText === createdRegion.snapshotManagerPassword ? (
                          <CheckIcon className="h-4 w-4" />
                        ) : (
                          <CopyIcon className="h-4 w-4" />
                        )}
                      </InputGroupButton>
                    </InputGroup>
                  </Field>
                )}
              </div>
            ) : (
              <form
                ref={formRef}
                id="create-region-form"
                className="space-y-6"
                onSubmit={(e) => {
                  e.preventDefault()
                  e.stopPropagation()
                  form.handleSubmit()
                }}
              >
                <form.Field name="name">
                  {(field) => {
                    const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                    return (
                      <Field data-invalid={isInvalid}>
                        <FieldLabel htmlFor={field.name}>Region Name</FieldLabel>
                        <Input
                          aria-invalid={isInvalid}
                          id={field.name}
                          name={field.name}
                          value={field.state.value}
                          onBlur={field.handleBlur}
                          onChange={(e) => field.handleChange(e.target.value)}
                          placeholder="us-east-1"
                        />
                        <FieldDescription>
                          Region name must contain only letters, numbers, underscores, periods, and hyphens.
                        </FieldDescription>
                        {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                          <FieldError errors={field.state.meta.errors} />
                        )}
                      </Field>
                    )
                  }}
                </form.Field>

                <form.Field name="proxyUrl">
                  {(field) => {
                    const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                    return (
                      <Field data-invalid={isInvalid}>
                        <FieldLabel htmlFor={field.name}>Proxy URL</FieldLabel>
                        <Input
                          aria-invalid={isInvalid}
                          id={field.name}
                          name={field.name}
                          value={field.state.value}
                          onBlur={field.handleBlur}
                          onChange={(e) => field.handleChange(e.target.value)}
                          placeholder="https://proxy.example.com"
                        />
                        <FieldDescription>(Optional) URL of the custom proxy for this region.</FieldDescription>
                        {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                          <FieldError errors={field.state.meta.errors} />
                        )}
                      </Field>
                    )
                  }}
                </form.Field>

                <form.Field name="sshGatewayUrl">
                  {(field) => {
                    const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                    return (
                      <Field data-invalid={isInvalid}>
                        <FieldLabel htmlFor={field.name}>SSH gateway URL</FieldLabel>
                        <Input
                          aria-invalid={isInvalid}
                          id={field.name}
                          name={field.name}
                          value={field.state.value}
                          onBlur={field.handleBlur}
                          onChange={(e) => field.handleChange(e.target.value)}
                          placeholder="https://ssh-gateway.example.com"
                        />
                        <FieldDescription>(Optional) URL of the custom SSH gateway for this region.</FieldDescription>
                        {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                          <FieldError errors={field.state.meta.errors} />
                        )}
                      </Field>
                    )
                  }}
                </form.Field>

                <form.Field name="snapshotManagerUrl">
                  {(field) => {
                    const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                    return (
                      <Field data-invalid={isInvalid}>
                        <FieldLabel htmlFor={field.name}>Snapshot manager URL</FieldLabel>
                        <Input
                          aria-invalid={isInvalid}
                          id={field.name}
                          name={field.name}
                          value={field.state.value}
                          onBlur={field.handleBlur}
                          onChange={(e) => field.handleChange(e.target.value)}
                          placeholder="https://snapshot-manager.example.com"
                        />
                        <FieldDescription>
                          (Optional) URL of the custom snapshot manager for this region.
                        </FieldDescription>
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
            {showCredentials ? 'Close' : 'Cancel'}
          </Button>
          {!showCredentials && (
            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
                <Button type="submit" form="create-region-form" variant="default" disabled={!canSubmit || isSubmitting}>
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
