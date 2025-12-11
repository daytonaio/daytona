/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useState, useEffect } from 'react'
import { Region, CreateRunner, CreateRunnerResponse } from '@daytonaio/api-client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { toast } from 'sonner'
import { Plus, Copy } from 'lucide-react'
import { getMaskedToken } from '@/lib/utils'

const DEFAULT_FORM_DATA = {
  name: '',
  regionId: '',
}

interface CreateRunnerDialogProps {
  regions: Region[]
  onCreateRunner: (data: CreateRunner) => Promise<CreateRunnerResponse | null>
}

export const CreateRunnerDialog: React.FC<CreateRunnerDialogProps> = ({ regions, onCreateRunner }) => {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)

  const [createdRunner, setCreatedRunner] = useState<CreateRunnerResponse | null>(null)
  const [isTokenRevealed, setIsTokenRevealed] = useState(false)

  const [formData, setFormData] = useState(DEFAULT_FORM_DATA)
  const [formErrors, setFormErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    if (regions.length > 0 && !formData.regionId) {
      setFormData((prev) => ({ ...prev, regionId: regions[0].id }))
    }
  }, [regions, formData.regionId])

  const validateForm = () => {
    const errors: Record<string, string> = {}

    if (!formData.name.trim()) {
      errors.name = 'Name is required'
    }

    if (!formData.regionId) {
      errors.regionId = 'Region is required'
    }

    setFormErrors(errors)
    return Object.keys(errors).length === 0
  }

  const handleCreate = async () => {
    if (!validateForm()) {
      return
    }

    setLoading(true)
    try {
      const runner = await onCreateRunner({
        name: formData.name,
        regionId: formData.regionId,
      })
      if (runner) {
        setCreatedRunner(runner)
        setFormData({
          ...DEFAULT_FORM_DATA,
          regionId: regions.length > 0 ? regions[0].id : '',
        })
        setFormErrors({})
      }
    } finally {
      setLoading(false)
    }
  }

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      toast.success('Copied to clipboard')
    } catch (err) {
      console.error('Failed to copy text:', err)
      toast.error('Failed to copy to clipboard')
    }
  }

  if (regions.length === 0) {
    return null
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        setOpen(isOpen)
        if (!isOpen) {
          setCreatedRunner(null)
          setFormData({
            ...DEFAULT_FORM_DATA,
            regionId: regions.length > 0 ? regions[0].id : '',
          })
          setFormErrors({})
        }
      }}
    >
      <DialogTrigger asChild>
        <Button variant="default" size="sm" className="w-auto px-4" title="Create Runner">
          <Plus className="w-4 h-4" />
          Create Runner
        </Button>
      </DialogTrigger>

      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Create New Runner</DialogTitle>
          <DialogDescription>Add configuration for a new runner in your selected region.</DialogDescription>
        </DialogHeader>

        {createdRunner ? (
          <div className="space-y-6">
            <div className="space-y-3">
              <Label htmlFor="token">Token</Label>
              <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                <span
                  className="overflow-x-auto pr-2 cursor-text select-all"
                  onMouseEnter={() => setIsTokenRevealed(true)}
                  onMouseLeave={() => setIsTokenRevealed(false)}
                >
                  {isTokenRevealed ? createdRunner.apiKey : getMaskedToken(createdRunner.apiKey)}
                </span>
                <Copy
                  className="w-4 h-4 cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                  onClick={() => copyToClipboard(createdRunner.apiKey)}
                />
              </div>
              <p className="text-sm text-muted-foreground">
                Save this token securely. You won't be able to see it again.
              </p>
            </div>
          </div>
        ) : (
          <form
            id="create-runner-form"
            className="space-y-6 overflow-y-auto px-1 pb-1"
            onSubmit={async (e) => {
              e.preventDefault()
              await handleCreate()
            }}
          >
            <div className="space-y-3">
              <Label htmlFor="regionId">Region</Label>
              <Select
                value={formData.regionId}
                onValueChange={(value) => {
                  setFormData((prev) => ({ ...prev, regionId: value }))
                  if (formErrors.regionId) {
                    setFormErrors((prev) => ({ ...prev, regionId: '' }))
                  }
                }}
              >
                <SelectTrigger className={`h-8 ${formErrors.regionId ? 'border-destructive' : ''}`}>
                  <SelectValue placeholder="Select a region" />
                </SelectTrigger>
                <SelectContent>
                  {regions.map((region) => (
                    <SelectItem key={region.id} value={region.id}>
                      {region.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {formErrors.regionId && <p className="text-sm text-destructive">{formErrors.regionId}</p>}
            </div>

            <div className="space-y-3">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => {
                  setFormData((prev) => ({ ...prev, name: e.target.value }))
                }}
                placeholder="runner-1"
              />
              {formErrors.name && <p className="text-sm text-destructive">{formErrors.name}</p>}
            </div>
          </form>
        )}

        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="secondary">
              {createdRunner ? 'Close' : 'Cancel'}
            </Button>
          </DialogClose>
          {!createdRunner &&
            (loading ? (
              <Button type="button" variant="default" disabled>
                Creating...
              </Button>
            ) : (
              <Button type="submit" form="create-runner-form" variant="default" disabled={loading}>
                Create
              </Button>
            ))}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
