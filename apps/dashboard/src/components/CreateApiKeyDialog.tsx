/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { DatePicker } from '@/components/ui/date-picker'
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
import { Spinner } from '@/components/ui/spinner'
import { AnimatePresence, motion } from 'framer-motion'
import { CheckIcon, CopyIcon, EyeIcon, EyeOffIcon, InfoIcon } from 'lucide-react'

import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { InputGroup, InputGroupButton, InputGroupInput } from '@/components/ui/input-group'
import { Label } from '@/components/ui/label'
import { CREATE_API_KEY_PERMISSIONS_GROUPS } from '@/constants/CreateApiKeyPermissionsGroups'
import { useCreateApiKeyMutation } from '@/hooks/mutations/useCreateApiKeyMutation'
import { useCopyToClipboard } from '@/hooks/useCopyToClipboard'
import { handleApiError } from '@/lib/error-handling'
import { getMaskedToken } from '@/lib/utils'
import { ApiKeyResponse, CreateApiKeyPermissionsEnum } from '@daytonaio/api-client'
import { useForm } from '@tanstack/react-form'
import { Plus } from 'lucide-react'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { toast } from 'sonner'
import { z } from 'zod'
import { Alert, AlertDescription, AlertTitle } from './ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from './ui/tabs'
import { ToggleGroup, ToggleGroupItem } from './ui/toggle-group'

interface CreateApiKeyDialogProps {
  availablePermissions: CreateApiKeyPermissionsEnum[]
  apiUrl: string
  className?: string
  organizationId?: string
}

const isReadPermission = (permission: CreateApiKeyPermissionsEnum) => permission.startsWith('read:')
const isWritePermission = (permission: CreateApiKeyPermissionsEnum) => permission.startsWith('write:')
const isDeletePermission = (permission: CreateApiKeyPermissionsEnum) => permission.startsWith('delete:')

const IMPLICIT_READ_RESOURCES = ['Sandboxes', 'Snapshots', 'Registries', 'Regions']

const formSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  expiresAt: z.date().optional(),
  permissions: z.array(z.enum(CreateApiKeyPermissionsEnum)),
})

type FormValues = z.infer<typeof formSchema>

