/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ViewerOrganizationRoleCheckbox } from '@/components/OrganizationMembers/ViewerOrganizationRoleCheckbox'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Field, FieldError, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
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
import { CreateOrganizationInvitationRoleEnum, OrganizationRole } from '@daytonaio/api-client'
import { useForm } from '@tanstack/react-form'
import { useMutation } from '@tanstack/react-query'
import { Plus } from 'lucide-react'
import React, { Ref, useCallback, useImperativeHandle, useMemo, useRef, useState } from 'react'
import { z } from 'zod'

interface CreateOrganizationInvitationSheetProps {
  availableRoles: OrganizationRole[]
  loadingAvailableRoles: boolean
  onCreateInvitation: (
    email: string,
    role: CreateOrganizationInvitationRoleEnum,
    assignedRoleIds: string[],
  ) => Promise<boolean>
  className?: string
  ref?: Ref<{ open: () => void }>
}

const formSchema = z
  .object({
    email: z.string().trim().min(1, 'A valid email address is required').email('A valid email address is required'),
    role: z.enum(CreateOrganizationInvitationRoleEnum),
    assignedRoleIds: z.array(z.string()),
  })
  .superRefine((value, ctx) => {
    if (value.role === CreateOrganizationInvitationRoleEnum.MEMBER && value.assignedRoleIds.length === 0) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['assignedRoleIds'],
        message: 'Select at least one assignment',
      })
    }
  })

type FormValues = z.infer<typeof formSchema>

const defaultValues: FormValues = {
  email: '',
  role: CreateOrganizationInvitationRoleEnum.MEMBER,
  assignedRoleIds: [],
}

