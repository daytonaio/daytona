/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateResourceButton } from '@/components/CreateResourceButton'
import { Button } from '@/components/ui/button'
import { Field, FieldDescription, FieldError, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
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
import { Spinner } from '@/components/ui/spinner'
import { useCreateVolumeMutation } from '@/hooks/mutations/useCreateVolumeMutation'
import { useRegions } from '@/hooks/useRegions'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { FeatureFlags } from '@/enums/FeatureFlags'
import { handleApiError } from '@/lib/error-handling'
import { getRegionFullDisplayName } from '@/lib/utils'
import { CreateVolume, OrganizationDefaultVolumeBackendEnum } from '@daytona/api-client'
import { useForm } from '@tanstack/react-form'
import { useFeatureFlagEnabled } from 'posthog-js/react'
import { Ref, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState } from 'react'
import { toast } from 'sonner'
import { z } from 'zod'

const formSchema = z.object({
  name: z.string().trim().min(1, 'Volume name is required'),
  backend: z.string().optional(),
  regionId: z.string().optional(),
})

type FormValues = z.infer<typeof formSchema>

const defaultValues: FormValues = {
  name: '',
  backend: undefined,
  regionId: undefined,
}

const backendOptions: { value: string; label: string }[] = [
  { value: OrganizationDefaultVolumeBackendEnum.S3FUSE, label: 'Standard' },
  { value: OrganizationDefaultVolumeBackendEnum.LAYERED, label: 'Layered' },
]

export const CreateVolumeSheet = ({
  className,
  disabled,
  ref,
}: {
  className?: string
  disabled?: boolean
  ref?: Ref<{ open: () => void }>
}) => {
  const [open, setOpen] = useState(false)

  const { selectedOrganization } = useSelectedOrganization()
  const { availableRegions, loadingAvailableRegions } = useRegions()
  const { reset: resetCreateVolumeMutation, ...createVolumeMutation } = useCreateVolumeMutation()
  const formRef = useRef<HTMLFormElement>(null)

  const backendPickerEnabled = !!useFeatureFlagEnabled(FeatureFlags.VOLUME_BACKEND_PICKER)
  const orgDefaultBackend = selectedOrganization?.defaultVolumeBackend ?? OrganizationDefaultVolumeBackendEnum.S3FUSE
  const regionOptions = availableRegions

  // Picker lets the user choose; otherwise fall back to the org default.
  const initialBackend = backendPickerEnabled ? orgDefaultBackend : undefined
  const formDefaultValues = useMemo<FormValues>(
    () => ({
      ...defaultValues,
      backend: initialBackend,
      regionId:
        orgDefaultBackend === OrganizationDefaultVolumeBackendEnum.LAYERED
          ? selectedOrganization?.defaultRegionId
          : undefined,
    }),
    [initialBackend, orgDefaultBackend, selectedOrganization?.defaultRegionId],
  )

  useImperativeHandle(ref, () => ({
    open: () => setOpen(true),
  }))

  const form = useForm({
    defaultValues: formDefaultValues,
    validators: {
      onSubmit: formSchema,
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
        toast.error('Select an organization to create a volume.')
        return
      }

      try {
        const volumeName = value.name.trim()

        const effectiveBackend = backendPickerEnabled ? value.backend : undefined
        const isLayered = (effectiveBackend ?? orgDefaultBackend) === OrganizationDefaultVolumeBackendEnum.LAYERED

        await createVolumeMutation.mutateAsync({
          volume: {
            name: volumeName,
            backend: effectiveBackend as CreateVolume['backend'],
            regionId: isLayered ? value.regionId : undefined,
          },
          organizationId: selectedOrganization.id,
        })

        setOpen(false)
        toast.success(`Creating volume ${volumeName}`)
      } catch (error) {
        handleApiError(error, 'Failed to create volume')
      }
    },
  })
  const { reset: resetForm } = form

  const resetState = useCallback(() => {
    resetForm(formDefaultValues)
    resetCreateVolumeMutation()
  }, [resetForm, formDefaultValues, resetCreateVolumeMutation])

  useEffect(() => {
    if (open) {
      resetState()
    }
  }, [open, resetState])

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <CreateResourceButton resource="Volume" disabled={disabled} className={className} />
      </SheetTrigger>
      <SheetContent className="w-dvw sm:w-[420px] p-0 flex flex-col gap-0">
        <SheetHeader className="border-b border-border p-4 px-5 items-center flex text-left flex-row">
          <SheetTitle>Create Volume</SheetTitle>
          <SheetDescription className="sr-only">Create a new volume for shared, persistent storage.</SheetDescription>
        </SheetHeader>

        <ScrollArea fade="mask" className="flex-1 min-h-0">
          <form
            ref={formRef}
            id="create-volume-form"
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
                    <FieldLabel htmlFor={field.name}>Volume Name</FieldLabel>
                    <Input
                      aria-invalid={isInvalid}
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      placeholder="my-volume"
                    />
                    <FieldDescription>Used to mount this volume in your sandboxes.</FieldDescription>
                    {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                      <FieldError errors={field.state.meta.errors} />
                    )}
                  </Field>
                )
              }}
            </form.Field>

            {backendPickerEnabled && (
              <form.Field name="backend">
                {(field) => (
                  <Field>
                    <FieldLabel htmlFor={field.name}>Backend</FieldLabel>
                    <Select
                      value={field.state.value ?? orgDefaultBackend}
                      onValueChange={(v) => {
                        field.handleChange(v)
                        // Avoid sending a stale region for non-layered backends.
                        if (v !== OrganizationDefaultVolumeBackendEnum.LAYERED) {
                          form.setFieldValue('regionId', undefined)
                        } else if (!form.getFieldValue('regionId')) {
                          form.setFieldValue('regionId', selectedOrganization?.defaultRegionId)
                        }
                      }}
                    >
                      <SelectTrigger className="h-8" id={field.name}>
                        <SelectValue placeholder="Select a backend" />
                      </SelectTrigger>
                      <SelectContent>
                        {backendOptions.map((option) => (
                          <SelectItem key={option.value} value={option.value}>
                            {option.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FieldDescription>How the volume's data is stored and mounted.</FieldDescription>
                  </Field>
                )}
              </form.Field>
            )}

            <form.Subscribe selector={(state) => state.values.backend}>
              {(selectedBackend) => {
                const effectiveBackend = backendPickerEnabled
                  ? (selectedBackend ?? orgDefaultBackend)
                  : orgDefaultBackend
                if (effectiveBackend !== OrganizationDefaultVolumeBackendEnum.LAYERED) {
                  return null
                }
                return (
                  <form.Field name="regionId">
                    {(field) => (
                      <Field>
                        <FieldLabel htmlFor={field.name}>Region</FieldLabel>
                        <Select value={field.state.value ?? ''} onValueChange={field.handleChange}>
                          <SelectTrigger
                            className="h-8"
                            id={field.name}
                            disabled={loadingAvailableRegions}
                            loading={loadingAvailableRegions}
                          >
                            <SelectValue
                              placeholder={loadingAvailableRegions ? 'Loading regions...' : 'Select a region'}
                            />
                          </SelectTrigger>
                          <SelectContent>
                            {regionOptions.map((region) => (
                              <SelectItem key={region.id} value={region.id}>
                                {getRegionFullDisplayName(region)}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <FieldDescription>The region where the volume's data will be stored.</FieldDescription>
                      </Field>
                    )}
                  </form.Field>
                )
              }}
            </form.Subscribe>
          </form>
        </ScrollArea>

        <SheetFooter className="border-t border-border p-4 px-5">
          <Button type="button" size="sm" variant="secondary" onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <form.Subscribe
            selector={(state) => [state.canSubmit, state.isSubmitting]}
            children={([canSubmit, isSubmitting]) => (
              <Button
                type="submit"
                size="sm"
                form="create-volume-form"
                variant="default"
                disabled={!canSubmit || isSubmitting || !selectedOrganization?.id}
              >
                {isSubmitting && <Spinner />}
                Create
              </Button>
            )}
          />
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
