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
import { Switch } from '@/components/ui/switch'
import { useCreateSandboxMutation } from '@/hooks/mutations/useCreateSandboxMutation'
import { useSnapshotsQuery } from '@/hooks/queries/useSnapshotsQuery'
import { useRegions } from '@/hooks/useRegions'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { getRegionFullDisplayName } from '@/lib/utils'
import { useForm } from '@tanstack/react-form'
import { Minus, Plus } from 'lucide-react'
import { useCallback, useEffect, useRef, useState } from 'react'
import { toast } from 'sonner'
import { z } from 'zod'
import { ScrollArea } from '../ui/scroll-area'

const NAME_REGEX = /^[a-zA-Z0-9][a-zA-Z0-9._-]*$/

const keyValuePairSchema = z.object({
  key: z.string(),
  value: z.string(),
})

const formSchema = z.object({
  name: z
    .string()
    .optional()
    .refine((val) => !val || NAME_REGEX.test(val), 'Only letters, digits, dots, underscores and dashes are allowed'),
  snapshot: z.string().optional(),
  regionId: z.string().optional(),
  cpu: z.number().min(1).optional(),
  memory: z.number().min(1).optional(),
  disk: z.number().min(1).optional(),
  autoStopInterval: z.number().min(0).optional(),
  autoArchiveInterval: z.number().min(0).optional(),
  autoDeleteInterval: z.number().optional(),
  envVars: z.array(keyValuePairSchema).optional(),
  labels: z.array(keyValuePairSchema).optional(),
  public: z.boolean().optional(),
  networkBlockAll: z.boolean().optional(),
})

type FormValues = z.infer<typeof formSchema>

const defaultValues: FormValues = {
  name: '',
  snapshot: undefined,
  regionId: undefined,
  cpu: undefined,
  memory: undefined,
  disk: undefined,
  autoStopInterval: undefined,
  autoArchiveInterval: undefined,
  autoDeleteInterval: undefined,
  envVars: [],
  labels: [],
  public: false,
  networkBlockAll: false,
}

