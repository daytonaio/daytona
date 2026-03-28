/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useCallback, useEffect, useRef, useState } from 'react'
import { Check, Copy } from 'lucide-react'
import { Organization, Region } from '@daytonaio/api-client'
import { useForm } from '@tanstack/react-form'
import { useMutation } from '@tanstack/react-query'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Link } from 'react-router-dom'
import { Field, FieldDescription, FieldError, FieldLabel } from '@/components/ui/field'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useCopyToClipboard } from '@/hooks/useCopyToClipboard'
import { RoutePath } from '@/enums/RoutePath'

interface CreateOrganizationSheetProps {
  open: boolean
  billingApiUrl?: string
  regions: Region[]
  loadingRegions: boolean
  getRegionName: (regionId: string) => string | undefined
  onOpenChange: (open: boolean) => void
  onCreateOrganization: (name: string, defaultRegionId: string) => Promise<Organization | null>
}

const formSchema = z.object({
  name: z.string().min(1, 'Organization name is required'),
  defaultRegionId: z.string().min(1, 'Region is required'),
})

type FormValues = z.infer<typeof formSchema>

const defaultValues: FormValues = {
  name: '',
  defaultRegionId: '',
}

export const CreateOrganizationSheet: React.FC<CreateOrganizationSheetProps> = ({
  open,
  billingApiUrl,
  regions,
  loadingRegions,
  getRegionName,
  onOpenChange,
  onCreateOrganization,
}) => {
  const [createdOrg, setCreatedOrg] = useState<Organization | null>(null)
  const [copiedText, copyToClipboard] = useCopyToClipboard()
  const formRef = useRef<HTMLFormElement>(null)

  const createOrganizationMutation = useMutation({
    mutationFn: async (value: FormValues) => {
      return onCreateOrganization(value.name.trim(), value.defaultRegionId)
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
      const organization = await createOrganizationMutation.mutateAsync(value)
      if (!organization) {
        return
      }

      setCreatedOrg(organization)
      resetForm(defaultValues)
    },
  })
  const { reset: resetForm } = form

  useEffect(() => {
    if (!open) {
      return
    }

    if (!form.getFieldValue('defaultRegionId') && regions[0]?.id) {
      form.setFieldValue('defaultRegionId', regions[0].id)
    }
  }, [form, open, regions])

  const { reset: resetMutation } = createOrganizationMutation

  const resetState = useCallback(() => {
    setCreatedOrg(null)
    resetForm({
      ...defaultValues,
      defaultRegionId: regions[0]?.id ?? '',
    })
    resetMutation()
  }, [resetForm, resetMutation, regions])

  useEffect(() => {
    if (open) {
      resetState()
    }
  }, [open, resetState])

  return (
    <Sheet
      open={open}
      onOpenChange={(isOpen) => {
        onOpenChange(isOpen)
        if (!isOpen) {
          resetState()
        }
      }}
    >
      <SheetContent className="w-dvw sm:w-[560px] p-0 flex flex-col gap-0">
        <SheetHeader className="border-b border-border p-4 px-5 items-center flex text-left flex-row">
          <SheetTitle className="text-2xl">{createdOrg ? 'New Organization' : 'Create New Organization'}</SheetTitle>
          <SheetDescription className="sr-only">
            {createdOrg
              ? 'You can switch between organizations in the top left corner of the sidebar.'
              : 'Create a new organization to share resources and collaborate with others.'}
          </SheetDescription>
        </SheetHeader>

        <ScrollArea fade="mask" className="flex-1 min-h-0">
          <div className="p-5">
            {createdOrg ? (
              <div className="space-y-6">
                <div className="space-y-3">
                  <FieldLabel htmlFor="organization-id">Organization ID</FieldLabel>
                  <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                    <span className="overflow-x-auto pr-2 cursor-text select-all">{createdOrg.id}</span>
                    {(copiedText === createdOrg.id && <Check className="w-4 h-4" />) || (
                      <Copy className="w-4 h-4 cursor-pointer" onClick={() => copyToClipboard(createdOrg.id)} />
                    )}
                  </div>
                </div>

                {createdOrg.defaultRegionId && (
                  <div className="space-y-3">
                    <FieldLabel htmlFor="organization-default-region">Default Region</FieldLabel>
                    <Input
                      id="organization-default-region"
                      value={getRegionName(createdOrg.defaultRegionId) ?? createdOrg.defaultRegionId}
                      readOnly
                    />
                  </div>
                )}

                <div className="p-3 rounded-md bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400">
                  <p className="font-medium">Your organization is created.</p>
                  <p className="text-sm mt-1">
                    {billingApiUrl ? (
                      <>
                        To get started, add a payment method on the{' '}
                        <Link
                          to={RoutePath.BILLING_WALLET}
                          className="text-blue-500 hover:underline"
                          onClick={() => {
                            onOpenChange(false)
                          }}
                        >
                          wallet page
                        </Link>
                        .
                      </>
                    ) : null}
                  </p>
                </div>
              </div>
            ) : !loadingRegions && regions.length === 0 ? (
              <div className="p-3 rounded-md bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400">
                <p className="font-medium">No regions available</p>
                <p className="text-sm mt-1">Organization cannot be created because no regions are available.</p>
              </div>
            ) : (
              <form
                ref={formRef}
                id="create-organization-form"
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
                        <FieldLabel htmlFor={field.name}>Organization Name</FieldLabel>
                        <Input
                          aria-invalid={isInvalid}
                          id={field.name}
                          name={field.name}
                          value={field.state.value}
                          onBlur={field.handleBlur}
                          onChange={(e) => field.handleChange(e.target.value)}
                          placeholder="Name"
                        />
                        {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                          <FieldError errors={field.state.meta.errors} />
                        )}
                      </Field>
                    )
                  }}
                </form.Field>

                <form.Field name="defaultRegionId">
                  {(field) => {
                    const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                    return (
                      <Field data-invalid={isInvalid}>
                        <FieldLabel htmlFor={field.name}>Region</FieldLabel>
                        <Select value={field.state.value} onValueChange={field.handleChange}>
                          <SelectTrigger
                            className="h-8"
                            id={field.name}
                            disabled={loadingRegions}
                            loading={loadingRegions}
                            aria-invalid={isInvalid}
                          >
                            <SelectValue placeholder={loadingRegions ? 'Loading regions...' : 'Select a region'} />
                          </SelectTrigger>
                          <SelectContent>
                            {regions.map((region) => (
                              <SelectItem key={region.id} value={region.id}>
                                {region.name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <FieldDescription>
                          The region that will be used as the default target for creating sandboxes in this
                          organization.
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
          <Button
            type="button"
            variant="secondary"
            disabled={createOrganizationMutation.isPending}
            onClick={() => onOpenChange(false)}
          >
            {createdOrg ? 'Close' : 'Cancel'}
          </Button>
          {!createdOrg && (
            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
                <Button
                  type="submit"
                  form="create-organization-form"
                  variant="default"
                  disabled={!canSubmit || isSubmitting || regions.length === 0}
                >
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
