/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import TooltipButton from '@/components/TooltipButton'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Field, FieldContent, FieldDescription, FieldError, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Spinner } from '@/components/ui/spinner'
import { useDeleteOrganizationOtelConfigMutation } from '@/hooks/mutations/useDeleteOrganizationOtelConfigMutation'
import { useUpdateOrganizationOtelConfigMutation } from '@/hooks/mutations/useUpdateOrganizationOtelConfigMutation'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import type { Organization } from '@daytona/api-client'
import { useForm } from '@tanstack/react-form'
import { Minus, Plus } from 'lucide-react'
import React, { useEffect, useRef } from 'react'
import { toast } from 'sonner'
import { z } from 'zod'

type HeaderEntry = { key: string; value: string }

const emptyHeader = (): HeaderEntry => ({ key: '', value: '' })

const headerRowIsComplete = (headers: HeaderEntry[]) =>
  headers.every(({ key, value }) => {
    const hasKey = !!key.trim()
    const hasValue = !!value.trim()
    return hasKey === hasValue
  })

const noDuplicateHeaderKeys = (headers: HeaderEntry[]) => {
  const keys = headers.map(({ key }) => key.trim().toLowerCase()).filter(Boolean)

  return new Set(keys).size === keys.length
}

const formSchema = z.object({
  endpoint: z
    .string()
    .trim()
    .refine((value) => {
      try {
        new URL(value)
        return true
      } catch {
        return false
      }
    }, 'A valid OTLP endpoint URL is required'),
  headers: z
    .array(
      z.object({
        key: z.string().trim(),
        value: z.string().trim(),
      }),
    )
    .refine(headerRowIsComplete, 'Header rows need both a key and a value')
    .refine(noDuplicateHeaderKeys, 'Header keys must be unique'),
})

type FormValues = z.infer<typeof formSchema>

const headersFromOrganization = (organization: Organization | null | undefined): HeaderEntry[] => {
  const headers = organization?.otelConfig?.headers
  if (!headers) {
    return []
  }
  return Object.entries(headers).map(([key, value]) => ({ key, value: value as string }))
}

const endpointFromOrganization = (organization: Organization | null | undefined): string =>
  organization?.otelConfig?.endpoint ?? ''

const valuesFromOrganization = (organization: Organization | null | undefined): FormValues => ({
  endpoint: endpointFromOrganization(organization),
  headers: headersFromOrganization(organization).length ? headersFromOrganization(organization) : [emptyHeader()],
})

const getHeaderRowErrors = (headers: HeaderEntry[]) => {
  const errors: Record<number, { key?: string; value?: string }> = {}
  const keys = new Map<string, number>()

  headers.forEach(({ key, value }, index) => {
    const trimmedKey = key.trim()
    const normalizedKey = trimmedKey.toLowerCase()
    const trimmedValue = value.trim()

    if (!trimmedKey && !trimmedValue) {
      return
    }

    if (!trimmedKey) {
      errors[index] = { ...errors[index], key: 'Header key is required' }
    }

    if (!trimmedValue) {
      errors[index] = { ...errors[index], value: 'Header value is required' }
    }

    if (!trimmedKey || !trimmedValue) {
      return
    }

    const duplicateIndex = keys.get(normalizedKey)
    if (duplicateIndex != null) {
      errors[index] = { ...errors[index], key: 'Header keys must be unique' }
      errors[duplicateIndex] = { ...errors[duplicateIndex], key: 'Header keys must be unique' }
      return
    }

    keys.set(normalizedKey, index)
  })

  return errors
}

