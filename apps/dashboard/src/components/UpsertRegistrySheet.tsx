/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

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
  SheetTrigger,
  SheetTitle,
} from '@/components/ui/sheet'
import { Spinner } from '@/components/ui/spinner'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import {
  REGISTRY_PROVIDER_LABELS,
  REGISTRY_PROVIDER_SPECS,
  REGISTRY_PROVIDER_TAB_CONTENT,
  REGISTRY_PROVIDER_VALUES,
  type ProviderFieldSpec,
  type ProviderFormSpec,
  type RegistryProvider,
} from '@/constants/RegistryProviders'
import { useCreateRegistryMutation } from '@/hooks/mutations/useCreateRegistryMutation'
import { useUpdateRegistryMutation } from '@/hooks/mutations/useUpdateRegistryMutation'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { DockerRegistry } from '@daytona/api-client'
import { useForm } from '@tanstack/react-form'
import { EyeIcon, EyeOffIcon, Plus } from 'lucide-react'
import { Ref, type ReactNode, useCallback, useEffect, useImperativeHandle, useRef, useState } from 'react'
import { toast } from 'sonner'
import { z } from 'zod'

const baseFormSchema = z.object({
  name: z.string().min(1, 'Registry name is required'),
  url: z.string(),
  username: z.string().min(1, 'Username is required'),
  project: z.string(),
})

const createFormSchema = baseFormSchema.extend({
  password: z.string().min(1, 'Password is required'),
})

const editFormSchema = baseFormSchema.extend({
  password: z.string(),
})

// ECR resolves credentials server-side via STS:AssumeRole, so the form
// doesn't collect a password. URL is required (no docker.io fallback).
const ecrCreateFormSchema = baseFormSchema.extend({
  url: z.string().trim().min(1, 'Registry URL is required'),
  username: z.string().trim().min(1, 'Role ARN is required'),
  password: z.string(),
})

// Google Artifact Registry: URL is region-specific, must be provided.
const gcpCreateFormSchema = baseFormSchema.extend({
  url: z.string().trim().min(1, 'Registry URL is required'),
  password: z.string().trim().min(1, 'Service Account JSON Key is required'),
})

type FormValues = z.infer<typeof createFormSchema>

const CREATE_SCHEMAS: Partial<Record<RegistryProvider, typeof createFormSchema>> = {
  ecr: ecrCreateFormSchema as typeof createFormSchema,
  gcp: gcpCreateFormSchema as typeof createFormSchema,
}

const defaultValues: FormValues = {
  name: '',
  url: '',
  username: '',
  password: '',
  project: '',
}

const createDefaultsFor = (spec: ProviderFormSpec): FormValues => ({
  ...defaultValues,
  url: spec.url.defaultValue ?? '',
  username: spec.username.defaultValue ?? '',
  password: spec.password.defaultValue ?? '',
  project: spec.project.defaultValue ?? '',
})

// Hidden fields skip user input and submit their spec defaultValue
// (e.g. docker.io, _json_key, ghcr.io). Visible fields use the typed value.
const resolveField = (spec: ProviderFieldSpec, raw: string): string =>
  spec.hidden ? (spec.defaultValue ?? '') : raw.trim()

type UpsertRegistrySheetMode = 'create' | 'edit'

interface UpsertRegistrySheetProps {
  className?: string
  disabled?: boolean
  trigger?: ReactNode | null
  ref?: Ref<{ open: () => void }>
  mode?: UpsertRegistrySheetMode
  open?: boolean
  onOpenChange?: (open: boolean) => void
  registry?: DockerRegistry | null
}

