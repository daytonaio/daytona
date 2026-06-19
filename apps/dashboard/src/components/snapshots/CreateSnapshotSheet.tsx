/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateResourceButton } from '@/components/CreateResourceButton'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Field, FieldDescription, FieldError, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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
import { useCreateSnapshotMutation } from '@/hooks/mutations/useCreateSnapshotMutation'
import { useOrganizationUsageOverviewQuery } from '@/hooks/queries/useOrganizationUsageOverviewQuery'
import { useAvailableRegionsQuery } from '@/hooks/queries/useRegionsQuery'
import { useAvailableSandboxClasses } from '@/hooks/useAvailableSandboxClasses'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { GPU_TYPE_LABELS } from '@/lib/gpu-types'
import { EMPTY_REGIONS } from '@/lib/regions'
import { imageNameSchema } from '@/lib/schema'
import { cn, getRegionFullDisplayName } from '@/lib/utils'
import type { SnapshotDto } from '@daytona/api-client'
import { GpuType, SandboxClass } from '@daytona/api-client'
import { useForm, useStore } from '@tanstack/react-form'
import { Ref, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState } from 'react'
import { toast } from 'sonner'
import { z } from 'zod'
import { ScrollArea } from '../ui/scroll-area'

const IMAGE_NAME_REGEX = /^[a-zA-Z0-9_.\-:]+(\/[a-zA-Z0-9_.\-:]+)*(@sha256:[a-f0-9]{64})?$/

const snapshotNameSchema = z
  .string()
  .min(1, 'Snapshot name is required')
  .refine((name) => IMAGE_NAME_REGEX.test(name), 'Only letters, digits, dots, colons, slashes and dashes are allowed')

const SANDBOX_CLASS_OPTIONS: { value: SandboxClass; label: string }[] = [
  { value: SandboxClass.CONTAINER, label: 'Container' },
  { value: SandboxClass.LINUX_VM, label: 'Linux VM' },
  { value: SandboxClass.ANDROID, label: 'Android' },
]

const SELECTABLE_GPU_TYPES = (Object.values(GpuType) as GpuType[]).filter((t) => t !== GpuType.UNKNOWN_DEFAULT_OPEN_API)

const resolveAllowedGpuTypes = (regionAllowed: GpuType[] | null | undefined): GpuType[] => {
  const filteredRegion = (regionAllowed ?? []).filter((t) => t !== GpuType.UNKNOWN_DEFAULT_OPEN_API)
  return filteredRegion.length > 0 ? filteredRegion : SELECTABLE_GPU_TYPES
}

const formSchema = z.object({
  name: snapshotNameSchema,
  imageName: imageNameSchema,
  entrypoint: z.string().optional(),
  cpu: z.number().min(1).optional(),
  memory: z.number().min(1).optional(),
  disk: z.number().min(1).optional(),
  gpu: z.boolean().optional(),
  gpuType: z.nativeEnum(GpuType).optional(),
  regionId: z.string().optional(),
  sandboxClass: z.nativeEnum(SandboxClass).optional(),
})

type FormValues = z.infer<typeof formSchema>

const defaultValues: FormValues = {
  name: '',
  imageName: '',
  entrypoint: '',
  cpu: undefined,
  memory: undefined,
  disk: undefined,
  gpu: false,
  gpuType: undefined,
  regionId: undefined,
  sandboxClass: SandboxClass.CONTAINER,
}