export const CreateApiKeyDialog: React.FC<CreateApiKeyDialogProps> = ({
  availablePermissions,
  apiUrl,
  className,
  organizationId,
}) => {
  const [open, setOpen] = useState(false)

  const { reset: resetCreateApiKeyMutation, ...createApiKeyMutation } = useCreateApiKeyMutation()

  const availableGroups = useMemo(() => {
    return CREATE_API_KEY_PERMISSIONS_GROUPS.map((group) => ({
      ...group,
      permissions: group.permissions.filter((p) => availablePermissions.includes(p)),
    })).filter((group) => group.permissions.length > 0)
  }, [availablePermissions])

  const form = useForm({
    defaultValues: {
      name: '',
      expiresAt: undefined,
      permissions: availablePermissions,
    } as FormValues,
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: async ({ value }) => {
      if (!organizationId) {
        toast.error('Select an organization to create an API key.')
        return
      }

      try {
        await createApiKeyMutation.mutateAsync({
          organizationId,
          name: value.name.trim(),
          permissions: value.permissions,
          expiresAt: value.expiresAt ?? null,
        })

        toast.success('API key created successfully')
      } catch (error) {
        handleApiError(error, 'Failed to create API key')
      }
    },
  })

  const resetState = useCallback(() => {
    form.reset({
      name: '',
      expiresAt: undefined,
      permissions: availablePermissions,
    })
    resetCreateApiKeyMutation()
  }, [resetCreateApiKeyMutation, form, availablePermissions])

  useEffect(() => {
    if (open) {
      resetState()
    }
  }, [open, resetState])

  const createdKey = createApiKeyMutation.data

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        setOpen(isOpen)
      }}
    >
      <DialogTrigger asChild>
        <Button variant="default" size="sm" title="Create Key" className={className}>
          <Plus className="w-4 h-4" />
          Create Key
        </Button>
      </DialogTrigger>

      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{createdKey ? 'API Key Created' : 'Create New API Key'}</DialogTitle>
          <DialogDescription>
            {createdKey
              ? 'Your API key has been created successfully.'
              : 'Choose which actions this API key will be authorized to perform.'}
          </DialogDescription>
        </DialogHeader>
        {createdKey ? (
          <CreatedKeyDisplay createdKey={createdKey} apiUrl={apiUrl} key={createdKey.value} />
        ) : (
          <div className="overflow-y-auto px-1">
            <form
              id="create-api-key-form"
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
                      <FieldLabel htmlFor={field.name}>Key Name</FieldLabel>
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

              <form.Field name="expiresAt">
                {(field) => (
                  <Field>
                    <FieldLabel htmlFor={field.name}>Expires</FieldLabel>
                    <DatePicker
                      id={field.name}
                      value={field.state.value}
                      onChange={field.handleChange}
                      disabledBefore={new Date()}
                    />
                    <FieldDescription>Optional expiration date for the API key.</FieldDescription>
                  </Field>
                )}
              </form.Field>

              <Tabs
                defaultValue="full-access"
                className="items-start pb-4"
                onValueChange={(value) => {
                  if (value === 'full-access') {
                    form.setFieldValue('permissions', availablePermissions)
                  } else if (value === 'sandbox-access') {
                    form.setFieldValue('permissions', [
                      CreateApiKeyPermissionsEnum.WRITE_SANDBOXES,
                      CreateApiKeyPermissionsEnum.DELETE_SANDBOXES,
                    ])
                  } else {
                    form.setFieldValue('permissions', [])
                  }
                }}
              >
                <Label className="mb-1">Permissions</Label>

                <TabsList className="bg-muted w-full [&>*]:flex-1">
                  <TabsTrigger value="full-access">Full Access</TabsTrigger>
                  <TabsTrigger value="sandbox-access">Sandboxes</TabsTrigger>
                  <TabsTrigger value="restricted-access">Restricted </TabsTrigger>
                </TabsList>
                <TabsContent value="sandbox-access" className="w-full">
                  <Alert variant="info">
                    <InfoIcon />
                    <AlertTitle>Sandboxes Access</AlertTitle>
                    <AlertDescription>
                      This key grants read and write access to the Sandboxes resource.
                    </AlertDescription>
                  </Alert>
                </TabsContent>
                <TabsContent value="full-access" className="w-full">
                  <Alert variant="info">
                    <InfoIcon />
                    <AlertTitle>Full Access</AlertTitle>
                    <AlertDescription>
                      This key grants full access to all resources. For better security, we recommend creating a
                      restricted key.
                    </AlertDescription>
                  </Alert>
                </TabsContent>
                <TabsContent value="restricted-access" className="w-full">
                  {availableGroups.length > 0 && (
                    <form.Field name="permissions">
                      {(field) => (
                        <Field data-invalid={field.state.meta.isTouched && !field.state.meta.isValid}>
                          <FieldGroup className="gap-4 xs:gap-2">
                            {availableGroups.map((group) => {
                              const readPermission = group.permissions.find(isReadPermission)
                              const writePermission = group.permissions.find(isWritePermission)
                              const deletePermission = group.permissions.find(isDeletePermission)
                              const hasImplicitRead = IMPLICIT_READ_RESOURCES.includes(group.name)

                              return (
                                <div
                                  key={group.name}
                                  className="flex gap-2 justify-between xs:items-center flex-col xs:flex-row"
                                >
                                  <Label htmlFor={`group-${group.name}`} className="font-normal">
                                    {group.name}
                                  </Label>

                                  <ToggleGroup
                                    type="multiple"
                                    variant="outline"
                                    size="sm"
                                    spacing={0}
                                    value={group.permissions.filter((p) => field.state.value.includes(p))}
                                    onValueChange={(newGroupSelection) => {
                                      const permissionsWithoutThisGroup = field.state.value.filter(
                                        (p) => !group.permissions.includes(p),
                                      )

                                      field.handleChange([
                                        ...permissionsWithoutThisGroup,
                                        ...newGroupSelection,
                                      ] as CreateApiKeyPermissionsEnum[])
                                    }}
                                  >
                                    {hasImplicitRead ? (
                                      <ToggleGroupItem
                                        value=""
                                        aria-label="Implicit read access"
                                        className="min-w-[64px]"
                                        disabled
                                        data-state="on"
                                      >
                                        Read*
                                      </ToggleGroupItem>
                                    ) : (
                                      <ToggleGroupItem
                                        value={readPermission ?? ''}
                                        aria-label="Toggle read"
                                        className="min-w-[64px]"
                                        disabled={!readPermission}
                                      >
                                        {readPermission ? 'Read' : '-'}
                                      </ToggleGroupItem>
                                    )}
                                    <ToggleGroupItem
                                      value={writePermission ?? ''}
                                      aria-label="Toggle write"
                                      className="min-w-[64px]"
                                      disabled={!writePermission}
                                    >
                                      {writePermission ? 'Write' : '-'}
                                    </ToggleGroupItem>
                                    <ToggleGroupItem
                                      value={deletePermission ?? ''}
                                      aria-label="Toggle delete"
                                      className="min-w-[64px]"
                                      disabled={!deletePermission}
                                    >
                                      {deletePermission ? 'Delete' : '-'}
                                    </ToggleGroupItem>
                                  </ToggleGroup>
                                </div>
                              )
                            })}
                          </FieldGroup>
                          {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                            <FieldError errors={field.state.meta.errors} />
                          )}
                          <p className="text-sm text-muted-foreground mt-3">
                            *Read access is always granted for these resources.
                          </p>
                        </Field>
                      )}
                    </form.Field>
                  )}
                </TabsContent>
              </Tabs>
            </form>
          </div>
        )}
        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="secondary">
              Close
            </Button>
          </DialogClose>
          {!createdKey && (
            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
                <Button
                  type="submit"
                  form="create-api-key-form"
                  variant="default"
                  disabled={!canSubmit || isSubmitting || !organizationId}
                >
                  {isSubmitting && <Spinner />}
                  Create
                </Button>
              )}
            />
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