export const UpsertRegistrySheet = ({
  className,
  disabled,
  trigger,
  ref,
  mode = 'create',
  open,
  onOpenChange,
  registry,
}: UpsertRegistrySheetProps) => {
  const [internalOpen, setInternalOpen] = useState(false)
  const [passwordVisible, setPasswordVisible] = useState(false)
  const [provider, setProvider] = useState<RegistryProvider>('generic')

  const isEditMode = mode === 'edit'
  const isControlled = open !== undefined
  const isOpen = open ?? internalOpen

  // Edit mode keeps today's static labels & validation by always reading the
  // generic spec. Create mode reads the active provider's spec.
  const activeSpec = isEditMode ? REGISTRY_PROVIDER_SPECS.generic : REGISTRY_PROVIDER_SPECS[provider]

  const { selectedOrganization } = useSelectedOrganization()
  const { reset: resetCreateRegistryMutation, ...createRegistryMutation } = useCreateRegistryMutation()
  const { reset: resetUpdateRegistryMutation, ...updateRegistryMutation } = useUpdateRegistryMutation()
  const formRef = useRef<HTMLFormElement>(null)

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
    if (!isEditMode || !registry) {
      return createDefaultsFor(REGISTRY_PROVIDER_SPECS.generic)
    }

    return {
      name: registry.name,
      url: registry.url,
      username: registry.username,
      password: '',
      project: registry.project,
    }
  }, [isEditMode, registry])

  const form = useForm({
    defaultValues: getDefaultValues(),
    validators: {
      onSubmit: isEditMode ? editFormSchema : (CREATE_SCHEMAS[provider] ?? createFormSchema),
    },
    onSubmitInvalid: () => {
      const formEl = formRef.current
      if (!formEl) return
      const invalidInput = formEl.querySelector('[aria-invalid="true"]') as HTMLInputElement | null
      if (invalidInput) {
        invalidInput.scrollIntoView({ behavior: 'smooth', block: 'center' })
        invalidInput.focus()
      }
    },
    onSubmit: async ({ value }) => {
      if (!selectedOrganization?.id) {
        toast.error(`Select an organization to ${isEditMode ? 'edit' : 'create'} a registry.`)
        return
      }

      // docker.io fallback applies only to Generic / Edit (today's behavior);
      // other providers require URL via Zod validation above.
      const resolvedUrl = resolveField(activeSpec.url, value.url)
      const url = !resolvedUrl && (isEditMode || provider === 'generic') ? 'docker.io' : resolvedUrl

      const payload = {
        name: value.name.trim(),
        url,
        username: resolveField(activeSpec.username, value.username),
        password: resolveField(activeSpec.password, value.password),
        project: resolveField(activeSpec.project, value.project),
      }

      try {
        if (isEditMode) {
          if (!registry) {
            toast.error('No registry selected for editing.')
            return
          }

          await updateRegistryMutation.mutateAsync({
            registryId: registry.id,
            registry: payload,
            organizationId: selectedOrganization.id,
          })
          toast.success('Registry edited successfully')
        } else {
          await createRegistryMutation.mutateAsync({
            registry: payload,
            organizationId: selectedOrganization.id,
          })
          toast.success('Registry created successfully')
        }

        handleOpenChange(false)
      } catch (error) {
        handleApiError(error, `Failed to ${isEditMode ? 'edit' : 'create'} registry`)
      }
    },
  })

  const { reset: resetForm } = form

  const resetState = useCallback(() => {
    setProvider('generic')
    resetForm(getDefaultValues())
    setPasswordVisible(false)
    resetCreateRegistryMutation()
    resetUpdateRegistryMutation()
  }, [resetForm, getDefaultValues, resetCreateRegistryMutation, resetUpdateRegistryMutation])

  useEffect(() => {
    if (isOpen) {
      resetState()
    }
  }, [isOpen, resetState])

  // Tab switch wipes per-tab fields so prior provider values can't leak into
  // the next provider's submission. Name is preserved as a quality-of-life win.
  const handleProviderChange = useCallback(
    (next: RegistryProvider) => {
      setProvider(next)
      const currentName = form.getFieldValue('name')
      form.reset({ ...createDefaultsFor(REGISTRY_PROVIDER_SPECS[next]), name: currentName })
      setPasswordVisible(false)
    },
    [form],
  )

  return (
    <Sheet open={isOpen} onOpenChange={handleOpenChange}>
      {trigger === undefined ? (
        <SheetTrigger asChild>
          <Button
            variant="default"
            size="sm"
            disabled={disabled}
            className={className}
            title={isEditMode ? 'Edit Registry' : 'Add Registry'}
          >
            {!isEditMode && <Plus className="w-4 h-4" />}
            {isEditMode ? 'Edit Registry' : 'Add Registry'}
          </Button>
        </SheetTrigger>
      ) : (
        trigger
      )}
      <SheetContent className="w-dvw sm:w-[460px] p-0 flex flex-col gap-0">
        <SheetHeader className="border-b border-border p-4 px-5 items-center flex text-left flex-row">
          <SheetTitle className="text-2xl">{isEditMode ? 'Edit Registry' : 'Add Registry'}</SheetTitle>
          <SheetDescription className="sr-only">
            Registry details must be provided for images that are not publicly available.
          </SheetDescription>
        </SheetHeader>

        <ScrollArea fade="mask" className="flex-1 min-h-0">
          <form
            ref={formRef}
            id={isEditMode ? 'edit-registry-form' : 'create-registry-form'}
            className="space-y-6 p-5"
            onSubmit={(e) => {
              e.preventDefault()
              e.stopPropagation()
              form.handleSubmit()
            }}
          >
            {!isEditMode && (
              <Tabs value={provider} onValueChange={(v) => handleProviderChange(v as RegistryProvider)}>
                <TabsList className="bg-muted w-full [&>*]:flex-1">
                  {REGISTRY_PROVIDER_VALUES.map((p) => (
                    <TabsTrigger
                      key={p}
                      value={p}
                      title={REGISTRY_PROVIDER_LABELS[p]}
                      aria-label={REGISTRY_PROVIDER_LABELS[p]}
                    >
                      {REGISTRY_PROVIDER_TAB_CONTENT[p]}
                    </TabsTrigger>
                  ))}
                </TabsList>
              </Tabs>
            )}

            <form.Field name="name">
              {(field) => {
                const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                return (
                  <Field data-invalid={isInvalid}>
                    <FieldLabel htmlFor={field.name}>Registry Name</FieldLabel>
                    <Input
                      aria-invalid={isInvalid}
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      placeholder="My Registry"
                    />
                    {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                      <FieldError errors={field.state.meta.errors} />
                    )}
                  </Field>
                )
              }}
            </form.Field>

            {!activeSpec.url.hidden && (
              <form.Field name="url">
                {(field) => {
                  const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                  return (
                    <Field data-invalid={isInvalid}>
                      <FieldLabel htmlFor={field.name}>{activeSpec.url.label}</FieldLabel>
                      <Input
                        aria-invalid={isInvalid}
                        id={field.name}
                        name={field.name}
                        value={field.state.value}
                        onBlur={field.handleBlur}
                        onChange={(e) => field.handleChange(e.target.value)}
                        placeholder={activeSpec.url.placeholder}
                      />
                      {activeSpec.url.helper && <FieldDescription>{activeSpec.url.helper}</FieldDescription>}
                      {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                        <FieldError errors={field.state.meta.errors} />
                      )}
                    </Field>
                  )
                }}
              </form.Field>
            )}

            {!activeSpec.username.hidden && (
              <form.Field name="username">
                {(field) => {
                  const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                  return (
                    <Field data-invalid={isInvalid}>
                      <FieldLabel htmlFor={field.name}>{activeSpec.username.label}</FieldLabel>
                      <Input
                        aria-invalid={isInvalid}
                        id={field.name}
                        name={field.name}
                        value={field.state.value}
                        onBlur={field.handleBlur}
                        onChange={(e) => field.handleChange(e.target.value)}
                        placeholder={activeSpec.username.placeholder}
                      />
                      {activeSpec.username.helper && <FieldDescription>{activeSpec.username.helper}</FieldDescription>}
                      {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                        <FieldError errors={field.state.meta.errors} />
                      )}
                    </Field>
                  )
                }}
              </form.Field>
            )}

            {!activeSpec.password.hidden && (
              <form.Field name="password">
                {(field) => {
                  const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                  return (
                    <Field data-invalid={isInvalid}>
                      <FieldLabel htmlFor={field.name}>{activeSpec.password.label}</FieldLabel>
                      {activeSpec.password.multiline ? (
                        <Textarea
                          aria-invalid={isInvalid}
                          id={field.name}
                          name={field.name}
                          rows={8}
                          className="font-mono text-xs"
                          value={field.state.value}
                          onBlur={field.handleBlur}
                          onChange={(e) => field.handleChange(e.target.value)}
                          placeholder={activeSpec.password.placeholder}
                        />
                      ) : (
                        <InputGroup className="pr-1 flex-1">
                          <InputGroupInput
                            aria-invalid={isInvalid}
                            id={field.name}
                            name={field.name}
                            type={passwordVisible ? 'text' : 'password'}
                            value={field.state.value}
                            onBlur={field.handleBlur}
                            onChange={(e) => field.handleChange(e.target.value)}
                          />
                          <InputGroupButton
                            variant="ghost"
                            size="icon-xs"
                            onClick={() => setPasswordVisible((visible) => !visible)}
                            aria-label={passwordVisible ? 'Hide password' : 'Show password'}
                          >
                            {passwordVisible ? <EyeOffIcon className="h-4 w-4" /> : <EyeIcon className="h-4 w-4" />}
                          </InputGroupButton>
                        </InputGroup>
                      )}
                      {isEditMode && <FieldDescription>Leave empty to keep the current password.</FieldDescription>}
                      {activeSpec.password.helper && <FieldDescription>{activeSpec.password.helper}</FieldDescription>}
                      {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                        <FieldError errors={field.state.meta.errors} />
                      )}
                    </Field>
                  )
                }}
              </form.Field>
            )}

            {!activeSpec.project.hidden && (
              <form.Field name="project">
                {(field) => {
                  const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                  return (
                    <Field data-invalid={isInvalid}>
                      <FieldLabel htmlFor={field.name}>{activeSpec.project.label}</FieldLabel>
                      <Input
                        aria-invalid={isInvalid}
                        id={field.name}
                        name={field.name}
                        value={field.state.value}
                        onBlur={field.handleBlur}
                        onChange={(e) => field.handleChange(e.target.value)}
                        placeholder={activeSpec.project.placeholder}
                      />
                      {activeSpec.project.helper && <FieldDescription>{activeSpec.project.helper}</FieldDescription>}
                      {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                        <FieldError errors={field.state.meta.errors} />
                      )}
                    </Field>
                  )
                }}
              </form.Field>
            )}
          </form>
        </ScrollArea>

        <SheetFooter className="border-t border-border p-4 px-5">
          <Button type="button" variant="secondary" onClick={() => handleOpenChange(false)}>
            Cancel
          </Button>
          <form.Subscribe
            selector={(state) => [state.canSubmit, state.isSubmitting]}
            children={([canSubmit, isSubmitting]) => (
              <Button
                type="submit"
                form={isEditMode ? 'edit-registry-form' : 'create-registry-form'}
                variant="default"
                disabled={
                  !canSubmit ||
                  isSubmitting ||
                  !selectedOrganization?.id ||
                  (isEditMode ? updateRegistryMutation.isPending : createRegistryMutation.isPending)
                }
              >
                {isSubmitting && <Spinner />}
                {isEditMode ? 'Edit' : 'Add'}
              </Button>
            )}
          />
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
