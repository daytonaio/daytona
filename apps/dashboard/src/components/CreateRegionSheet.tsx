/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { Ref, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState } from 'react'
import { CreateRegion, CreateRegionResponse } from '@daytonaio/api-client'
import { useForm } from '@tanstack/react-form'
import { useMutation } from '@tanstack/react-query'
import { z } from 'zod'
import { Copy, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Field, FieldDescription, FieldError, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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

  const [, copyToClipboard] = useCopyToClipboard()

  const showCredentials = useMemo(() => hasRegionCredentials(createdRegion), [createdRegion])

  if (!writePermitted) {
    return null
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="default" size="sm" disabled={loadingData} className="w-auto px-4" title="Create Region">
          <Plus className="w-4 h-4" />
          Create Region
        </Button>
      </SheetTrigger>

      <SheetContent className="w-dvw sm:w-[560px] p-0 flex flex-col gap-0">
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
                  <div className="space-y-3">
                    <Label htmlFor="proxy-api-key">Proxy API Key</Label>
                    <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                      <span
                        className="overflow-x-auto pr-2 cursor-text select-all"
                        onMouseEnter={() => setIsProxyApiKeyRevealed(true)}
                        onMouseLeave={() => setIsProxyApiKeyRevealed(false)}
                      >
                        {isProxyApiKeyRevealed ? createdRegion.proxyApiKey : getMaskedToken(createdRegion.proxyApiKey)}
                      </span>
                      <Copy
                        className="w-4 h-4 cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                        onClick={() => copyToClipboard(createdRegion.proxyApiKey ?? '')}
                      />
                    </div>
                  </div>
                )}

                {createdRegion?.sshGatewayApiKey && (
                  <div className="space-y-3">
                    <Label htmlFor="ssh-gateway-api-key">SSH Gateway API Key</Label>
                    <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                      <span
                        className="overflow-x-auto pr-2 cursor-text select-all"
                        onMouseEnter={() => setIsSshGatewayApiKeyRevealed(true)}
                        onMouseLeave={() => setIsSshGatewayApiKeyRevealed(false)}
                      >
                        {isSshGatewayApiKeyRevealed
                          ? createdRegion.sshGatewayApiKey
                          : getMaskedToken(createdRegion.sshGatewayApiKey)}
                      </span>
                      <Copy
                        className="w-4 h-4 cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                        onClick={() => copyToClipboard(createdRegion.sshGatewayApiKey ?? '')}
                      />
                    </div>
                  </div>
                )}

                {createdRegion?.snapshotManagerUsername && (
                  <div className="space-y-3">
                    <Label htmlFor="snapshot-manager-username">Snapshot manager username</Label>
                    <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                      <span className="overflow-x-auto pr-2 cursor-text select-all">
                        {createdRegion.snapshotManagerUsername}
                      </span>
                      <Copy
                        className="w-4 h-4 cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                        onClick={() => copyToClipboard(createdRegion.snapshotManagerUsername ?? '')}
                      />
                    </div>
                  </div>
                )}

                {createdRegion?.snapshotManagerPassword && (
                  <div className="space-y-3">
                    <Label htmlFor="snapshot-manager-password">Snapshot manager password</Label>
                    <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                      <span
                        className="overflow-x-auto pr-2 cursor-text select-all"
                        onMouseEnter={() => setIsSnapshotManagerPasswordRevealed(true)}
                        onMouseLeave={() => setIsSnapshotManagerPasswordRevealed(false)}
                      >
                        {isSnapshotManagerPasswordRevealed
                          ? createdRegion.snapshotManagerPassword
                          : getMaskedToken(createdRegion.snapshotManagerPassword)}
                      </span>
                      <Copy
                        className="w-4 h-4 cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                        onClick={() => copyToClipboard(createdRegion.snapshotManagerPassword ?? '')}
                      />
                    </div>
                  </div>
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
                  {isSubmitting ? 'Creating...' : 'Create'}
                </Button>
              )}
            />
          )}
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