export const CreateSandboxDialog = ({ className }: { className?: string }) => {
  const [open, setOpen] = useState(false)

  const { availableRegions: regions, loadingAvailableRegions: loadingRegions } = useRegions()
  const { selectedOrganization } = useSelectedOrganization()
  const { reset: resetCreateSandboxMutation, ...createSandboxMutation } = useCreateSandboxMutation()
  const formRef = useRef<HTMLFormElement>(null)

  const { data: snapshotsData, isLoading: snapshotsLoading } = useSnapshotsQuery({
    page: 1,
    pageSize: 100,
  })

  const form = useForm({
    defaultValues,
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
        toast.error('Select an organization to create a sandbox.')
        return
      }

      const env: Record<string, string> = {}
      value.envVars?.forEach(({ key, value: val }) => {
        if (key) env[key] = val
      })

      const labels: Record<string, string> = {}
      value.labels?.forEach(({ key, value: val }) => {
        if (key) labels[key] = val
      })

      try {
        await createSandboxMutation.mutateAsync({
          sandbox: {
            name: value.name?.trim() || undefined,
            snapshot: value.snapshot || undefined,
            target: value.regionId || undefined,
            cpu: value.cpu,
            memory: value.memory,
            disk: value.disk,
            autoStopInterval: value.autoStopInterval,
            autoArchiveInterval: value.autoArchiveInterval,
            autoDeleteInterval: value.autoDeleteInterval,
            env: Object.keys(env).length > 0 ? env : undefined,
            labels: Object.keys(labels).length > 0 ? labels : undefined,
            public: value.public || undefined,
            networkBlockAll: value.networkBlockAll || undefined,
          },
          organizationId: selectedOrganization.id,
        })

        toast.success(`Creating sandbox${value.name ? ` ${value.name.trim()}` : ''}`)
        setOpen(false)
      } catch (error) {
        handleApiError(error, 'Failed to create sandbox')
      }
    },
  })

  const resetState = useCallback(() => {
    form.reset(defaultValues)
    resetCreateSandboxMutation()
  }, [resetCreateSandboxMutation, form])

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
        <Button variant="default" size="sm" className="ml-auto" title="Create Sandbox">
          <Plus className="w-4 h-4" />
          Create Sandbox
        </Button>
      </DialogTrigger>
      <DialogContent className={className}>
        <DialogHeader>
          <DialogTitle>Create New Sandbox</DialogTitle>
          <DialogDescription>Create a new sandbox in your organization.</DialogDescription>
        </DialogHeader>
        <ScrollArea fade="mask" className="h-[500px] -mx-5">
          <form
            ref={formRef}
            id="create-sandbox-form"
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
                    <FieldLabel htmlFor={field.name}>Name</FieldLabel>
                    <Input
                      aria-invalid={isInvalid}
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      placeholder="my-sandbox"
                    />
                    <FieldDescription>
                      Optional. If not provided, the sandbox ID will be used as the name.
                    </FieldDescription>
                    {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                      <FieldError errors={field.state.meta.errors} />
                    )}
                  </Field>
                )
              }}
            </form.Field>

            <form.Field name="snapshot">
              {(field) => (
                <Field>
                  <FieldLabel htmlFor={field.name}>Snapshot</FieldLabel>
                  <Select value={field.state.value} onValueChange={field.handleChange}>
                    <SelectTrigger
                      className="h-8"
                      id={field.name}
                      disabled={snapshotsLoading}
                      loading={snapshotsLoading}
                    >
                      <SelectValue placeholder={snapshotsLoading ? 'Loading snapshots...' : 'Select a snapshot'} />
                    </SelectTrigger>
                    <SelectContent>
                      {snapshotsData?.items?.map((snapshot) => (
                        <SelectItem key={snapshot.id} value={snapshot.name}>
                          {snapshot.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FieldDescription>
                    The snapshot to use for the sandbox. If not specified, the default snapshot will be used.
                  </FieldDescription>
                </Field>
              )}
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
                    The region where the sandbox will be created. If not specified, your organization's default region
                    will be used.
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

            <div className="flex flex-col gap-2">
              <Label className="text-sm font-medium">Lifecycle</Label>
              <div className="flex flex-col gap-2">
                <form.Field name="autoStopInterval">
                  {(field) => (
                    <div className="flex items-center gap-4">
                      <Label htmlFor={field.name} className="w-40 flex-shrink-0">
                        Auto-stop (min):
                      </Label>
                      <Input
                        id={field.name}
                        type="number"
                        className="w-full"
                        min="0"
                        placeholder="15"
                        value={field.state.value ?? ''}
                        onChange={(e) => field.handleChange(parseInt(e.target.value) || undefined)}
                      />
                    </div>
                  )}
                </form.Field>
                <form.Field name="autoArchiveInterval">
                  {(field) => (
                    <div className="flex items-center gap-4">
                      <Label htmlFor={field.name} className="w-40 flex-shrink-0">
                        Auto-archive (min):
                      </Label>
                      <Input
                        id={field.name}
                        type="number"
                        className="w-full"
                        min="0"
                        placeholder="10080"
                        value={field.state.value ?? ''}
                        onChange={(e) => field.handleChange(parseInt(e.target.value) || undefined)}
                      />
                    </div>
                  )}
                </form.Field>
                <form.Field name="autoDeleteInterval">
                  {(field) => (
                    <div className="flex items-center gap-4">
                      <Label htmlFor={field.name} className="w-40 flex-shrink-0">
                        Auto-delete (min):
                      </Label>
                      <Input
                        id={field.name}
                        type="number"
                        className="w-full"
                        placeholder="Disabled"
                        value={field.state.value ?? ''}
                        onChange={(e) => {
                          const val = e.target.value
                          field.handleChange(val === '' ? undefined : parseInt(val))
                        }}
                      />
                    </div>
                  )}
                </form.Field>
              </div>
              <ul className="list-disc list-inside space-y-0.5 text-muted-foreground text-xs">
                <li>Auto-stop: minutes of inactivity before stopping (0 = disabled, default 15)</li>
                <li>Auto-archive: minutes after stopping before archiving (0 = max, default 7 days)</li>
                <li>Auto-delete: minutes after stopping before deletion (negative = disabled)</li>
              </ul>
            </div>

            <form.Field name="envVars">
              {(field) => (
                <Field>
                  <FieldLabel>Environment Variables</FieldLabel>
                  <div className="flex flex-col gap-2">
                    {(field.state.value ?? []).map((_, index) => (
                      <div key={index} className="flex items-center gap-2">
                        <Input
                          placeholder="Key"
                          value={field.state.value?.[index]?.key ?? ''}
                          onChange={(e) => {
                            const updated = [...(field.state.value ?? [])]
                            updated[index] = { ...updated[index], key: e.target.value }
                            field.handleChange(updated)
                          }}
                        />
                        <Input
                          placeholder="Value"
                          value={field.state.value?.[index]?.value ?? ''}
                          onChange={(e) => {
                            const updated = [...(field.state.value ?? [])]
                            updated[index] = { ...updated[index], value: e.target.value }
                            field.handleChange(updated)
                          }}
                        />
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="flex-shrink-0 h-8 w-8"
                          onClick={() => {
                            const updated = (field.state.value ?? []).filter((_, i) => i !== index)
                            field.handleChange(updated)
                          }}
                        >
                          <Minus className="w-4 h-4" />
                        </Button>
                      </div>
                    ))}
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      className="w-fit"
                      onClick={() => field.handleChange([...(field.state.value ?? []), { key: '', value: '' }])}
                    >
                      <Plus className="w-4 h-4" />
                      Add Variable
                    </Button>
                  </div>
                </Field>
              )}
            </form.Field>

            <form.Field name="labels">
              {(field) => (
                <Field>
                  <FieldLabel>Labels</FieldLabel>
                  <div className="flex flex-col gap-2">
                    {(field.state.value ?? []).map((_, index) => (
                      <div key={index} className="flex items-center gap-2">
                        <Input
                          placeholder="Key"
                          value={field.state.value?.[index]?.key ?? ''}
                          onChange={(e) => {
                            const updated = [...(field.state.value ?? [])]
                            updated[index] = { ...updated[index], key: e.target.value }
                            field.handleChange(updated)
                          }}
                        />
                        <Input
                          placeholder="Value"
                          value={field.state.value?.[index]?.value ?? ''}
                          onChange={(e) => {
                            const updated = [...(field.state.value ?? [])]
                            updated[index] = { ...updated[index], value: e.target.value }
                            field.handleChange(updated)
                          }}
                        />
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="flex-shrink-0 h-8 w-8"
                          onClick={() => {
                            const updated = (field.state.value ?? []).filter((_, i) => i !== index)
                            field.handleChange(updated)
                          }}
                        >
                          <Minus className="w-4 h-4" />
                        </Button>
                      </div>
                    ))}
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      className="w-fit"
                      onClick={() => field.handleChange([...(field.state.value ?? []), { key: '', value: '' }])}
                    >
                      <Plus className="w-4 h-4" />
                      Add Label
                    </Button>
                  </div>
                </Field>
              )}
            </form.Field>

            <div className="flex flex-col gap-4">
              <Label className="text-sm font-medium">Network</Label>
              <form.Field name="public">
                {(field) => (
                  <div className="flex items-center justify-between">
                    <div className="flex flex-col gap-1">
                      <Label htmlFor={field.name}>Public HTTP Preview</Label>
                      <FieldDescription>Allow public access to HTTP preview URLs.</FieldDescription>
                    </div>
                    <Switch id={field.name} checked={field.state.value ?? false} onCheckedChange={field.handleChange} />
                  </div>
                )}
              </form.Field>
              <form.Field name="networkBlockAll">
                {(field) => (
                  <div className="flex items-center justify-between">
                    <div className="flex flex-col gap-1">
                      <Label htmlFor={field.name}>Block All Network Access</Label>
                      <FieldDescription>Block all outbound network access from the sandbox.</FieldDescription>
                    </div>
                    <Switch id={field.name} checked={field.state.value ?? false} onCheckedChange={field.handleChange} />
                  </div>
                )}
              </form.Field>
            </div>
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
                form="create-sandbox-form"
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