export const CreateSnapshotSheet = ({
  className,
  onSnapshotCreated,
  ref,
}: {
  className?: string
  onSnapshotCreated?: (snapshot: SnapshotDto) => void
  ref?: Ref<{ open: () => void }>
}) => {
  const [open, setOpen] = useState(false)

  const { selectedOrganization } = useSelectedOrganization()
  const { data: regions = EMPTY_REGIONS, isLoading: loadingRegions } = useAvailableRegionsQuery(
    selectedOrganization?.id,
  )
  const { reset: resetCreateSnapshotMutation, ...createSnapshotMutation } = useCreateSnapshotMutation()
  const formRef = useRef<HTMLFormElement>(null)
  const { data: usageOverview } = useOrganizationUsageOverviewQuery({
    organizationId: selectedOrganization?.id || '',
  })
  const formDefaultValues = useMemo<FormValues>(
    () => ({
      ...defaultValues,
      regionId: selectedOrganization?.defaultRegionId,
    }),
    [selectedOrganization?.defaultRegionId],
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
      const form = formRef.current
      if (!form) return
      const invalidInput = form.querySelector('[aria-invalid="true"]') as HTMLInputElement | null
      if (invalidInput) {
        invalidInput.scrollIntoView({ behavior: 'smooth', block: 'center' })
        invalidInput.focus()
      }
    },
    onSubmit: async ({ value }) => {
      if (!selectedOrganization?.id) {
        toast.error('Select an organization to create a snapshot.')
        return
      }

      const trimmedEntrypoint = value.entrypoint?.trim()

      try {
        const snapshot = await createSnapshotMutation.mutateAsync({
          snapshot: {
            name: value.name.trim(),
            imageName: value.imageName.trim(),
            entrypoint: trimmedEntrypoint ? trimmedEntrypoint.split(' ') : undefined,
            cpu: value.cpu,
            memory: value.memory,
            disk: value.disk,
            gpu: value.gpu ? 1 : undefined,
            gpuType: value.gpu && value.gpuType ? [value.gpuType] : undefined,
            regionId: value.regionId,
            sandboxClass: value.sandboxClass,
          },
          organizationId: selectedOrganization.id,
        })

        toast.success(`Creating snapshot ${value.name.trim()}`)
        onSnapshotCreated?.(snapshot)
        setOpen(false)
      } catch (error) {
        handleApiError(error, 'Failed to create snapshot')
      }
    },
  })
  const { reset: resetForm } = form

  const resetState = useCallback(() => {
    resetForm(formDefaultValues)
    resetCreateSnapshotMutation()
  }, [formDefaultValues, resetForm, resetCreateSnapshotMutation])

  useEffect(() => {
    if (open) {
      resetState()
    }
  }, [open, resetState])

  const selectedRegionId = useStore(form.store, (state) => state.values.regionId)
  const availableSandboxClasses = useAvailableSandboxClasses(selectedRegionId)

  useEffect(() => {
    if (availableSandboxClasses.length === 0) return
    const current = form.getFieldValue('sandboxClass')
    if (current && availableSandboxClasses.includes(current)) return
    const preferred = availableSandboxClasses.includes(SandboxClass.CONTAINER)
      ? SandboxClass.CONTAINER
      : availableSandboxClasses[0]
    form.setFieldValue('sandboxClass', preferred)
  }, [availableSandboxClasses, form])

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <CreateResourceButton resource="Snapshot" />
      </SheetTrigger>
      <SheetContent className={cn('w-dvw sm:w-[500px] p-0 flex flex-col gap-0', className)}>
        <SheetHeader className="border-b border-border p-4 px-5 items-center flex text-left flex-row">
          <SheetTitle>Create Snapshot</SheetTitle>
          <SheetDescription className="sr-only">
            Register a new snapshot to be used for spinning up sandboxes in your organization.
          </SheetDescription>
        </SheetHeader>
        <ScrollArea fade="mask" className="flex-1 min-h-0">
          <form
            ref={formRef}
            id="create-snapshot-form"
            className="gap-6 flex flex-col p-5"
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
                    <FieldLabel htmlFor={field.name}>Snapshot Name</FieldLabel>
                    <Input
                      aria-invalid={isInvalid}
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      placeholder="ubuntu-4vcpu-8ram-100gb"
                    />
                    <FieldDescription>
                      The name you will use in your client app (SDK, CLI) to reference the snapshot.
                    </FieldDescription>
                    {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                      <FieldError errors={field.state.meta.errors} />
                    )}
                  </Field>
                )
              }}
            </form.Field>

            <form.Field name="imageName">
              {(field) => {
                const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                return (
                  <Field data-invalid={isInvalid}>
                    <FieldLabel htmlFor={field.name}>Image</FieldLabel>
                    <Input
                      aria-invalid={isInvalid}
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      placeholder="ubuntu:22.04"
                    />
                    <FieldDescription>
                      Must include either a tag (e.g., ubuntu:22.04) or a digest. The tag "latest" is not allowed.
                    </FieldDescription>
                    {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                      <FieldError errors={field.state.meta.errors} />
                    )}
                  </Field>
                )
              }}
            </form.Field>

            <form.Field name="regionId">
              {(field) => (
                <Field>
                  <FieldLabel htmlFor={field.name}>Region</FieldLabel>
                  <Select value={field.state.value} onValueChange={field.handleChange}>
                    <SelectTrigger className="h-8" id={field.name} disabled={loadingRegions} loading={loadingRegions}>
                      <SelectValue placeholder={loadingRegions ? 'Loading regions...' : 'Select a region'} />
                    </SelectTrigger>
                    <SelectContent>
                      {regions.map((region) => (
                        <SelectItem key={region.id} value={region.id}>
                          {getRegionFullDisplayName(region)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FieldDescription>The region where the snapshot will be available.</FieldDescription>
                </Field>
              )}
            </form.Field>

            {availableSandboxClasses.length > 1 && (
              <form.Field name="sandboxClass">
                {(field) => (
                  <Field>
                    <FieldLabel htmlFor={field.name}>Sandbox Class</FieldLabel>
                    <Select
                      value={field.state.value ?? SandboxClass.CONTAINER}
                      onValueChange={(value) => field.handleChange(value as SandboxClass)}
                      disabled={!selectedRegionId}
                    >
                      <SelectTrigger className="h-8" id={field.name}>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {SANDBOX_CLASS_OPTIONS.filter((option) => availableSandboxClasses.includes(option.value)).map(
                          (option) => (
                            <SelectItem key={option.value} value={option.value}>
                              {option.label}
                            </SelectItem>
                          ),
                        )}
                      </SelectContent>
                    </Select>
                    <FieldDescription>
                      The target platform sandboxes created from this snapshot will run on.
                    </FieldDescription>
                  </Field>
                )}
              </form.Field>
            )}

            <div className="flex flex-col gap-2">
              <Label className="text-sm font-medium">Resources</Label>
              <div className="flex flex-col gap-2">
                <form.Field name="cpu">
                  {(field) => (
                    <div className="flex items-center gap-4">
                      <Label htmlFor={field.name} className="w-32 flex-shrink-0">
                        Compute (vCPU):
                      </Label>
                      <Input
                        id={field.name}
                        type="number"
                        className="w-full"
                        min="1"
                        placeholder="1"
                        value={field.state.value ?? ''}
                        onChange={(e) => field.handleChange(parseInt(e.target.value) || undefined)}
                      />
                    </div>
                  )}
                </form.Field>
                <form.Field name="memory">
                  {(field) => (
                    <div className="flex items-center gap-4">
                      <Label htmlFor={field.name} className="w-32 flex-shrink-0">
                        Memory (GiB):
                      </Label>
                      <Input
                        id={field.name}
                        type="number"
                        className="w-full"
                        min="1"
                        placeholder="1"
                        value={field.state.value ?? ''}
                        onChange={(e) => field.handleChange(parseInt(e.target.value) || undefined)}
                      />
                    </div>
                  )}
                </form.Field>
                <form.Field name="disk">
                  {(field) => (
                    <div className="flex items-center gap-4">
                      <Label htmlFor={field.name} className="w-32 flex-shrink-0">
                        Storage (GiB):
                      </Label>
                      <Input
                        id={field.name}
                        type="number"
                        className="w-full"
                        min="1"
                        placeholder="3"
                        value={field.state.value ?? ''}
                        onChange={(e) => field.handleChange(parseInt(e.target.value) || undefined)}
                      />
                    </div>
                  )}
                </form.Field>
                <form.Subscribe selector={(state) => state.values.regionId}>
                  {(regionId) => {
                    const region = usageOverview?.regionUsage.find((r) => r.regionId === regionId)
                    if ((region?.totalGpuQuota ?? 0) <= 0) return null
                    const allowedGpuTypes = resolveAllowedGpuTypes(region?.allowedGpuTypes)
                    return (
                      <div className="flex flex-col gap-3">
                        <form.Field name="gpu">
                          {(field) => (
                            <div className="flex items-start gap-2 pt-1">
                              <Checkbox
                                id={field.name}
                                className="mt-0.5"
                                checked={field.state.value ?? false}
                                onCheckedChange={(checked) => {
                                  const isChecked = checked === true
                                  field.handleChange(isChecked)
                                  if (isChecked) {
                                    if (!form.getFieldValue('gpuType') && allowedGpuTypes.length > 0) {
                                      form.setFieldValue('gpuType', allowedGpuTypes[0])
                                    }
                                  } else {
                                    form.setFieldValue('gpuType', undefined)
                                  }
                                }}
                              />
                              <div className="flex flex-col gap-1">
                                <Label htmlFor={field.name} className="text-sm font-normal">
                                  Allocate GPU
                                </Label>
                                <FieldDescription>
                                  Sandboxes created from this snapshot must be ephemeral - set autoDeleteInterval to 0
                                  when creating the sandbox.
                                </FieldDescription>
                              </div>
                            </div>
                          )}
                        </form.Field>
                        <form.Subscribe selector={(state) => state.values.gpu}>
                          {(gpuEnabled) =>
                            gpuEnabled && allowedGpuTypes.length > 0 ? (
                              <form.Field name="gpuType">
                                {(field) => (
                                  <Field>
                                    <FieldLabel htmlFor={field.name}>GPU type</FieldLabel>
                                    <Select
                                      value={field.state.value ?? allowedGpuTypes[0]}
                                      onValueChange={(val) => field.handleChange(val as GpuType)}
                                    >
                                      <SelectTrigger className="h-8" id={field.name}>
                                        <SelectValue />
                                      </SelectTrigger>
                                      <SelectContent>
                                        {allowedGpuTypes.map((gpuType) => (
                                          <SelectItem key={gpuType} value={gpuType}>
                                            {GPU_TYPE_LABELS[gpuType] || gpuType}
                                          </SelectItem>
                                        ))}
                                      </SelectContent>
                                    </Select>
                                  </Field>
                                )}
                              </form.Field>
                            ) : null
                          }
                        </form.Subscribe>
                      </div>
                    )
                  }}
                </form.Subscribe>
              </div>
              <FieldDescription>
                If not specified, default values will be used (1 vCPU, 1 GiB memory, 3 GiB storage).
              </FieldDescription>
            </div>

            <form.Field name="entrypoint">
              {(field) => (
                <Field>
                  <FieldLabel htmlFor={field.name}>Entrypoint (optional)</FieldLabel>
                  <Input
                    id={field.name}
                    name={field.name}
                    value={field.state.value ?? ''}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    placeholder="sleep infinity"
                  />
                  <FieldDescription>
                    Ensure that the entrypoint is a long running command. If not provided, or if the snapshot does not
                    have an entrypoint, 'sleep infinity' will be used as the default.
                  </FieldDescription>
                </Field>
              )}
            </form.Field>
          </form>
        </ScrollArea>
        <SheetFooter className="border-t border-border p-4 px-5">
          <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <form.Subscribe
            selector={(state) => [state.canSubmit, state.isSubmitting]}
            children={([canSubmit, isSubmitting]) => (
              <Button
                type="submit"
                form="create-snapshot-form"
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
