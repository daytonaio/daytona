/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Field, FieldError, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Spinner } from '@/components/ui/spinner'
import { ORGANIZATION_ROLE_PERMISSIONS_GROUPS } from '@/constants/OrganizationPermissionsGroups'
import { OrganizationRolePermissionGroup } from '@/types/OrganizationRolePermissionGroup'
import { OrganizationRolePermissionsEnum } from '@daytona/api-client'
import { useForm } from '@tanstack/react-form'
import { useMutation } from '@tanstack/react-query'
import { Plus } from 'lucide-react'
import React, { Ref, useCallback, useEffect, useImperativeHandle, useRef, useState } from 'react'
import { z } from 'zod'

interface CreateOrganizationRoleSheetProps {
  onCreateRole: (name: string, description: string, permissions: OrganizationRolePermissionsEnum[]) => Promise<boolean>
  className?: string
  ref?: Ref<{ open: () => void }>
}

const formSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  description: z.string().min(1, 'Description is required'),
  permissions: z.array(z.enum(OrganizationRolePermissionsEnum)).min(1, 'At least one permission is required'),
})

type FormValues = z.infer<typeof formSchema>

const defaultValues: FormValues = {
  name: '',
  description: '',
  permissions: [],
}

export const CreateOrganizationRoleSheet: React.FC<CreateOrganizationRoleSheetProps> = ({
  onCreateRole,
  className,
  ref,
}) => {
  const [open, setOpen] = useState(false)
  const formRef = useRef<HTMLFormElement>(null)

  useImperativeHandle(ref, () => ({
    open: () => setOpen(true),
  }))

  const createRoleMutation = useMutation({
    mutationFn: async (value: FormValues) => {
      return onCreateRole(value.name.trim(), value.description.trim(), value.permissions)
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
      const success = await createRoleMutation.mutateAsync(value)
      if (!success) {
        return
      }

      setOpen(false)
      resetForm(defaultValues)
    },
  })
  const { reset: resetForm } = form

  const { reset: resetMutation } = createRoleMutation

  const resetState = useCallback(() => {
    resetForm(defaultValues)
    resetMutation()
  }, [resetForm, resetMutation])

  useEffect(() => {
    if (open) {
      resetState()
    }
  }, [open, resetState])

  const isGroupChecked = (group: OrganizationRolePermissionGroup, permissions: OrganizationRolePermissionsEnum[]) => {
    return group.permissions.every((permission) => permissions.includes(permission))
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="default" size="sm" className={className}>
          <Plus className="w-4 h-4" />
          Create Role
        </Button>
      </SheetTrigger>
      <SheetContent className="w-dvw sm:w-[560px] p-0 flex flex-col gap-0">
        <SheetHeader className="border-b border-border p-4 px-5 items-center flex text-left flex-row">
          <SheetTitle className="text-2xl">Create Role</SheetTitle>
          <SheetDescription className="sr-only">
            Define a custom role for managing access to the organization.
          </SheetDescription>
        </SheetHeader>

        <ScrollArea fade="mask" className="flex-1 min-h-0">
          <form
            ref={formRef}
            id="create-role-form"
            className="space-y-6 p-5"
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
                    <FieldLabel htmlFor={field.name}>Name</FieldLabel>
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

            <form.Field name="description">
              {(field) => {
                const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                return (
                  <Field data-invalid={isInvalid}>
                    <FieldLabel htmlFor={field.name}>Description</FieldLabel>
                    <Input
                      aria-invalid={isInvalid}
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      placeholder="Description"
                    />
                    {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                      <FieldError errors={field.state.meta.errors} />
                    )}
                  </Field>
                )
              }}
            </form.Field>

            <form.Field name="permissions">
              {(field) => {
                const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid

                const handleGroupToggle = (group: OrganizationRolePermissionGroup) => {
                  if (isGroupChecked(group, field.state.value)) {
                    field.handleChange(
                      field.state.value.filter((permission) => !group.permissions.includes(permission)),
                    )
                    return
                  }

                  const newPermissions = [...field.state.value]
                  group.permissions.forEach((permission) => {
                    if (!newPermissions.includes(permission)) {
                      newPermissions.push(permission)
                    }
                  })
                  field.handleChange(newPermissions)
                }

                const handlePermissionToggle = (permission: OrganizationRolePermissionsEnum) => {
                  if (field.state.value.includes(permission)) {
                    field.handleChange(field.state.value.filter((current) => current !== permission))
                    return
                  }

                  field.handleChange([...field.state.value, permission])
                }

                return (
                  <Field data-invalid={isInvalid}>
                    <FieldLabel htmlFor={field.name}>Permissions</FieldLabel>
                    <div className="space-y-6">
                      {ORGANIZATION_ROLE_PERMISSIONS_GROUPS.map((group) => {
                        const groupIsChecked = isGroupChecked(group, field.state.value)

                        return (
                          <div key={group.name} className="space-y-3">
                            <div className="flex items-center space-x-2">
                              <Checkbox
                                id={`group-${group.name}`}
                                checked={groupIsChecked}
                                onCheckedChange={() => handleGroupToggle(group)}
                              />
                              <Label htmlFor={`group-${group.name}`} className="font-normal">
                                {group.name}
                              </Label>
                            </div>
                            <div className="ml-6 space-y-2">
                              {group.permissions.map((permission) => (
                                <div key={permission} className="flex items-center space-x-2">
                                  <Checkbox
                                    id={permission}
                                    checked={field.state.value.includes(permission)}
                                    onCheckedChange={() => handlePermissionToggle(permission)}
                                    disabled={groupIsChecked}
                                    className={groupIsChecked ? 'pointer-events-none' : ''}
                                  />
                                  <Label
                                    htmlFor={permission}
                                    className={`font-normal${groupIsChecked ? ' opacity-70 pointer-events-none' : ''}`}
                                  >
                                    {permission}
                                  </Label>
                                </div>
                              ))}
                            </div>
                          </div>
                        )
                      })}
                    </div>
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
              <Button type="submit" form="create-role-form" variant="default" disabled={!canSubmit || isSubmitting}>
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