export const OtelConfigCard: React.FC = () => {
  const { selectedOrganization } = useSelectedOrganization()
  const updateOtelConfigMutation = useUpdateOrganizationOtelConfigMutation()
  const deleteOtelConfigMutation = useDeleteOrganizationOtelConfigMutation()

  const formRef = useRef<HTMLFormElement>(null)
  const headerKeyInputRefs = useRef<Array<HTMLInputElement | null>>([])

  const hasOtelEnabled = !!selectedOrganization?.otelConfig
  const saving = updateOtelConfigMutation.isPending
  const disabling = deleteOtelConfigMutation.isPending

  const form = useForm({
    defaultValues: valuesFromOrganization(selectedOrganization),
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
      if (!selectedOrganization) {
        return
      }

      try {
        const headers = value.headers
          .filter(({ key, value }) => key.trim() || value.trim())
          .reduce(
            (acc, { key, value }) => {
              acc[key.trim()] = value.trim()
              return acc
            },
            {} as Record<string, string>,
          )

        await updateOtelConfigMutation.mutateAsync({
          organizationId: selectedOrganization.id,
          otelConfig: {
            endpoint: value.endpoint.trim(),
            headers,
          },
        })
        toast.success('OpenTelemetry configuration saved')
      } catch (error) {
        handleApiError(error, 'Failed to save OpenTelemetry configuration')
      }
    },
  })

  const handleDisable = async () => {
    if (!selectedOrganization) {
      return
    }

    try {
      await deleteOtelConfigMutation.mutateAsync({ organizationId: selectedOrganization.id })
      form.reset({ endpoint: '', headers: [emptyHeader()] })
      toast.success('OpenTelemetry configuration disabled')
    } catch (error) {
      handleApiError(error, 'Failed to disable OpenTelemetry configuration')
    }
  }

  const selectedOrganizationId = selectedOrganization?.id
  const selectedOtelEndpoint = selectedOrganization?.otelConfig?.endpoint
  const selectedOtelHeaders = selectedOrganization?.otelConfig?.headers

  useEffect(() => {
    form.reset(valuesFromOrganization(selectedOrganization))
  }, [selectedOrganizationId, selectedOtelEndpoint, selectedOtelHeaders])

  return (
    <Card>
      <CardHeader className="p-4">
        <CardTitle>OpenTelemetry</CardTitle>
      </CardHeader>
      <CardContent className="border-t border-border p-0">
        <form
          ref={formRef}
          id="otel-config-form"
          onSubmit={async (e) => {
            e.preventDefault()
            e.stopPropagation()
            form.handleSubmit()
          }}
        >
          <div className="border-b border-border p-4 last:border-b-0">
            <form.Field name="endpoint">
              {(field) => {
                const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid

                return (
                  <Field data-invalid={isInvalid} className="grid gap-3 sm:grid-cols-2 sm:items-center">
                    <FieldContent>
                      <FieldLabel htmlFor={field.name}>OTLP Endpoint</FieldLabel>
                      <FieldDescription>The OpenTelemetry collector endpoint URL.</FieldDescription>
                    </FieldContent>
                    <div className="space-y-1">
                      <Input
                        aria-invalid={isInvalid}
                        id={field.name}
                        name={field.name}
                        placeholder="https://otel-collector.example.com:4318"
                        value={field.state.value}
                        onBlur={field.handleBlur}
                        onChange={(e) => field.handleChange(e.target.value)}
                      />
                      {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                        <FieldError errors={field.state.meta.errors} />
                      )}
                    </div>
                  </Field>
                )
              }}
            </form.Field>
          </div>

          <div className="border-b border-border p-4 last:border-b-0">
            <form.Field name="headers">
              {(field) => {
                return (
                  <form.Subscribe
                    selector={(state) => state.submissionAttempts}
                    children={(submissionAttempts) => {
                      const headerErrors = getHeaderRowErrors(field.state.value)
                      const showErrors = submissionAttempts > 0 && Object.keys(headerErrors).length > 0

                      return (
                        <Field data-invalid={showErrors} className="gap-3">
                          <FieldContent>
                            <FieldLabel>Headers</FieldLabel>
                            <FieldDescription>
                              Optional headers to send with OTLP requests. Existing values are stored encrypted and
                              shown as <code>******</code>.
                            </FieldDescription>
                          </FieldContent>
                          <div className="space-y-2">
                            {field.state.value.map((header, index) => (
                              <HeaderInput
                                key={index}
                                keyInputRef={(element) => {
                                  headerKeyInputRefs.current[index] = element
                                }}
                                headerKey={header.key}
                                headerValue={header.value}
                                errors={showErrors ? headerErrors[index] : undefined}
                                onChangeKey={(key) => {
                                  const next = [...field.state.value]
                                  next[index] = { ...next[index], key }
                                  field.handleChange(next)
                                }}
                                onChangeValue={(value) => {
                                  const next = [...field.state.value]
                                  next[index] = { ...next[index], value }
                                  field.handleChange(next)
                                }}
                                onRemove={() => field.handleChange(field.state.value.filter((_, i) => i !== index))}
                              />
                            ))}
                            <div className="flex justify-start">
                              <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                onClick={() => {
                                  const nextIndex = field.state.value.length
                                  field.handleChange([...field.state.value, emptyHeader()])
                                  setTimeout(() => headerKeyInputRefs.current[nextIndex]?.focus())
                                }}
                              >
                                <Plus className="size-4" />
                                Add Header
                              </Button>
                            </div>
                          </div>
                        </Field>
                      )
                    }}
                  />
                )
              }}
            </form.Field>
          </div>

          <div className="flex justify-end gap-2 p-4">
            {hasOtelEnabled && (
              <form.Subscribe
                selector={(state) => state.isSubmitting}
                children={(isSubmitting) => (
                  <Button type="button" variant="outline" onClick={handleDisable} disabled={isSubmitting || disabling}>
                    {disabling && <Spinner />}
                    Disable
                  </Button>
                )}
              />
            )}
            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
                <Button
                  type="submit"
                  form="otel-config-form"
                  disabled={!canSubmit || isSubmitting || saving || disabling}
                >
                  {(isSubmitting || saving) && <Spinner />}
                  Save
                </Button>
              )}
            />
          </div>
        </form>
      </CardContent>
    </Card>
  )
}

const HeaderInput = ({
  keyInputRef,
  headerKey,
  headerValue,
  errors,
  onChangeKey,
  onChangeValue,
  onRemove,
}: {
  keyInputRef: (element: HTMLInputElement | null) => void
  headerKey: string
  headerValue: string
  errors?: { key?: string; value?: string }
  onChangeKey: (value: string) => void
  onChangeValue: (value: string) => void
  onRemove: () => void
}) => (
  <div className="space-y-1">
    <div className="grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)_2rem] gap-2">
      <Input
        ref={keyInputRef}
        aria-invalid={!!errors?.key}
        placeholder="Header key"
        value={headerKey}
        onChange={(e) => onChangeKey(e.target.value)}
      />
      <Input
        aria-invalid={!!errors?.value}
        placeholder="Header value"
        value={headerValue}
        onChange={(e) => onChangeValue(e.target.value)}
      />
      <TooltipButton
        type="button"
        tooltipText="Remove header"
        variant="ghost"
        size="icon"
        className="flex-shrink-0 h-8 w-8"
        onClick={onRemove}
      >
        <Minus className="size-4" />
      </TooltipButton>
    </div>
    {(errors?.key || errors?.value) && <FieldError>{[errors.key, errors.value].filter(Boolean).join('. ')}</FieldError>}
  </div>
)
