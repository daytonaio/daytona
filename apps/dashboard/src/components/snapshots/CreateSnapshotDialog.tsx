/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Field, FieldDescription, FieldError, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Spinner } from '@/components/ui/spinner'
import { useCreateSnapshotMutation } from '@/hooks/mutations/useCreateSnapshotMutation'
import { useRegions } from '@/hooks/useRegions'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { getRegionFullDisplayName } from '@/lib/utils'
import { useForm } from '@tanstack/react-form'
import { Plus } from 'lucide-react'
import { Ref, useCallback, useEffect, useImperativeHandle, useRef, useState } from 'react'
import { toast } from 'sonner'
import { z } from 'zod'
import { ScrollArea } from '../ui/scroll-area'

const IMAGE_NAME_REGEX = /^[a-zA-Z0-9_.\-:]+(\/[a-zA-Z0-9_.\-:]+)*(@sha256:[a-f0-9]{64})?$/
const IMAGE_TAG_OR_DIGEST_REGEX = /^[^@]+@sha256:[a-f0-9]{64}$|^(?!.*@sha256:).*:.+$/

const snapshotNameSchema = z
  .string()
  .min(1, 'Snapshot name is required')
  .refine((name) => IMAGE_NAME_REGEX.test(name), 'Only letters, digits, dots, colons, slashes and dashes are allowed')

const imageNameSchema = z
  .string()
  .min(1, 'Image name is required')
  .refine((name) => IMAGE_NAME_REGEX.test(name), 'Only letters, digits, dots, colons, slashes and dashes are allowed')
  .refine(
    (name) => IMAGE_TAG_OR_DIGEST_REGEX.test(name),
    'Image must include a tag (e.g., ubuntu:22.04) or digest (@sha256:...)',
  )
  .refine((name) => !name.endsWith(':latest'), 'Images with tag ":latest" are not allowed')

const formSchema = z.object({
  name: snapshotNameSchema,
  imageName: imageNameSchema,
  entrypoint: z.string().optional(),
  cpu: z.number().min(1).optional(),
  memory: z.number().min(1).optional(),
  disk: z.number().min(1).optional(),
  regionId: z.string().optional(),
})

type FormValues = z.infer<typeof formSchema>

const defaultValues: FormValues = {
  name: '',
  imageName: '',
  entrypoint: '',
  cpu: undefined,
  memory: undefined,
  disk: undefined,
  regionId: undefined,
}

export const CreateSnapshotDialog = ({ className, ref }: { className?: string; ref?: Ref<{ open: () => void }> }) => {
  const [open, setOpen] = useState(false)

  const { availableRegions: regions, loadingAvailableRegions: loadingRegions } = useRegions()
  const { selectedOrganization } = useSelectedOrganization()
  const { reset: resetCreateSnapshotMutation, ...createSnapshotMutation } = useCreateSnapshotMutation()
  const formRef = useRef<HTMLFormElement>(null)

  useImperativeHandle(ref, () => ({
    open: () => setOpen(true),
  }))

  const form = useForm({
    defaultValues,
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
        await createSnapshotMutation.mutateAsync({
          snapshot: {
            name: value.name.trim(),
            imageName: value.imageName.trim(),
            entrypoint: trimmedEntrypoint ? trimmedEntrypoint.split(' ') : undefined,
            cpu: value.cpu,
            memory: value.memory,
            disk: value.disk,
            regionId: value.regionId,
          },
          organizationId: selectedOrganization.id,
        })

        toast.success(`Creating snapshot ${value.name.trim()}`)
        setOpen(false)
      } catch (error) {
        handleApiError(error, 'Failed to create snapshot')
      }
    },
  })

  const resetState = useCallback(() => {
    form.reset(defaultValues)
    resetCreateSnapshotMutation()
  }, [resetCreateSnapshotMutation, form])

  useEffect(() => {
    if (open) {
      resetState()
    }
  }, [open, resetState])

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        setOpen(isOpen)
      }}
    >
      <DialogTrigger asChild>
        <Button variant="default" size="sm" className="ml-auto" title="Create Snapshot">
          <Plus className="w-4 h-4" />
          Create Snapshot
        </Button>
      </DialogTrigger>
      <DialogContent className={className}>
        <DialogHeader>
          <DialogTitle>Create New Snapshot</DialogTitle>
          <DialogDescription>
            Register a new snapshot to be used for spinning up sandboxes in your organization.
          </DialogDescription>
        </DialogHeader>
        <ScrollArea fade="mask" className="h-[500px] overflow-auto -mx-5">
          <form
            ref={formRef}
            id="create-snapshot-form"
            className="gap-6 flex flex-col px-5"
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
                  <FieldDescription>
                    The region where the snapshot will be available. If not specified, your organization's default
                    region will be used.
                  </FieldDescription>
                </Field>
              )}
            </form.Field>

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
        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="secondary">
              Cancel
            </Button>
          </DialogClose>
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
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