export const CreateOrganizationInvitationSheet: React.FC<CreateOrganizationInvitationSheetProps> = ({
  availableRoles,
  loadingAvailableRoles,
  onCreateInvitation,
  className,
  ref,
}) => {
  const [open, setOpen] = useState(false)
  const formRef = useRef<HTMLFormElement>(null)

  useImperativeHandle(ref, () => ({
    open: () => setOpen(true),
  }))

  const defaultAssignedRoleIds = useMemo(() => {
    const role = availableRoles.find((availableRole) => availableRole.name === 'Developer')
    return role ? [role.id] : []
  }, [availableRoles])

  const { reset: resetCreateInvitationMutation, ...createInvitationMutation } = useMutation({
    mutationFn: async (value: FormValues) => {
      return onCreateInvitation(
        value.email.trim(),
        value.role,
        value.role === CreateOrganizationInvitationRoleEnum.OWNER ? [] : value.assignedRoleIds,
      )
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
      const success = await createInvitationMutation.mutateAsync(value)
      if (!success) {
        return
      }

      setOpen(false)
      resetForm({
        ...defaultValues,
        assignedRoleIds: defaultAssignedRoleIds,
      })
    },
  })
  const { reset: resetForm } = form

  const resetState = useCallback(() => {
    resetForm({
      ...defaultValues,
      assignedRoleIds: defaultAssignedRoleIds,
    })
    resetCreateInvitationMutation()
  }, [resetForm, resetCreateInvitationMutation, defaultAssignedRoleIds])

  const handleOpenChange = useCallback(
    (isOpen: boolean) => {
      setOpen(isOpen)
      if (isOpen) {
        resetState()
      }
    },
    [resetState],
  )

  return (
    <Sheet open={open} onOpenChange={handleOpenChange}>
      <SheetTrigger asChild>
        <Button variant="default" size="sm" className={className} title="Invite Member">
          <Plus className="w-4 h-4" />
          Invite Member
        </Button>
      </SheetTrigger>

      <SheetContent className="w-dvw sm:w-[560px] p-0 flex flex-col gap-0">
        <SheetHeader className="border-b border-border p-4 px-5 items-center flex text-left flex-row">
          <SheetTitle className="text-2xl">Invite Member</SheetTitle>
          <SheetDescription className="sr-only">
            Give them access to the organization with an appropriate role and assignments.
          </SheetDescription>
        </SheetHeader>

        <ScrollArea fade="mask" className="flex-1 min-h-0">
          <form
            ref={formRef}
            id="invitation-form"
            className="gap-6 flex flex-col p-5"
            onSubmit={(e) => {
              e.preventDefault()
              e.stopPropagation()
              form.handleSubmit()
            }}
          >
            <form.Field name="email">
              {(field) => {
                const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                return (
                  <Field data-invalid={isInvalid}>
                    <FieldLabel htmlFor={field.name}>Email</FieldLabel>
                    <Input
                      aria-invalid={isInvalid}
                      id={field.name}
                      name={field.name}
                      type="email"
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      placeholder="mail@example.com"
                    />
                    {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                      <FieldError errors={field.state.meta.errors} />
                    )}
                  </Field>
                )
              }}
            </form.Field>

            <form.Field name="role">
              {(field) => (
                <Field>
                  <FieldLabel htmlFor={field.name}>Role</FieldLabel>
                  <RadioGroup
                    className="gap-6"
                    value={field.state.value}
                    onValueChange={(value: CreateOrganizationInvitationRoleEnum) => {
                      field.handleChange(value)
                      if (value === CreateOrganizationInvitationRoleEnum.OWNER) {
                        form.setFieldValue('assignedRoleIds', [])
                      } else if (defaultAssignedRoleIds.length > 0) {
                        form.setFieldValue('assignedRoleIds', defaultAssignedRoleIds)
                      }
                    }}
                  >
                    <div className="flex items-center space-x-4">
                      <RadioGroupItem value={CreateOrganizationInvitationRoleEnum.OWNER} id="role-owner" />
                      <div className="space-y-1">
                        <Label htmlFor="role-owner" className="font-normal">
                          Owner
                        </Label>
                        <p className="text-sm text-gray-500">
                          Full administrative access to the organization and its resources
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center space-x-4">
                      <RadioGroupItem value={CreateOrganizationInvitationRoleEnum.MEMBER} id="role-member" />
                      <div className="space-y-1">
                        <Label htmlFor="role-member" className="font-normal">
                          Member
                        </Label>
                        <p className="text-sm text-gray-500">
                          Access to organization resources is based on assignments
                        </p>
                      </div>
                    </div>
                  </RadioGroup>
                </Field>
              )}
            </form.Field>

            <form.Subscribe selector={(state) => state.values.role}>
              {(role) =>
                role === CreateOrganizationInvitationRoleEnum.MEMBER && !loadingAvailableRoles ? (
                  <form.Field name="assignedRoleIds">
                    {(field) => {
                      const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                      return (
                        <Field data-invalid={isInvalid}>
                          <FieldLabel htmlFor={field.name}>Assignments</FieldLabel>
                          <div className="space-y-6">
                            <ViewerOrganizationRoleCheckbox />
                            {availableRoles.map((availableRole) => (
                              <div key={availableRole.id} className="flex items-center space-x-4">
                                <Checkbox
                                  id={`role-${availableRole.id}`}
                                  checked={field.state.value.includes(availableRole.id)}
                                  onCheckedChange={() => {
                                    if (field.state.value.includes(availableRole.id)) {
                                      field.handleChange(
                                        field.state.value.filter((roleId) => roleId !== availableRole.id),
                                      )
                                      return
                                    }

                                    field.handleChange([...field.state.value, availableRole.id])
                                  }}
                                />
                                <div className="space-y-1">
                                  <Label htmlFor={`role-${availableRole.id}`} className="font-normal">
                                    {availableRole.name}
                                  </Label>
                                  {availableRole.description && (
                                    <p className="text-sm text-gray-500">{availableRole.description}</p>
                                  )}
                                </div>
                              </div>
                            ))}
                          </div>
                          {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                            <FieldError errors={field.state.meta.errors} />
                          )}
                        </Field>
                      )
                    }}
                  </form.Field>
                ) : null
              }
            </form.Subscribe>
          </form>
        </ScrollArea>

        <SheetFooter className="border-t border-border p-4 px-5">
          <Button type="button" size="sm" variant="secondary" onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <form.Subscribe
            selector={(state) => [state.canSubmit, state.isSubmitting]}
            children={([canSubmit, isSubmitting]) => (
              <Button
                type="submit"
                size="sm"
                form="invitation-form"
                variant="default"
                disabled={!canSubmit || isSubmitting}
              >
                {isSubmitting && <Spinner />}
                Invite
              </Button>
            )}
          />
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
