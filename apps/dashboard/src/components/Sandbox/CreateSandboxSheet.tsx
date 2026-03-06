/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Badge } from '@/components/ui/badge'
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { FeatureFlags } from '@/enums/FeatureFlags'
import { RoutePath } from '@/enums/RoutePath'
import { useCreateSandboxMutation } from '@/hooks/mutations/useCreateSandboxMutation'
import { useSnapshotsQuery } from '@/hooks/queries/useSnapshotsQuery'
import { useConfig } from '@/hooks/useConfig'
import { useRegions } from '@/hooks/useRegions'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { parseEnvFile } from '@/lib/env'
import { handleApiError } from '@/lib/error-handling'
import { getRegionFullDisplayName } from '@/lib/utils'
import { Sandbox } from '@daytonaio/sdk'
import { useForm } from '@tanstack/react-form'
import { Minus, Plus, Upload } from 'lucide-react'
import { useFeatureFlagEnabled } from 'posthog-js/react'
import { useCallback, useEffect, useRef, useState } from 'react'
import { NumericFormat } from 'react-number-format'
import { generatePath, useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { z } from 'zod'
import { ScrollArea } from '../ui/scroll-area'

const NAME_REGEX = /^[a-zA-Z0-9][a-zA-Z0-9._-]*$/

const NONE_VALUE = '__none__'

const SOURCE_SNAPSHOT = 'snapshot'
const SOURCE_IMAGE = 'image'

const keyValuePairSchema = z.object({
  key: z.string(),
  value: z.string(),
})

const noDuplicateKeys = (pairs: { key: string; value: string }[] | undefined) => {
  if (!pairs) return true
  const keys = pairs.filter((p) => p.key).map((p) => p.key)
  return new Set(keys).size === keys.length
}

const formSchema = z.object({
  name: z
    .string()
    .optional()
    .refine((val) => !val || NAME_REGEX.test(val), 'Only letters, digits, dots, underscores and dashes are allowed'),
  snapshot: z.string().optional(),
  image: z.string().optional(),
  regionId: z.string().optional(),
  cpu: z.number().min(1).optional(),
  memory: z.number().min(1).optional(),
  disk: z.number().min(1).optional(),
  autoStopInterval: z.number().min(0).optional(),
  autoArchiveInterval: z.number().min(0).optional(),
  autoDeleteInterval: z
    .number()
    .refine((val) => val === -1 || val >= 0, 'Must be -1 (disabled) or a non-negative number')
    .optional(),
  envVars: z.array(keyValuePairSchema).optional().refine(noDuplicateKeys, 'Duplicate keys are not allowed'),
  labels: z.array(keyValuePairSchema).optional().refine(noDuplicateKeys, 'Duplicate keys are not allowed'),
  public: z.boolean().optional(),
  networkBlockAll: z.boolean().optional(),
  ephemeral: z.boolean().optional(),
})

type FormValues = z.infer<typeof formSchema>

const defaultValues: FormValues = {
  name: '',
  snapshot: undefined,
  image: '',
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
  ephemeral: false,
}

export const CreateSandboxSheet = ({ className }: { className?: string }) => {
  const navigate = useNavigate()
  const createSandboxEnabled = useFeatureFlagEnabled(FeatureFlags.DASHBOARD_CREATE_SANDBOX)
  const [open, setOpen] = useState(false)
  const [source, setSource] = useState(SOURCE_SNAPSHOT)

  const config = useConfig()
  const { availableRegions: regions, loadingAvailableRegions: loadingRegions } = useRegions()
  const { selectedOrganization } = useSelectedOrganization()
  const { reset: resetCreateSandboxMutation, ...createSandboxMutation } = useCreateSandboxMutation()
  const formRef = useRef<HTMLFormElement>(null)

  const maxCpu = selectedOrganization?.maxCpuPerSandbox
  const maxMemory = selectedOrganization?.maxMemoryPerSandbox
  const maxDisk = selectedOrganization?.maxDiskPerSandbox

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

      const envVars: Record<string, string> = {}
      value.envVars?.forEach(({ key, value: val }) => {
        if (key) envVars[key] = val
      })

      const labels: Record<string, string> = {}
      value.labels?.forEach(({ key, value: val }) => {
        if (key) labels[key] = val
      })

      const isImage = source === SOURCE_IMAGE

      const baseParams = {
        name: value.name?.trim() || undefined,
        target: value.regionId || undefined,
        autoStopInterval: value.autoStopInterval,
        autoArchiveInterval: value.autoArchiveInterval,
        autoDeleteInterval: value.autoDeleteInterval,
        ephemeral: value.ephemeral || undefined,
        envVars: Object.keys(envVars).length > 0 ? envVars : undefined,
        labels: Object.keys(labels).length > 0 ? labels : undefined,
        public: value.public || undefined,
        networkBlockAll: value.networkBlockAll || undefined,
      }

      let sandbox: Sandbox | undefined = undefined
      try {
        if (isImage && value.image) {
          sandbox = await createSandboxMutation.mutateAsync({
            ...baseParams,
            image: value.image,
            resources:
              value.cpu || value.memory || value.disk
                ? { cpu: value.cpu, memory: value.memory, disk: value.disk }
                : undefined,
          })
        } else {
          sandbox = await createSandboxMutation.mutateAsync({
            ...baseParams,
            snapshot: value.snapshot || undefined,
          })
        }

        toast.success(`Sandbox created`)

        setOpen(false)

        if (sandbox?.id) {
          navigate(generatePath(RoutePath.SANDBOX_DETAILS, { sandboxId: sandbox.id }))
        }
      } catch (error) {
        handleApiError(error, 'Failed to create sandbox')
      }
    },
  })

  const handleSourceChange = useCallback(
    (val: string) => {
      setSource(val)
      if (val === SOURCE_SNAPSHOT) {
        form.setFieldValue('image', '')
        form.setFieldValue('cpu', undefined)
        form.setFieldValue('memory', undefined)
        form.setFieldValue('disk', undefined)
      } else {
        form.setFieldValue('snapshot', undefined)
      }
    },
    [form],
  )

  const resetState = useCallback(() => {
    form.reset(defaultValues)
    setSource(SOURCE_SNAPSHOT)
    resetCreateSandboxMutation()
  }, [resetCreateSandboxMutation, form])

  const handleEnvFileImport = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0]
      if (!file) return

      const reader = new FileReader()
      reader.onload = () => {
        const parsed = parseEnvFile(reader.result as string)
        if (parsed.length === 0) return

        const existing = form.getFieldValue('envVars') ?? []
        const nonEmpty = existing.filter((p) => p.key || p.value)
        form.setFieldValue('envVars', [...nonEmpty, ...parsed])
      }
      reader.readAsText(file)
      e.target.value = ''
    },
    [form],
  )

  const handleEnvPaste = useCallback(
    (e: React.ClipboardEvent<HTMLInputElement>, index: number) => {
      const text = e.clipboardData.getData('text')
      if (!text.includes('=') || !text.includes('\n')) return
      e.preventDefault()
      const parsed = parseEnvFile(text)
      if (parsed.length === 0) return

      const existing = form.getFieldValue('envVars') ?? []
      const current = existing[index]
      const isEmptyRow = !current?.key && !current?.value
      const before = existing.slice(0, index)
      const after = existing.slice(index + (isEmptyRow ? 1 : 0))

      form.setFieldValue('envVars', [...before, ...parsed, ...after])
    },
    [form],
  )

  useEffect(() => {
    if (open) {
      resetState()
    }
  }, [open, resetState])

  if (!createSandboxEnabled) {
    return null
  }

  return (
    <Sheet
      open={open}
      onOpenChange={(isOpen) => {
        setOpen(isOpen)
      }}
    >
      <SheetTrigger asChild>
        <Button variant="default" size="sm" className="ml-auto" title="Create Sandbox">
          <Plus className="size-4" />
          Create Sandbox
        </Button>
      </SheetTrigger>
      <SheetContent className={`w-dvw sm:w-[500px] p-0 flex flex-col gap-0 ${className ?? ''}`}>
        <SheetHeader className="border-b border-border p-4 px-5 items-center flex text-left flex-row">
          <SheetTitle className="text-2xl">Create Sandbox</SheetTitle>
          <SheetDescription className="sr-only">Create a new sandbox in your organization.</SheetDescription>
        </SheetHeader>
        <ScrollArea fade="mask" className="flex-1 min-h-0">
          <form
            ref={formRef}
            id="create-sandbox-form"
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
                      Optional. If not provided, the sandbox ID will be used as the name. Names are reusable once a
                      sandbox is destroyed.
                    </FieldDescription>
                    {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                      <FieldError errors={field.state.meta.errors} />
                    )}
                  </Field>
                )
              }}
            </form.Field>

            <Tabs value={source} onValueChange={handleSourceChange} className="gap-3">
              <div className="flex flex-col gap-2">
                <FieldLabel>Source</FieldLabel>
                <TabsList className="w-full">
                  <TabsTrigger value={SOURCE_SNAPSHOT} className="flex-1">
                    Snapshot
                  </TabsTrigger>
                  <TabsTrigger value={SOURCE_IMAGE} className="flex-1">
                    Image
                  </TabsTrigger>
                </TabsList>
              </div>

              <TabsContent value={SOURCE_SNAPSHOT}>
                <form.Field name="snapshot">
                  {(field) => (
                    <Field>
                      <FieldLabel htmlFor={field.name}>Snapshot</FieldLabel>
                      <Select
                        value={field.state.value || NONE_VALUE}
                        onValueChange={(val) => field.handleChange(val === NONE_VALUE ? '' : val)}
                      >
                        <SelectTrigger
                          className="h-8"
                          id={field.name}
                          disabled={snapshotsLoading}
                          loading={snapshotsLoading}
                        >
                          <SelectValue placeholder={snapshotsLoading ? 'Loading snapshots...' : 'Select a snapshot'} />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value={NONE_VALUE}>
                            {config.defaultSnapshot} <Badge variant="secondary">default</Badge>
                          </SelectItem>
                          {snapshotsData?.items?.map((snapshot) => (
                            <SelectItem key={snapshot.id} value={snapshot.name}>
                              {snapshot.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </Field>
                  )}
                </form.Field>
              </TabsContent>

              <TabsContent value={SOURCE_IMAGE} className="flex flex-col gap-4">
                <form.Field name="image">
                  {(field) => {
                    const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                    return (
                      <Field data-invalid={isInvalid}>
                        <FieldLabel htmlFor={field.name}>Image</FieldLabel>
                        <Input
                          aria-invalid={isInvalid}
                          id={field.name}
                          value={field.state.value}
                          onBlur={field.handleBlur}
                          onChange={(e) => field.handleChange(e.target.value)}
                          placeholder="ubuntu:22.04"
                        />
                        <FieldDescription>
                          Must include either a tag (e.g., ubuntu:22.04) or a digest. The tag &quot;latest&quot; is not
                          allowed.
                        </FieldDescription>
                        {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                          <FieldError errors={field.state.meta.errors} />
                        )}
                      </Field>
                    )
                  }}
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
                          <NumericFormat
                            customInput={Input}
                            id={field.name}
                            className="w-full"
                            placeholder="1"
                            decimalScale={0}
                            allowNegative={false}
                            isAllowed={(values) => {
                              if (values.floatValue === undefined) return true
                              return !maxCpu || values.floatValue <= maxCpu
                            }}
                            value={field.state.value ?? ''}
                            onValueChange={(values) => field.handleChange(values.floatValue ?? undefined)}
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
                          <NumericFormat
                            customInput={Input}
                            id={field.name}
                            className="w-full"
                            placeholder="1"
                            decimalScale={0}
                            allowNegative={false}
                            isAllowed={(values) => {
                              if (values.floatValue === undefined) return true
                              return !maxMemory || values.floatValue <= maxMemory
                            }}
                            value={field.state.value ?? ''}
                            onValueChange={(values) => field.handleChange(values.floatValue ?? undefined)}
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
                          <NumericFormat
                            customInput={Input}
                            id={field.name}
                            className="w-full"
                            placeholder="3"
                            decimalScale={0}
                            allowNegative={false}
                            isAllowed={(values) => {
                              if (values.floatValue === undefined) return true
                              return !maxDisk || values.floatValue <= maxDisk
                            }}
                            value={field.state.value ?? ''}
                            onValueChange={(values) => field.handleChange(values.floatValue ?? undefined)}
                          />
                        </div>
                      )}
                    </form.Field>
                  </div>
                  <FieldDescription>
                    {`Defaults: 1 vCPU, 1 GiB memory, 3 GiB storage.`}
                    <br />
                    {maxCpu ? ` Limits: ${maxCpu} vCPU, ${maxMemory} GiB memory, ${maxDisk} GiB storage.` : ''}
                  </FieldDescription>
                </div>
              </TabsContent>
            </Tabs>

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
              <Label className="text-sm font-medium">Lifecycle</Label>
              <div className="flex flex-col gap-2">
                <form.Field name="autoStopInterval">
                  {(field) => (
                    <div className="flex items-center gap-4">
                      <Label htmlFor={field.name} className="w-40 flex-shrink-0">
                        Auto-stop (min):
                      </Label>
                      <NumericFormat
                        customInput={Input}
                        id={field.name}
                        className="w-full"
                        placeholder="15"
                        decimalScale={0}
                        allowNegative={false}
                        value={field.state.value ?? ''}
                        onValueChange={(values) => field.handleChange(values.floatValue ?? undefined)}
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
                      <NumericFormat
                        customInput={Input}
                        id={field.name}
                        className="w-full"
                        placeholder="10080 (7 days)"
                        decimalScale={0}
                        allowNegative={false}
                        value={field.state.value ?? ''}
                        onValueChange={(values) => field.handleChange(values.floatValue ?? undefined)}
                      />
                    </div>
                  )}
                </form.Field>
                <form.Field name="autoDeleteInterval">
                  {(field) => (
                    <form.Subscribe selector={(state) => state.values.ephemeral}>
                      {(ephemeral) => (
                        <div className="flex items-center gap-4">
                          <Label htmlFor={field.name} className="w-40 flex-shrink-0">
                            Auto-delete (min):
                          </Label>
                          <NumericFormat
                            customInput={Input}
                            id={field.name}
                            className="w-full"
                            placeholder="Disabled"
                            disabled={ephemeral}
                            decimalScale={0}
                            allowNegative
                            isAllowed={(values) => {
                              if (values.floatValue === undefined) return true
                              return values.floatValue === -1 || values.floatValue >= 0
                            }}
                            value={ephemeral ? 0 : (field.state.value ?? '')}
                            onValueChange={(values) => field.handleChange(values.floatValue ?? undefined)}
                          />
                        </div>
                      )}
                    </form.Subscribe>
                  )}
                </form.Field>
                <form.Field name="ephemeral">
                  {(field) => (
                    <div className="flex items-start gap-2 mt-1">
                      <Checkbox
                        id={field.name}
                        checked={field.state.value ?? false}
                        onCheckedChange={(checked) => {
                          const isEphemeral = checked === true
                          field.handleChange(isEphemeral)
                          if (isEphemeral) {
                            form.setFieldValue('autoDeleteInterval', 0)
                          }
                        }}
                      />
                      <Label htmlFor={field.name} className="text-sm font-normal">
                        Ephemeral — automatically delete the sandbox when it stops
                      </Label>
                    </div>
                  )}
                </form.Field>
              </div>
              <ul className="list-disc list-inside space-y-0.5 text-muted-foreground text-xs">
                <li>
                  Auto-stop: minutes of inactivity before stopping. Resets on preview access, SSH, or Toolbox API calls.
                  0 = disabled.
                </li>
                <li>
                  Auto-archive: minutes a sandbox must remain continuously stopped before archiving. 0 = max (30 days).
                </li>
                <li>
                  Auto-delete: minutes a sandbox must remain continuously stopped before permanent deletion. 0 = deleted
                  on stop. -1 = disabled.
                </li>
              </ul>
            </div>

            <form.Field name="envVars">
              {(field) => {
                const hasErrors = field.state.meta.errors.length > 0
                return (
                  <Field data-invalid={hasErrors}>
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
                            onPaste={(e) => handleEnvPaste(e, index)}
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
                            aria-label="Remove variable"
                            className="flex-shrink-0 h-8 w-8"
                            onClick={() => {
                              const updated = (field.state.value ?? []).filter((_, i) => i !== index)
                              field.handleChange(updated)
                            }}
                          >
                            <Minus className="size-4" />
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
                        <Plus className="size-4" />
                        Add Variable
                      </Button>
                    </div>
                    <FieldDescription>
                      <input
                        type="file"
                        accept="env"
                        className="hidden"
                        onChange={handleEnvFileImport}
                        aria-hidden="true"
                        id="env-file-input"
                      />
                      <label
                        className="inline-flex items-center gap-1 underline hover:text-foreground cursor-pointer"
                        htmlFor="env-file-input"
                      >
                        <Upload className="size-3" />
                        Import .env file
                      </label>{' '}
                      or paste .env contents into any key field.
                    </FieldDescription>
                    {hasErrors && <FieldError errors={field.state.meta.errors} />}
                  </Field>
                )
              }}
            </form.Field>

            <form.Field name="labels">
              {(field) => {
                const hasErrors = field.state.meta.errors.length > 0
                return (
                  <Field data-invalid={hasErrors}>
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
                            aria-label="Remove label"
                            className="flex-shrink-0 h-8 w-8"
                            onClick={() => {
                              const updated = (field.state.value ?? []).filter((_, i) => i !== index)
                              field.handleChange(updated)
                            }}
                          >
                            <Minus className="size-4" />
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
                        <Plus className="size-4" />
                        Add Label
                      </Button>
                    </div>
                    {hasErrors && <FieldError errors={field.state.meta.errors} />}
                  </Field>
                )
              }}
            </form.Field>

            <div className="flex flex-col gap-4">
              <Label className="text-sm font-medium">Network</Label>
              <form.Field name="public">
                {(field) => (
                  <div className="flex items-start gap-2">
                    <Checkbox
                      id={field.name}
                      className="mt-0.5"
                      checked={field.state.value ?? false}
                      onCheckedChange={(checked) => field.handleChange(checked === true)}
                    />
                    <div className="flex flex-col gap-1">
                      <Label htmlFor={field.name} className="text-sm font-normal">
                        Public HTTP Preview
                      </Label>
                      <FieldDescription>Allow public access to HTTP preview URLs.</FieldDescription>
                    </div>
                  </div>
                )}
              </form.Field>
              <form.Field name="networkBlockAll">
                {(field) => (
                  <div className="flex items-start gap-2">
                    <Checkbox
                      id={field.name}
                      className="mt-0.5"
                      checked={field.state.value ?? false}
                      onCheckedChange={(checked) => field.handleChange(checked === true)}
                    />
                    <div className="flex flex-col gap-1">
                      <Label htmlFor={field.name} className="text-sm font-normal">
                        Block All Network Access
                      </Label>
                      <FieldDescription>Block all outbound network access from the sandbox.</FieldDescription>
                    </div>
                  </div>
                )}
              </form.Field>
            </div>
          </form>
        </ScrollArea>
        <SheetFooter className="p-5 pt-3 border-t border-border">
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
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