const MotionCopyIcon = motion(CopyIcon)
const MotionCheckIcon = motion(CheckIcon)

const iconProps = {
  initial: { opacity: 0, y: 5 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -5 },
  transition: { duration: 0.1 },
}

function CreatedKeyDisplay({ createdKey, apiUrl }: { createdKey: ApiKeyResponse; apiUrl: string }) {
  const [copiedApiKey, copyApiKey] = useCopyToClipboard()
  const [copiedApiUrl, copyApiUrl] = useCopyToClipboard()

  const [apiKeyRevealed, setApiKeyRevealed] = useState(false)

  return (
    <div className="space-y-6">
      <Alert variant="warning">
        <InfoIcon />
        <AlertDescription>You can only view this key once. Store it safely.</AlertDescription>
      </Alert>
      <FieldGroup className="gap-4">
        <Field>
          <FieldLabel htmlFor="api-key">API Key</FieldLabel>

          <InputGroup className="pr-1 flex-1">
            <InputGroupInput
              id="api-key"
              value={apiKeyRevealed ? createdKey.value : getMaskedToken(createdKey.value)}
              readOnly
            />
            <InputGroupButton variant="ghost" size="icon-xs" onClick={() => setApiKeyRevealed(!apiKeyRevealed)}>
              {apiKeyRevealed ? <EyeOffIcon className="h-4 w-4" /> : <EyeIcon className="h-4 w-4" />}
            </InputGroupButton>
            <InputGroupButton variant="ghost" size="icon-xs" onClick={() => copyApiKey(createdKey.value)}>
              <AnimatePresence initial={false} mode="wait">
                {copiedApiKey ? (
                  <MotionCheckIcon className="h-4 w-4" key="copied" {...iconProps} />
                ) : (
                  <MotionCopyIcon className="h-4 w-4" key="copy" {...iconProps} />
                )}
              </AnimatePresence>
            </InputGroupButton>
          </InputGroup>
        </Field>

        <Field>
          <FieldLabel htmlFor="api-url">API URL</FieldLabel>

          <InputGroup className="pr-1 flex-1">
            <InputGroupInput id="api-url" value={apiUrl} readOnly />
            <InputGroupButton variant="ghost" size="icon-xs" onClick={() => copyApiUrl(apiUrl)}>
              <AnimatePresence initial={false} mode="wait">
                {copiedApiUrl ? (
                  <MotionCheckIcon className="h-4 w-4" key="copied" {...iconProps} />
                ) : (
                  <MotionCopyIcon className="h-4 w-4" key="copy" {...iconProps} />
                )}
              </AnimatePresence>
            </InputGroupButton>
          </InputGroup>
        </Field>
      </FieldGroup>
    </div>
  )
}
