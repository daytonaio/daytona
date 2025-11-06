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
  domain: '',
  apiUrl: '',
  proxyUrl: '',
  cpu: 16,
  memoryGiB: 64,
  diskGiB: 4000,
  regionId: '',
}

interface CreateRunnerDialogProps {
  regions: Region[]
  onCreateRunner: (data: CreateRunner) => Promise<CreateRunnerResponse | null>
  writePermitted: boolean
}

export const CreateRunnerDialog: React.FC<CreateRunnerDialogProps> = ({ regions, onCreateRunner, writePermitted }) => {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)

  const [createdRunner, setCreatedRunner] = useState<CreateRunnerResponse | null>(null)
  const [isApiKeyRevealed, setIsApiKeyRevealed] = useState(false)

  const [formData, setFormData] = useState(DEFAULT_FORM_DATA)
  const [formErrors, setFormErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    if (regions.length > 0 && !formData.regionId) {
      setFormData((prev) => ({ ...prev, regionId: regions[0].id }))
    }
  }, [regions, formData.regionId])

  const validateForm = () => {
    const errors: Record<string, string> = {}

    if (!formData.domain.trim()) {
      errors.domain = 'Domain is required'
    }

    if (!formData.apiUrl.trim()) {
      errors.apiUrl = 'API URL is required'
    } else if (!formData.apiUrl.startsWith('http')) {
      errors.apiUrl = 'API URL must start with http:// or https://'
    }

    if (!formData.proxyUrl.trim()) {
      errors.proxyUrl = 'Proxy URL is required'
    } else if (!formData.proxyUrl.startsWith('http')) {
      errors.proxyUrl = 'Proxy URL must start with http:// or https://'
    }

    if (formData.cpu < 16) {
      errors.cpu = 'vCPU must be at least 16 cores'
    }

    if (formData.memoryGiB < 64) {
      errors.memoryGiB = 'Memory must be at least 64 GiB'
    }

    if (formData.diskGiB < 4000) {
      errors.diskGiB = 'Disk must be at least 4000 GiB'
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
        domain: formData.domain,
        apiUrl: formData.apiUrl,
        proxyUrl: formData.proxyUrl,
        cpu: formData.cpu,
        memoryGiB: formData.memoryGiB,
        diskGiB: formData.diskGiB,
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

  if (!writePermitted || regions.length === 0) {
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
              <Label htmlFor="api-key">API Key</Label>
              <div className="p-3 flex justify-between items-center rounded-md bg-green-100 text-green-600 dark:bg-green-900/50 dark:text-green-400">
                <span
                  className="overflow-x-auto pr-2 cursor-text select-all"
                  onMouseEnter={() => setIsApiKeyRevealed(true)}
                  onMouseLeave={() => setIsApiKeyRevealed(false)}
                >
                  {isApiKeyRevealed ? createdRunner.token : getMaskedToken(createdRunner.token)}
                </span>
                <Copy
                  className="w-4 h-4 cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                  onClick={() => copyToClipboard(createdRunner.token)}
                />
              </div>
              <p className="text-sm text-muted-foreground">
                Save this API key securely. You won't be able to see it again.
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
              <Label htmlFor="domain">Domain</Label>
              <Input
                id="domain"
                value={formData.domain}
                onChange={(e) => {
                  setFormData((prev) => ({ ...prev, domain: e.target.value }))
                  if (formErrors.domain) {
                    setFormErrors((prev) => ({ ...prev, domain: '' }))
                  }
                }}
                placeholder="runner.example.com"
                className={formErrors.domain ? 'border-destructive' : ''}
              />
              {formErrors.domain && <p className="text-sm text-destructive">{formErrors.domain}</p>}
            </div>

            <div className="space-y-3">
              <Label htmlFor="apiUrl">API URL</Label>
              <Input
                id="apiUrl"
                value={formData.apiUrl}
                onChange={(e) => {
                  setFormData((prev) => ({ ...prev, apiUrl: e.target.value }))
                  if (formErrors.apiUrl) {
                    setFormErrors((prev) => ({ ...prev, apiUrl: '' }))
                  }
                }}
                placeholder="https://api.runner.example.com"
                className={formErrors.apiUrl ? 'border-destructive' : ''}
              />
              {formErrors.apiUrl && <p className="text-sm text-destructive">{formErrors.apiUrl}</p>}
            </div>

            <div className="space-y-3">
              <Label htmlFor="proxyUrl">Proxy URL</Label>
              <Input
                id="proxyUrl"
                value={formData.proxyUrl}
                onChange={(e) => {
                  setFormData((prev) => ({ ...prev, proxyUrl: e.target.value }))
                  if (formErrors.proxyUrl) {
                    setFormErrors((prev) => ({ ...prev, proxyUrl: '' }))
                  }
                }}
                placeholder="https://proxy.runner.example.com"
                className={formErrors.proxyUrl ? 'border-destructive' : ''}
              />
              {formErrors.proxyUrl && <p className="text-sm text-destructive">{formErrors.proxyUrl}</p>}
            </div>

            <div className="grid grid-cols-3 gap-4">
              <div className="space-y-3">
                <Label htmlFor="cpu">vCPU</Label>
                <Input
                  id="cpu"
                  type="number"
                  min="1"
                  value={formData.cpu}
                  onChange={(e) => {
                    setFormData((prev) => ({ ...prev, cpu: parseInt(e.target.value) || 1 }))
                    if (formErrors.cpu) {
                      setFormErrors((prev) => ({ ...prev, cpu: '' }))
                    }
                  }}
                  className={formErrors.cpu ? 'border-destructive' : ''}
                />
                {formErrors.cpu && <p className="text-sm text-destructive">{formErrors.cpu}</p>}
              </div>

              <div className="space-y-3">
                <Label htmlFor="memoryGiB">Memory (GiB)</Label>
                <Input
                  id="memoryGiB"
                  type="number"
                  min="1"
                  value={formData.memoryGiB}
                  onChange={(e) => {
                    setFormData((prev) => ({ ...prev, memoryGiB: parseInt(e.target.value) || 1 }))
                    if (formErrors.memoryGiB) {
                      setFormErrors((prev) => ({ ...prev, memoryGiB: '' }))
                    }
                  }}
                  className={formErrors.memoryGiB ? 'border-destructive' : ''}
                />
                {formErrors.memoryGiB && <p className="text-sm text-destructive">{formErrors.memoryGiB}</p>}
              </div>

              <div className="space-y-3">
                <Label htmlFor="diskGiB">Disk (GiB)</Label>
                <Input
                  id="diskGiB"
                  type="number"
                  min="1"
                  value={formData.diskGiB}
                  onChange={(e) => {
                    setFormData((prev) => ({ ...prev, diskGiB: parseInt(e.target.value) || 1 }))
                    if (formErrors.diskGiB) {
                      setFormErrors((prev) => ({ ...prev, diskGiB: '' }))
                    }
                  }}
                  className={formErrors.diskGiB ? 'border-destructive' : ''}
                />
                {formErrors.diskGiB && <p className="text-sm text-destructive">{formErrors.diskGiB}</p>}
              </div>
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
