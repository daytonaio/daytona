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

type FormValues = z.infer<typeof createFormSchema>

const defaultValues: FormValues = {
  name: '',
  url: '',
  username: '',
  password: '',
  project: '',
}

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

  const isEditMode = mode === 'edit'
  const isControlled = open !== undefined
  const isOpen = open ?? internalOpen

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
      return defaultValues
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
      onSubmit: isEditMode ? editFormSchema : createFormSchema,
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

      try {
        if (isEditMode) {
          if (!registry) {
            toast.error('No registry selected for editing.')
            return
          }

          await updateRegistryMutation.mutateAsync({
            registryId: registry.id,
            registry: {
              name: value.name.trim(),
              url: value.url.trim() || 'docker.io',
              username: value.username.trim(),
              password: value.password.trim(),
              project: value.project.trim(),
            },
            organizationId: selectedOrganization.id,
          })
          toast.success('Registry edited successfully')
        } else {
          await createRegistryMutation.mutateAsync({
            registry: {
              name: value.name.trim(),
              url: value.url.trim() || 'docker.io',
              username: value.username.trim(),
              password: value.password.trim(),
              project: value.project.trim(),
            },
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

            <form.Field name="url">
              {(field) => {
                const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                return (
                  <Field data-invalid={isInvalid}>
                    <FieldLabel htmlFor={field.name}>Registry URL</FieldLabel>
                    <Input
                      aria-invalid={isInvalid}
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      placeholder="https://registry.example.com"
                    />
                    <FieldDescription>Defaults to docker.io when left blank.</FieldDescription>
                    {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                      <FieldError errors={field.state.meta.errors} />
                    )}
                  </Field>
                )
              }}
            </form.Field>

            <form.Field name="username">
              {(field) => {
                const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                return (
                  <Field data-invalid={isInvalid}>
                    <FieldLabel htmlFor={field.name}>Username</FieldLabel>
                    <Input
                      aria-invalid={isInvalid}
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                    />
                    {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                      <FieldError errors={field.state.meta.errors} />
                    )}
                  </Field>
                )
              }}
            </form.Field>

            <form.Field name="password">
              {(field) => {
                const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                return (
                  <Field data-invalid={isInvalid}>
                    <FieldLabel htmlFor={field.name}>Password</FieldLabel>
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
                    {isEditMode && <FieldDescription>Leave empty to keep the current password.</FieldDescription>}
                    {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                      <FieldError errors={field.state.meta.errors} />
                    )}
                  </Field>
                )
              }}
            </form.Field>

            <form.Field name="project">
              {(field) => {
                const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                return (
                  <Field data-invalid={isInvalid}>
                    <FieldLabel htmlFor={field.name}>Project</FieldLabel>
                    <Input
                      aria-invalid={isInvalid}
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      placeholder="my-project"
                    />
                    <FieldDescription>Leave empty for private Docker Hub entries.</FieldDescription>
                    {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                      <FieldError errors={field.state.meta.errors} />
                    )}
                  </Field>
                )
              }}
            </form.Field>
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
