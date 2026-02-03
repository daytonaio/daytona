/*
 * Copyright 2025 Daytona Platforms Inc.
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
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Field, FieldError, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Spinner } from '@/components/ui/spinner'
import { WEBHOOK_EVENT_CATEGORIES, WEBHOOK_EVENTS } from '@/constants/webhook-events'
import { handleApiError } from '@/lib/error-handling'
import { useForm } from '@tanstack/react-form'
import { ChevronsUpDown, Plus } from 'lucide-react'
import React, { useCallback, useEffect, useState } from 'react'
import { toast } from 'sonner'
import { useSvix } from 'svix-react'
import { z } from 'zod'

const formSchema = z.object({
  url: z.string().min(1, 'URL is required').url('Must be a valid URL'),
  description: z.string().optional(),
  filterTypes: z.array(z.string()),
})

type FormValues = z.infer<typeof formSchema>

interface CreateEndpointDialogProps {
  onSuccess: () => void
  className?: string
}

export const CreateEndpointDialog: React.FC<CreateEndpointDialogProps> = ({ onSuccess, className }) => {
  const [open, setOpen] = useState(false)
  const [eventsPopoverOpen, setEventsPopoverOpen] = useState(false)

  const { svix, appId } = useSvix()

  const form = useForm({
    defaultValues: {
      url: '',
      description: '',
      filterTypes: [],
    } as FormValues,
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: async ({ value }) => {
      try {
        await svix.endpoint.create(appId, {
          url: value.url.trim(),
          description: value.description?.trim() || undefined,
          filterTypes: value.filterTypes.length > 0 ? value.filterTypes : undefined,
        })
        toast.success('Endpoint created')
        onSuccess()
        setOpen(false)
      } catch (error) {
        handleApiError(error, 'Failed to create endpoint')
      }
    },
  })

  const resetState = useCallback(() => {
    form.reset()
  }, [form])

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
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="default" size="sm" title="Add Endpoint" className={className}>
          <Plus className="w-4 h-4" />
          Add Endpoint
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Add Webhook Endpoint</DialogTitle>
          <DialogDescription>Configure a new endpoint to receive webhook events.</DialogDescription>
        </DialogHeader>
        <form
          id="create-endpoint-form"
          className="space-y-6 overflow-y-auto px-1 pb-1"
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
              return (
                <Field>
                  <FieldLabel>Events</FieldLabel>
                  <Popover open={eventsPopoverOpen} onOpenChange={setEventsPopoverOpen} modal>
                    <PopoverTrigger asChild>
                      <Button
                        variant="outline"
                        role="combobox"
                        aria-expanded={eventsPopoverOpen}
                        className="w-full justify-between h-auto min-h-10"
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
                        <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
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
                </Field>
              )
            }}
          </form.Field>
        </form>
        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="secondary">
              Cancel
            </Button>
          </DialogClose>
          <form.Subscribe
            selector={(state) => [state.canSubmit, state.isSubmitting]}
            children={([canSubmit, isSubmitting]) => (
              <Button type="submit" form="create-endpoint-form" variant="default" disabled={!canSubmit || isSubmitting}>
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
