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
  DialogTrigger,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { ChevronsUpDown, Plus } from 'lucide-react'
import React, { useState } from 'react'
import { useNewEndpoint } from 'svix-react'

const AVAILABLE_EVENTS = [
  { value: 'sandbox.created', label: 'Sandbox Created', category: 'Sandbox' },
  { value: 'sandbox.state.updated', label: 'Sandbox State Updated', category: 'Sandbox' },
  { value: 'snapshot.created', label: 'Snapshot Created', category: 'Snapshot' },
  { value: 'snapshot.removed', label: 'Snapshot Removed', category: 'Snapshot' },
  { value: 'snapshot.state.updated', label: 'Snapshot State Updated', category: 'Snapshot' },
  { value: 'volume.created', label: 'Volume Created', category: 'Volume' },
  { value: 'volume.state.updated', label: 'Volume State Updated', category: 'Volume' },
]

interface CreateEndpointDialogProps {
  onSuccess: () => void
  className?: string
}

export const CreateEndpointDialog: React.FC<CreateEndpointDialogProps> = ({ onSuccess, className }) => {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [eventsPopoverOpen, setEventsPopoverOpen] = useState(false)

  const { url, description, eventTypes, createEndpoint } = useNewEndpoint()

  const handleCreateEndpoint = async () => {
    if (!url.value.trim()) {
      return
    }

    setLoading(true)
    try {
      const result = await createEndpoint()
      if (result.endpoint) {
        onSuccess()
        setOpen(false)
        // Reset form
        url.setValue('')
        description.setValue('')
        eventTypes.setValue([])
      }
    } catch (error) {
      console.error('Failed to create endpoint:', error)
    } finally {
      setLoading(false)
    }
  }

  const toggleEvent = (eventValue: string) => {
    const currentEvents = eventTypes.value || []
    if (currentEvents.includes(eventValue)) {
      eventTypes.setValue(currentEvents.filter((e) => e !== eventValue))
    } else {
      eventTypes.setValue([...currentEvents, eventValue])
    }
  }

  const selectedEvents = eventTypes.value || []

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        setOpen(isOpen)
        if (!isOpen) {
          url.setValue('')
          description.setValue('')
          eventTypes.setValue([])
        }
      }}
    >
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
          onSubmit={async (e) => {
            e.preventDefault()
            await handleCreateEndpoint()
          }}
        >
          <div className="space-y-3">
            <Label htmlFor="endpoint-name">Endpoint Name</Label>
            <Input
              id="endpoint-name"
              type="text"
              value={description.value}
              onChange={(e) => description.setValue(e.target.value)}
              placeholder="My Webhook Endpoint"
            />
          </div>
          <div className="space-y-3">
            <Label htmlFor="endpoint-url">Endpoint URL</Label>
            <Input
              id="endpoint-url"
              type="url"
              value={url.value}
              onChange={(e) => url.setValue(e.target.value)}
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
                    {selectedEvents.length === 0 ? (
                      <span className="text-muted-foreground">Select events...</span>
                    ) : selectedEvents.length > 2 ? (
                      <Badge variant="secondary" className="rounded-sm px-1 font-normal">
                        {selectedEvents.length} events selected
                      </Badge>
                    ) : (
                      selectedEvents.map((event) => (
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
                            checked={selectedEvents.includes(event.value)}
                            onSelect={() => toggleEvent(event.value)}
                          >
                            {event.label}
                          </CommandCheckboxItem>
                        ))}
                      </CommandGroup>
                    ))}
                    {selectedEvents.length > 0 && (
                      <>
                        <CommandSeparator />
                        <CommandGroup>
                          <CommandItem onSelect={() => eventTypes.setValue([])} className="justify-center text-center">
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
          {loading ? (
            <Button type="button" variant="default" disabled>
              Creating...
            </Button>
          ) : (
            <Button type="submit" form="create-endpoint-form" variant="default" disabled={!url.value.trim()}>
              Create
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
