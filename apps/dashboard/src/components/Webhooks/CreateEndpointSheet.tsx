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
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { Field, FieldError, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Spinner } from '@/components/ui/spinner'
import { WEBHOOK_EVENT_CATEGORIES, WEBHOOK_EVENTS } from '@/constants/webhook-events'
import { handleApiError } from '@/lib/error-handling'
import { cn } from '@/lib/utils'
import { useForm } from '@tanstack/react-form'
import { useMutation } from '@tanstack/react-query'
import { ChevronsUpDown, Plus } from 'lucide-react'
import React, { Ref, useCallback, useEffect, useImperativeHandle, useRef, useState } from 'react'
import { toast } from 'sonner'
import { useSvix } from 'svix-react'
import { z } from 'zod'

const formSchema = z.object({
  url: z.string().min(1, 'URL is required').url('Must be a valid URL'),
  description: z.string(),
  filterTypes: z.array(z.string()).min(1, 'At least one event is required'),
})

type FormValues = z.infer<typeof formSchema>

interface CreateEndpointSheetProps {
  onSuccess: () => void
  className?: string
  ref?: Ref<{ open: () => void }>
}

export const CreateEndpointSheet: React.FC<CreateEndpointSheetProps> = ({ onSuccess, className, ref }) => {
  const [open, setOpen] = useState(false)
  const [eventsPopoverOpen, setEventsPopoverOpen] = useState(false)
  const formRef = useRef<HTMLFormElement>(null)

  useImperativeHandle(ref, () => ({
    open: () => setOpen(true),
  }))

  const { svix, appId } = useSvix()
  const createEndpointMutation = useMutation({
    mutationFn: async (value: FormValues) => {
      await svix.endpoint.create(appId, {
        url: value.url.trim(),
        description: value.description?.trim() || undefined,
        filterTypes: value.filterTypes.length > 0 ? value.filterTypes : undefined,
      })
    },
  })

  const form = useForm({
    defaultValues: {
      url: '',
      description: '',
      filterTypes: [],
    } as FormValues,
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
      try {
        await createEndpointMutation.mutateAsync(value)
        toast.success('Endpoint created')
        onSuccess()
        setOpen(false)
      } catch (error) {
        handleApiError(error, 'Failed to create endpoint')
      }
    },
  })
  const { reset: resetForm } = form

  const resetState = useCallback(() => {
    resetForm()
    setEventsPopoverOpen(false)
  }, [resetForm])

  useEffect(() => {
    if (open) {
      resetState()
    }
  }, [open, resetState])

  const toggleEvent = (eventValue: string) => {
    const currentEvents = form.getFieldValue('filterTypes')
    if (currentEvents.includes(eventValue)) {
      form.setFieldValue(
        'filterTypes',
        currentEvents.filter((e) => e !== eventValue),
      )
    } else {
      form.setFieldValue('filterTypes', [...currentEvents, eventValue])
    }
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="default" size="sm" className={className}>
          <Plus className="w-4 h-4" />
          Add Endpoint
        </Button>
      </SheetTrigger>
      <SheetContent className="w-dvw sm:w-[500px] p-0 flex flex-col gap-0">
        <SheetHeader className="border-b border-border p-4 px-5 items-center flex text-left flex-row">
          <SheetTitle className="text-2xl">Add Webhook Endpoint</SheetTitle>
          <SheetDescription className="sr-only">Configure a new endpoint to receive webhook events.</SheetDescription>
        </SheetHeader>
        <ScrollArea fade="mask" className="flex-1 min-h-0">
          <form
            ref={formRef}
            id="create-endpoint-form"
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
                      autoComplete="off"
                      aria-invalid={isInvalid}
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
                                  {WEBHOOK_EVENTS.find((e) => e.value === event)?.label || event}
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
          <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <form.Subscribe
            selector={(state) => [state.canSubmit, state.isSubmitting]}
            children={([canSubmit, isSubmitting]) => (
              <Button type="submit" form="create-endpoint-form" variant="default" disabled={!canSubmit || isSubmitting}>
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
