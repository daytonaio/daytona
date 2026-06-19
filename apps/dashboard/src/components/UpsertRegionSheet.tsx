/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateResourceButton } from '@/components/CreateResourceButton'
import { Button } from '@/components/ui/button'
import { Field, FieldDescription, FieldError, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { InputGroup, InputGroupButton, InputGroupInput } from '@/components/ui/input-group'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { Spinner } from '@/components/ui/spinner'
import { useCopyToClipboard } from '@/hooks/useCopyToClipboard'
import { useCreateRegionMutation } from '@/hooks/mutations/useCreateRegionMutation'
import { useUpdateRegionMutation } from '@/hooks/mutations/useUpdateRegionMutation'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { getMaskedToken } from '@/lib/utils'
import { CreateRegionResponse, Region, UpdateRegion } from '@daytona/api-client'
import { useForm } from '@tanstack/react-form'
import { CheckIcon, CopyIcon, EyeIcon, EyeOffIcon } from 'lucide-react'
import { Ref, type ReactNode, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState } from 'react'
import { toast } from 'sonner'
import { z } from 'zod'

const REGION_NAME_REGEX = /^[a-zA-Z0-9._-]+$/

const optionalUrlSchema = z
  .string()
  .trim()
  .refine((value) => value.length === 0 || z.string().url().safeParse(value).success, 'Must be a valid URL')

const regionUrlsSchema = z.object({
  proxyUrl: optionalUrlSchema,
  sshGatewayUrl: optionalUrlSchema,
  snapshotManagerUrl: optionalUrlSchema,
})

const createFormSchema = regionUrlsSchema.extend({
  name: z
    .string()
    .min(1, 'Region name is required')
    .refine(
      (value) => REGION_NAME_REGEX.test(value),
      'Only letters, numbers, underscores, periods, and hyphens are allowed',
    ),
})

const editFormSchema = regionUrlsSchema.extend({
  name: z.string(),
})

type FormValues = z.infer<typeof createFormSchema>

const defaultValues: FormValues = {
  name: '',
  proxyUrl: '',
  sshGatewayUrl: '',
  snapshotManagerUrl: '',
}

type UpsertRegionSheetMode = 'create' | 'edit'

interface UpsertRegionSheetProps {
  className?: string
  disabled?: boolean
  trigger?: ReactNode | null
  ref?: Ref<{ open: () => void }>
  mode?: UpsertRegionSheetMode
  open?: boolean
  onOpenChange?: (open: boolean) => void
  region?: Region | null
}

const hasRegionCredentials = (region: CreateRegionResponse | null) => {
  if (!region) return false

  return Boolean(
    region.proxyApiKey || region.sshGatewayApiKey || region.snapshotManagerUsername || region.snapshotManagerPassword,
  )
}

const getRegionUpdate = (value: FormValues, region: Region): UpdateRegion => {
  const updateData: UpdateRegion = {}

  const proxyUrlValue = value.proxyUrl.trim() || null
  const sshGatewayUrlValue = value.sshGatewayUrl.trim() || null
  const snapshotManagerUrlValue = value.snapshotManagerUrl.trim() || null

  if (proxyUrlValue !== (region.proxyUrl || null)) {
    updateData.proxyUrl = proxyUrlValue
  }
  if (sshGatewayUrlValue !== (region.sshGatewayUrl || null)) {
    updateData.sshGatewayUrl = sshGatewayUrlValue
  }
  if (snapshotManagerUrlValue !== (region.snapshotManagerUrl || null)) {
    updateData.snapshotManagerUrl = snapshotManagerUrlValue
  }

  return updateData
}

const hasRegionChanges = (value: FormValues, region: Region) => Object.keys(getRegionUpdate(value, region)).length > 0

export const UpsertRegionSheet = ({
  className,
  disabled,
  trigger,
  ref,
  mode = 'create',
  open,
  onOpenChange,
  region,
}: UpsertRegionSheetProps) => {
  const [internalOpen, setInternalOpen] = useState(false)
  const [createdRegion, setCreatedRegion] = useState<CreateRegionResponse | null>(null)
  const [isProxyApiKeyRevealed, setIsProxyApiKeyRevealed] = useState(false)
  const [isSshGatewayApiKeyRevealed, setIsSshGatewayApiKeyRevealed] = useState(false)
  const [isSnapshotManagerPasswordRevealed, setIsSnapshotManagerPasswordRevealed] = useState(false)

  const isEditMode = mode === 'edit'
  const isControlled = open !== undefined
  const isOpen = open ?? internalOpen

  const formRef = useRef<HTMLFormElement>(null)
  const { selectedOrganization } = useSelectedOrganization()
  const { reset: resetCreateRegionMutation, ...createRegionMutation } = useCreateRegionMutation()
  const { reset: resetUpdateRegionMutation, ...updateRegionMutation } = useUpdateRegionMutation()

  const handleOpenChange = useCallback(
    (nextOpen: boolean) => {
      if (!isControlled) {
        setInternalOpen(nextOpen)
      }
      onOpenChange?.(nextOpen)
    },
    [isControlled, onOpenChange],
  )

  useImperativeHandle(ref, () => ({
    open: () => handleOpenChange(true),
  }))

  const getDefaultValues = useCallback((): FormValues => {
    if (!isEditMode || !region) {
      return defaultValues
    }

    return {
      name: region.name,
      proxyUrl: region.proxyUrl || '',
      sshGatewayUrl: region.sshGatewayUrl || '',
      snapshotManagerUrl: region.snapshotManagerUrl || '',
    }
  }, [isEditMode, region])

  const form = useForm({
    defaultValues: getDefaultValues(),
    validators: {
      onSubmit: isEditMode ? editFormSchema : createFormSchema,
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
      if (!selectedOrganization?.id) {
        toast.error(`Select an organization to ${isEditMode ? 'update' : 'create'} a region.`)
        return
      }

      try {
        if (isEditMode) {
          if (!region) {
            toast.error('No region selected for editing.')
            return
          }

          const updateData = getRegionUpdate(value, region)
          if (Object.keys(updateData).length === 0) {
            return
          }

          await updateRegionMutation.mutateAsync({
            regionId: region.id,
            region: updateData,
            organizationId: selectedOrganization.id,
          })
          toast.success('Region updated successfully')
          handleOpenChange(false)
          return
        }

        const createRegionData = {
          name: value.name.trim(),
          proxyUrl: value.proxyUrl.trim() || null,
          sshGatewayUrl: value.sshGatewayUrl.trim() || null,
          snapshotManagerUrl: value.snapshotManagerUrl.trim() || null,
        }

        const created = await createRegionMutation.mutateAsync({
          region: createRegionData,
          organizationId: selectedOrganization.id,
        })
        toast.success(`Creating region ${createRegionData.name}`)

        if (hasRegionCredentials(created)) {
          setCreatedRegion(created)
          resetForm(defaultValues)
          return
        }

        setCreatedRegion(null)
        resetForm(defaultValues)
        handleOpenChange(false)
      } catch (error) {
        handleApiError(error, `Failed to ${isEditMode ? 'update' : 'create'} region`)
      }
    },
  })
  const { reset: resetForm } = form

  const resetState = useCallback(() => {
    setCreatedRegion(null)
    setIsProxyApiKeyRevealed(false)
    setIsSshGatewayApiKeyRevealed(false)
    setIsSnapshotManagerPasswordRevealed(false)
    resetForm(getDefaultValues())
    resetCreateRegionMutation()
    resetUpdateRegionMutation()
  }, [getDefaultValues, resetCreateRegionMutation, resetForm, resetUpdateRegionMutation])

  useEffect(() => {
    if (isOpen) {
      resetState()
    }
  }, [isOpen, resetState])

  const [copiedText, copyToClipboard] = useCopyToClipboard()
  const showCredentials = useMemo(() => !isEditMode && hasRegionCredentials(createdRegion), [createdRegion, isEditMode])
  const formId = isEditMode ? 'edit-region-form' : 'create-region-form'

  return (
    <Sheet open={isOpen} onOpenChange={handleOpenChange}>
      {trigger === undefined ? (
        <SheetTrigger asChild>
          {isEditMode ? (
            <Button variant="default" size="sm" disabled={disabled} className={className} title="Edit Region">
              Edit Region
            </Button>
          ) : (
            <CreateResourceButton resource="Region" disabled={disabled} className={className} title="Create Region" />
          )}
        </SheetTrigger>
      ) : (
        trigger
      )}

      <SheetContent className="w-dvw sm:w-[500px] p-0 flex flex-col gap-0">
        <SheetHeader className="border-b border-border p-4 px-5 items-center flex text-left flex-row">
          <SheetTitle>{showCredentials ? 'Region Created' : isEditMode ? 'Edit Region' : 'Create Region'}</SheetTitle>
          <SheetDescription className="sr-only">
            {showCredentials
              ? "Save these credentials securely. You won't be able to see them again."
              : isEditMode
                ? 'Update the region configuration.'
                : 'Add a new region for grouping runners and sandboxes.'}
          </SheetDescription>
        </SheetHeader>

        <ScrollArea fade="mask" fadeOffset={30} className="flex-1 min-h-0">
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
                id={formId}
                className="space-y-6"
                onSubmit={(e) => {
                  e.preventDefault()
                  e.stopPropagation()
                  form.handleSubmit()
                }}
              >
                {!isEditMode && (
                  <form.Field name="name">
                    {(field) => {
                      const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                      return (
                        <Field data-invalid={isInvalid}>
                          <FieldLabel htmlFor={field.name}>Region Name</FieldLabel>
                          <Input
                            aria-invalid={isInvalid}
                            autoComplete="off"
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
                )}

                <form.Field name="proxyUrl">
                  {(field) => {
                    const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                    return (
                      <Field data-invalid={isInvalid}>
                        <FieldLabel htmlFor={field.name}>Proxy URL</FieldLabel>
                        <Input
                          aria-invalid={isInvalid}
                          autoComplete="off"
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
                          autoComplete="off"
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
                          autoComplete="off"
                          id={field.name}
                          name={field.name}
                          value={field.state.value}
                          onBlur={field.handleBlur}
                          onChange={(e) => field.handleChange(e.target.value)}
                          placeholder="https://snapshot-manager.example.com"
                        />
                        <FieldDescription>
                          {isEditMode
                            ? '(Optional) URL of the custom snapshot manager for this region. Cannot be changed if snapshots exist in this region.'
                            : '(Optional) URL of the custom snapshot manager for this region.'}
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
          <Button type="button" variant="secondary" onClick={() => handleOpenChange(false)}>
            {showCredentials ? 'Close' : 'Cancel'}
          </Button>
          {!showCredentials && (
            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting, state.values] as const}
              children={([canSubmit, isSubmitting, values]) => {
                const hasChanges = !isEditMode || !region || hasRegionChanges(values, region)

                return (
                  <Button
                    type="submit"
                    form={formId}
                    variant="default"
                    disabled={
                      !canSubmit ||
                      isSubmitting ||
                      !hasChanges ||
                      (isEditMode ? updateRegionMutation.isPending : createRegionMutation.isPending)
                    }
                  >
                    {isSubmitting && <Spinner />}
                    {isEditMode ? 'Save' : 'Create'}
                  </Button>
                )
              }}
            />
          )}
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
