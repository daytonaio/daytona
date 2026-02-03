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
  CommandItem,
  CommandList,
  CommandSeparator,
} from '@/components/ui/command'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { useUpdateWebhookEndpointMutation } from '@/hooks/mutations/useUpdateWebhookEndpointMutation'
import { handleApiError } from '@/lib/error-handling'
import { ChevronsUpDown } from 'lucide-react'
import React, { useEffect, useState } from 'react'
import { EndpointOut } from 'svix'

const AVAILABLE_EVENTS = [
  { value: 'sandbox.created', label: 'Sandbox Created', category: 'Sandbox' },
  { value: 'sandbox.state.updated', label: 'Sandbox State Updated', category: 'Sandbox' },
  { value: 'snapshot.created', label: 'Snapshot Created', category: 'Snapshot' },
  { value: 'snapshot.removed', label: 'Snapshot Removed', category: 'Snapshot' },
  { value: 'snapshot.state.updated', label: 'Snapshot State Updated', category: 'Snapshot' },
  { value: 'volume.created', label: 'Volume Created', category: 'Volume' },
  { value: 'volume.state.updated', label: 'Volume State Updated', category: 'Volume' },
]

interface EditEndpointDialogProps {
  endpoint: EndpointOut | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
}

export const EditEndpointDialog: React.FC<EditEndpointDialogProps> = ({ endpoint, open, onOpenChange, onSuccess }) => {
  const [eventsPopoverOpen, setEventsPopoverOpen] = useState(false)
  const [url, setUrl] = useState('')
  const [description, setDescription] = useState('')
  const [eventTypes, setEventTypes] = useState<string[]>([])

  const updateMutation = useUpdateWebhookEndpointMutation()

  useEffect(() => {
    if (endpoint && open) {
      setUrl(endpoint.url)
      setDescription(endpoint.description || '')
      setEventTypes(endpoint.filterTypes || [])
    }
  }, [endpoint, open])

  const handleUpdateEndpoint = async () => {
    if (!url.trim() || !endpoint) {
      return
    }

    try {
      await updateMutation.mutateAsync({
        endpointId: endpoint.id,
        update: {
          url: url.trim(),
          description: description.trim() || undefined,
          filterTypes: eventTypes.length > 0 ? eventTypes : undefined,
        },
      })
      onSuccess()
      onOpenChange(false)
    } catch (error) {
      handleApiError(error, 'Failed to update endpoint')
    }
  }

  const toggleEvent = (eventValue: string) => {
    if (eventTypes.includes(eventValue)) {
      setEventTypes(eventTypes.filter((e) => e !== eventValue))
    } else {
      setEventTypes([...eventTypes, eventValue])
    }
  }

  const handleOpenChange = (isOpen: boolean) => {
    onOpenChange(isOpen)
    if (!isOpen) {
      setUrl('')
      setDescription('')
      setEventTypes([])
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Edit Webhook Endpoint</DialogTitle>
          <DialogDescription>Update the endpoint configuration.</DialogDescription>
        </DialogHeader>
        <form
          id="edit-endpoint-form"
          className="space-y-6 overflow-y-auto px-1 pb-1"
          onSubmit={async (e) => {
            e.preventDefault()
            await handleUpdateEndpoint()
          }}
        >
          <div className="space-y-3">
            <Label htmlFor="endpoint-name">Endpoint Name</Label>
            <Input
              id="endpoint-name"
              type="text"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="My Webhook Endpoint"
            />
          </div>
          <div className="space-y-3">
            <Label htmlFor="endpoint-url">Endpoint URL</Label>
            <Input
              id="endpoint-url"
              type="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://example.com/webhook"
            />
          </div>
          <div className="space-y-3">
            <Label>Events</Label>
            <Popover open={eventsPopoverOpen} onOpenChange={setEventsPopoverOpen}>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  role="combobox"
                  aria-expanded={eventsPopoverOpen}
                  className="w-full justify-between h-auto min-h-10"
                >
                  <div className="flex flex-wrap gap-1">
                    {eventTypes.length === 0 ? (
                      <span className="text-muted-foreground">Select events...</span>
                    ) : eventTypes.length > 2 ? (
                      <Badge variant="secondary" className="rounded-sm px-1 font-normal">
                        {eventTypes.length} events selected
                      </Badge>
                    ) : (
                      eventTypes.map((event) => (
                        <Badge key={event} variant="secondary" className="rounded-sm px-1 font-normal">
                          {AVAILABLE_EVENTS.find((e) => e.value === event)?.label || event}
                        </Badge>
                      ))
                    )}
                  </div>
                  <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-[400px] p-0" align="start">
                <Command>
                  <CommandInput placeholder="Search events..." />
                  <CommandList>
                    <CommandEmpty>No events found.</CommandEmpty>
                    {['Sandbox', 'Snapshot', 'Volume'].map((category) => (
                      <CommandGroup key={category} heading={category}>
                        {AVAILABLE_EVENTS.filter((event) => event.category === category).map((event) => (
                          <CommandCheckboxItem
                            key={event.value}
                            checked={eventTypes.includes(event.value)}
                            onSelect={() => toggleEvent(event.value)}
                          >
                            {event.label}
                          </CommandCheckboxItem>
                        ))}
                      </CommandGroup>
                    ))}
                    {eventTypes.length > 0 && (
                      <>
                        <CommandSeparator />
                        <CommandGroup>
                          <CommandItem onSelect={() => setEventTypes([])} className="justify-center text-center">
                            Clear selection
                          </CommandItem>
                        </CommandGroup>
                      </>
                    )}
                  </CommandList>
                </Command>
              </PopoverContent>
            </Popover>
          </div>
        </form>
        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="secondary">
              Cancel
            </Button>
          </DialogClose>
          {updateMutation.isPending ? (
            <Button type="button" variant="default" disabled>
              Saving...
            </Button>
          ) : (
            <Button type="submit" form="edit-endpoint-form" variant="default" disabled={!url.trim()}>
              Save
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
