/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Command,
  CommandCheckboxItem,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandList,
} from '@/components/ui/command'
import { Field, FieldError, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
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
import { WEBHOOK_EVENT_CATEGORIES, WEBHOOK_EVENTS } from '@/constants/webhook-events'
import { useRefreshWebhookEndpointFlagMutation } from '@/hooks/mutations/useRefreshWebhookEndpointFlagMutation'
import { useUpdateWebhookEndpointMutation } from '@/hooks/mutations/useUpdateWebhookEndpointMutation'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { cn } from '@/lib/utils'
import { useForm } from '@tanstack/react-form'
import { useMutation } from '@tanstack/react-query'
import { ChevronsUpDown, Plus } from 'lucide-react'
import { Ref, type ReactNode, useCallback, useEffect, useImperativeHandle, useRef, useState } from 'react'
import { toast } from 'sonner'
import { EndpointOut } from 'svix'
import { useSvix } from 'svix-react'
import { z } from 'zod'

const formSchema = z.object({
  url: z.string().min(1, 'URL is required').url('Must be a valid URL'),
  description: z.string().trim().min(1, 'Name is required'),
  filterTypes: z.array(z.string()).min(1, 'At least one event is required'),
})

type FormValues = z.infer<typeof formSchema>

const defaultValues: FormValues = {
  url: '',
  description: '',
  filterTypes: [],
}

type UpsertEndpointSheetMode = 'create' | 'edit'

interface UpsertEndpointSheetProps {
  className?: string
  disabled?: boolean
  trigger?: ReactNode | null
  ref?: Ref<{ open: () => void }>
  mode?: UpsertEndpointSheetMode
  open?: boolean
  onOpenChange?: (open: boolean) => void
  endpoint?: EndpointOut | null
  onSuccess?: () => void
}

export const UpsertEndpointSheet = ({
  className,
  disabled,
  trigger,
  ref,
  mode = 'create',
  open,
  onOpenChange,
  endpoint,
  onSuccess,
}: UpsertEndpointSheetProps) => {
  const [internalOpen, setInternalOpen] = useState(false)
  const [eventsPopoverOpen, setEventsPopoverOpen] = useState(false)

  const isEditMode = mode === 'edit'
  const isControlled = open !== undefined
  const isOpen = open ?? internalOpen

  const formRef = useRef<HTMLFormElement>(null)
  const { svix, appId } = useSvix()
  const { selectedOrganization } = useSelectedOrganization()
  const refreshEndpointFlag = useRefreshWebhookEndpointFlagMutation()
  const { reset: resetUpdateEndpointMutation, ...updateEndpointMutation } = useUpdateWebhookEndpointMutation()
  const { reset: resetCreateEndpointMutation, ...createEndpointMutation } = useMutation({
    mutationFn: async (value: FormValues) => {
      await svix.endpoint.create(appId, {
        url: value.url.trim(),
        description: value.description?.trim() || undefined,
        filterTypes: value.filterTypes.length > 0 ? value.filterTypes : undefined,
      })
      if (selectedOrganization?.id) {
        refreshEndpointFlag.mutate(selectedOrganization.id)
      }
    },
  })

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
    if (!isEditMode || !endpoint) {
      return defaultValues
    }

    return {
      url: endpoint.url,
      description: endpoint.description || '',
      filterTypes: endpoint.filterTypes || [],
    }
  }, [endpoint, isEditMode])

  const form = useForm({
    defaultValues: getDefaultValues(),
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
      try {
        if (isEditMode) {
          if (!endpoint) {
            toast.error('No endpoint selected for editing.')
            return
          }

          await updateEndpointMutation.mutateAsync({
            endpointId: endpoint.id,
            update: {
              url: value.url.trim(),
              description: value.description?.trim() || undefined,
              filterTypes: value.filterTypes.length > 0 ? value.filterTypes : undefined,
            },
          })
          toast.success('Endpoint updated')
        } else {
          await createEndpointMutation.mutateAsync(value)
          toast.success('Endpoint created')
        }

        onSuccess?.()
        handleOpenChange(false)
      } catch (error) {
        handleApiError(error, `Failed to ${isEditMode ? 'update' : 'create'} endpoint`)
      }
    },
  })
  const { reset: resetForm } = form

  const resetState = useCallback(() => {
    resetForm(getDefaultValues())
    setEventsPopoverOpen(false)
    resetCreateEndpointMutation()
    resetUpdateEndpointMutation()
  }, [getDefaultValues, resetCreateEndpointMutation, resetForm, resetUpdateEndpointMutation])

  useEffect(() => {
    if (isOpen) {
      resetState()
    }
  }, [isOpen, resetState])

  const toggleEvent = (eventValue: string) => {
    const currentEvents = form.getFieldValue('filterTypes')
    if (currentEvents.includes(eventValue)) {
      form.setFieldValue(
        'filterTypes',
        currentEvents.filter((event) => event !== eventValue),
      )
    } else {
      form.setFieldValue('filterTypes', [...currentEvents, eventValue])
    }
  }

  const formId = isEditMode ? 'edit-endpoint-form' : 'create-endpoint-form'

  return (
    <Sheet open={isOpen} onOpenChange={handleOpenChange}>
      {trigger === undefined ? (
        <SheetTrigger asChild>
          <Button
            variant="default"
            size="sm"
            disabled={disabled}
            className={className}
            title={isEditMode ? 'Edit Endpoint' : 'Add Endpoint'}
          >
            {!isEditMode && <Plus className="w-4 h-4" />}
            {isEditMode ? 'Edit Endpoint' : 'Add Endpoint'}
          </Button>
        </SheetTrigger>
      ) : (
        trigger
      )}
      <SheetContent className="w-dvw sm:w-[500px] p-0 flex flex-col gap-0">
        <SheetHeader className="border-b border-border p-4 px-5 items-center flex text-left flex-row">
          <SheetTitle className="text-2xl">{isEditMode ? 'Edit Webhook Endpoint' : 'Add Webhook Endpoint'}</SheetTitle>
          <SheetDescription className="sr-only">
            {isEditMode ? 'Update the endpoint configuration.' : 'Configure a new endpoint to receive webhook events.'}
          </SheetDescription>
        </SheetHeader>

        <ScrollArea fade="mask" className="flex-1 min-h-0">
          <form
            ref={formRef}
            id={formId}
            className="space-y-6 p-5"
            onSubmit={(e) => {
              e.preventDefault()
              e.stopPropagation()
              form.handleSubmit()
            }}
          >
            <form.Field name="description">
              {(field) => {
                const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                return (
                  <Field data-invalid={isInvalid}>
                    <FieldLabel htmlFor={field.name}>Endpoint Name</FieldLabel>
                    <Input
                      aria-invalid={isInvalid}
                      autoComplete="off"
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      placeholder="My Webhook Endpoint"
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
                    <FieldLabel htmlFor={field.name}>Endpoint URL</FieldLabel>
                    <Input
                      aria-invalid={isInvalid}
                      autoComplete="off"
                      id={field.name}
                      name={field.name}
                      type="url"
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      placeholder="https://example.com/webhook"
                    />
                    {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                      <FieldError errors={field.state.meta.errors} />
                    )}
                  </Field>
                )
              }}
            </form.Field>

            <form.Field name="filterTypes">
              {(field) => {
                const selectedEvents = field.state.value
                const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                return (
                  <Field data-invalid={isInvalid}>
                    <FieldLabel>Events</FieldLabel>
                    <Popover open={eventsPopoverOpen} onOpenChange={setEventsPopoverOpen} modal>
                      <PopoverTrigger asChild>
                        <Button
                          variant="outline"
                          role="combobox"
                          aria-invalid={isInvalid}
                          aria-expanded={eventsPopoverOpen}
                          className={cn('w-full justify-between h-auto min-h-10', {
                            '!pl-2': selectedEvents.length > 0,
                          })}
                        >
                          <div className="flex flex-wrap gap-1">
                            {selectedEvents.length === 0 ? (
                              <span className="text-muted-foreground">Select events...</span>
                            ) : selectedEvents.length > 2 ? (
                              <Badge variant="secondary" className="rounded-sm px-1 font-normal">
                                {selectedEvents.length} events selected
                              </Badge>
                            ) : (
                              selectedEvents.map((event) => (
                                <Badge key={event} variant="secondary" className="rounded-sm px-1 font-normal">
                                  {WEBHOOK_EVENTS.find((webhookEvent) => webhookEvent.value === event)?.label || event}
                                </Badge>
                              ))
                            )}
                          </div>
                          <ChevronsUpDown className="ml-2 size-4 shrink-0 opacity-50" />
                        </Button>
                      </PopoverTrigger>
                      <PopoverContent className="w-[var(--radix-popover-trigger-width)] p-0" align="start">
                        <Command>
                          <CommandInput placeholder="Search events..." />
                          <CommandList>
                            <CommandEmpty>No events found.</CommandEmpty>
                            {WEBHOOK_EVENT_CATEGORIES.map((category) => (
                              <CommandGroup key={category} heading={category}>
                                {WEBHOOK_EVENTS.filter((event) => event.category === category).map((event) => (
                                  <CommandCheckboxItem
                                    key={event.value}
                                    checked={selectedEvents.includes(event.value)}
                                    onSelect={() => toggleEvent(event.value)}
                                  >
                                    {event.label}
                                  </CommandCheckboxItem>
                                ))}
                              </CommandGroup>
                            ))}
                          </CommandList>
                        </Command>
                      </PopoverContent>
                    </Popover>
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
                form={formId}
                variant="default"
                disabled={
                  !canSubmit ||
                  isSubmitting ||
                  (isEditMode ? updateEndpointMutation.isPending : createEndpointMutation.isPending)
                }
              >
                {isSubmitting && <Spinner />}
                {isEditMode ? 'Save' : 'Create'}
              </Button>
            )}
          />
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
