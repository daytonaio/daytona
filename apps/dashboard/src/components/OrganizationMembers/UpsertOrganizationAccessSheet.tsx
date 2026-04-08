/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Field, FieldError, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTrigger,
  SheetTitle,
} from '@/components/ui/sheet'
import { Spinner } from '@/components/ui/spinner'
import { useOrganizationRolesQuery } from '@/hooks/queries/useOrganizationRolesQuery'
import { CreateOrganizationInvitationRoleEnum, OrganizationRole } from '@daytona/api-client'
import { useForm } from '@tanstack/react-form'
import { AlertTriangle, Plus } from 'lucide-react'
import React, {
  Ref,
  type ReactNode,
  useCallback,
  useEffect,
  useImperativeHandle,
  useMemo,
  useRef,
  useState,
} from 'react'
import { z } from 'zod'

const baseFormSchema = z.object({
  role: z.enum(CreateOrganizationInvitationRoleEnum),
  assignedRoleIds: z.array(z.string()),
})

const formSchema = baseFormSchema.extend({
  email: z.email('A valid email address is required'),
})
type FormValues = z.infer<typeof formSchema>

type UpsertOrganizationAccessSheetMode = 'create' | 'edit'

interface InitialMemberAccess {
  email: string
  role: CreateOrganizationInvitationRoleEnum
  assignedRoleIds: string[]
}

interface UpsertOrganizationAccessSheetProps {
  mode?: UpsertOrganizationAccessSheetMode
  onSubmit: (payload: {
    email: string
    role: CreateOrganizationInvitationRoleEnum
    assignedRoleIds: string[]
  }) => Promise<boolean>
  className?: string
  disabled?: boolean
  trigger?: ReactNode | null
  ref?: Ref<{ open: () => void }>
  open?: boolean
  onOpenChange?: (open: boolean) => void
  initialMember?: Partial<InitialMemberAccess>
  title?: ReactNode
  description?: ReactNode
  reducedRoleWarning?: ReactNode
}

const getDefaultAssignedRoleIds = (availableRoles: OrganizationRole[]) => {
  const defaultRole = availableRoles.find((availableRole) => availableRole.name === 'Developer')
  return defaultRole ? [defaultRole.id] : []
}

export const UpsertOrganizationAccessSheet: React.FC<UpsertOrganizationAccessSheetProps> = ({
  mode = 'create',
  onSubmit,
  className,
  disabled,
  trigger,
  ref,
  open,
  onOpenChange,
  initialMember,
  title,
  description,
  reducedRoleWarning,
}) => {
  const { data: availableRoles = [], isLoading: loadingAvailableRoles } = useOrganizationRolesQuery()
  const [internalOpen, setInternalOpen] = useState(false)
  const formRef = useRef<HTMLFormElement>(null)
  const wasOpenRef = useRef(false)

  const isCreateMode = mode === 'create'
  const isControlled = open !== undefined
  const isOpen = open ?? internalOpen

  const defaultAssignedRoleIds = useMemo(() => getDefaultAssignedRoleIds(availableRoles), [availableRoles])

  const defaultValues = useMemo<FormValues>(() => {
    const role = initialMember?.role ?? CreateOrganizationInvitationRoleEnum.MEMBER
    const assignedRoleIds =
      initialMember?.assignedRoleIds ??
      (isCreateMode && role !== CreateOrganizationInvitationRoleEnum.OWNER ? defaultAssignedRoleIds : [])

    return {
      email: initialMember?.email ?? '',
      role,
      assignedRoleIds,
    }
  }, [defaultAssignedRoleIds, initialMember, isCreateMode])

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
      const role = value.role
      const assignedRoleIds = role === CreateOrganizationInvitationRoleEnum.OWNER ? [] : value.assignedRoleIds
      const success = await onSubmit({
        email: value.email.trim(),
        role,
        assignedRoleIds,
      })

      if (!success) {
        return
      }

      handleOpenChange(false)
      resetForm(defaultValues)
    },
  })

  const handleAssignedRolesOnRoleChange = useCallback(
    (previousRole: CreateOrganizationInvitationRoleEnum, nextRole: CreateOrganizationInvitationRoleEnum) => {
      const currentAssignedRoleIds = form.getFieldValue('assignedRoleIds')

      if (nextRole === CreateOrganizationInvitationRoleEnum.OWNER) {
        if (currentAssignedRoleIds.length > 0) {
          form.setFieldValue('assignedRoleIds', [])
        }
        return
      }

      if (currentAssignedRoleIds.length > 0) {
        return
      }

      const fallbackAssignedRoleIds =
        initialMember?.assignedRoleIds && initialMember.assignedRoleIds.length > 0
          ? initialMember.assignedRoleIds
          : isCreateMode
            ? defaultAssignedRoleIds
            : []

      const switchedFromOwnerToMember =
        previousRole === CreateOrganizationInvitationRoleEnum.OWNER &&
        nextRole === CreateOrganizationInvitationRoleEnum.MEMBER

      if (switchedFromOwnerToMember) {
        form.setFieldValue('assignedRoleIds', fallbackAssignedRoleIds)
        return
      }

      if (fallbackAssignedRoleIds.length === 0) {
        return
      }

      form.setFieldValue('assignedRoleIds', fallbackAssignedRoleIds)
    },
    [defaultAssignedRoleIds, form, initialMember?.assignedRoleIds, isCreateMode],
  )

  const { reset: resetForm } = form

  const resetState = useCallback(() => {
    resetForm(defaultValues)
  }, [defaultValues, resetForm])

  useEffect(() => {
    if (isOpen && !wasOpenRef.current) {
      resetState()
    }
    wasOpenRef.current = isOpen
  }, [isOpen, resetState])

  const initialAssignedRoleIdSet = useMemo(() => new Set(initialMember?.assignedRoleIds ?? []), [initialMember])

  const resolvedTitle = title ?? (isCreateMode ? 'Invite Member' : 'Update Access')

  const resolvedDescription =
    description ??
    (isCreateMode
      ? 'Give them access to the organization with an appropriate role and assignments.'
      : 'Manage access to the organization with an appropriate role and assignments.')

  const submitLabel = isCreateMode ? 'Invite' : 'Save'

  const formId = `${mode}-organization-access-form`

  return (
    <Sheet open={isOpen} onOpenChange={handleOpenChange}>
      {trigger === undefined ? (
        <SheetTrigger asChild>
          <Button variant="default" size="sm" className={className} disabled={disabled}>
            {isCreateMode && <Plus className="w-4 h-4" />}
            {isCreateMode ? 'Invite Member' : 'Update Access'}
          </Button>
        </SheetTrigger>
      ) : (
        trigger
      )}

      <SheetContent className="w-dvw sm:w-[560px] p-0 flex flex-col gap-0">
        <SheetHeader className="border-b border-border p-4 px-5 items-center flex text-left flex-row">
          <SheetTitle className="text-2xl">{resolvedTitle}</SheetTitle>
          <SheetDescription className="sr-only">{resolvedDescription}</SheetDescription>
        </SheetHeader>

        <ScrollArea fade="mask" className="flex-1 min-h-0">
          <form
            ref={formRef}
            id={formId}
            className="gap-6 flex flex-col p-5"
            onSubmit={(e) => {
              e.preventDefault()
              e.stopPropagation()
              form.handleSubmit()
            }}
          >
            {isCreateMode ? (
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
            ) : initialMember?.email ? (
              <Field>
                <FieldLabel htmlFor="email">Email</FieldLabel>
                <Input id="email" value={initialMember.email} type="email" disabled readOnly />
              </Field>
            ) : null}

            <form.Field name="role">
              {(field) => (
                <Field>
                  <FieldLabel htmlFor={field.name}>Role</FieldLabel>
                  <RadioGroup
                    className="gap-6"
                    value={field.state.value}
                    onValueChange={(value: CreateOrganizationInvitationRoleEnum) => {
                      const previousRole = field.state.value
                      field.handleChange(value)
                      handleAssignedRolesOnRoleChange(previousRole, value)
                    }}
                  >
                    <div className="grid grid-cols-[auto_1fr] items-start gap-4">
                      <RadioGroupItem value={CreateOrganizationInvitationRoleEnum.OWNER} id="role-owner" />
                      <div className="flex flex-col gap-1">
                        <Label htmlFor="role-owner" className="font-normal">
                          Owner
                        </Label>
                        <p className="text-sm text-muted-foreground">
                          Full administrative access to the organization and its resources
                        </p>
                      </div>
                    </div>
                    <div className="grid grid-cols-[auto_1fr] items-start gap-4">
                      <RadioGroupItem value={CreateOrganizationInvitationRoleEnum.MEMBER} id="role-member" />
                      <div className="flex flex-col gap-1">
                        <Label htmlFor="role-member" className="font-normal">
                          Member
                        </Label>
                        <p className="text-sm text-muted-foreground">
                          Access to organization resources is based on assignments
                        </p>
                      </div>
                    </div>
                  </RadioGroup>
                </Field>
              )}
            </form.Field>

            <form.Subscribe
              selector={(state) => ({ role: state.values.role, assignedRoleIds: state.values.assignedRoleIds })}
            >
              {({ role, assignedRoleIds }) => {
                const effectiveAssignedRoleIds =
                  role === CreateOrganizationInvitationRoleEnum.OWNER ? [] : assignedRoleIds

                const hasRemovedAssignments = Array.from(initialAssignedRoleIdSet).some(
                  (roleId) => !effectiveAssignedRoleIds.includes(roleId),
                )

                return (
                  <>
                    {role === CreateOrganizationInvitationRoleEnum.MEMBER && !loadingAvailableRoles && (
                      <form.Field name="assignedRoleIds">
                        {(field) => {
                          const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid
                          return (
                            <Field data-invalid={isInvalid}>
                              <FieldLabel htmlFor={field.name}>Assignments</FieldLabel>
                              <div className="grid gap-6">
                                <div className="grid grid-cols-[auto_1fr] items-start gap-4">
                                  <Checkbox id="role-viewer" checked={true} disabled={true} />
                                  <div className="flex flex-col gap-1">
                                    <Label htmlFor="role-viewer" className="font-normal">
                                      Viewer
                                    </Label>
                                    <p className="text-sm text-muted-foreground">
                                      Grants read access to sandboxes, snapshots, and registries in the organization
                                    </p>
                                  </div>
                                </div>
                                {availableRoles.map((availableRole) => (
                                  <div key={availableRole.id} className="grid grid-cols-[auto_1fr] items-start gap-4">
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
                                    <div className="flex flex-col gap-1">
                                      <Label htmlFor={`role-${availableRole.id}`} className="font-normal">
                                        {availableRole.name}
                                      </Label>
                                      {availableRole.description && (
                                        <p className="text-sm text-muted-foreground">{availableRole.description}</p>
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
                    )}

                    {hasRemovedAssignments && reducedRoleWarning && (
                      <Alert variant="warning">
                        <AlertTriangle className="h-4 w-4" />
                        <AlertDescription>{reducedRoleWarning}</AlertDescription>
                      </Alert>
                    )}
                  </>
                )
              }}
            </form.Subscribe>
          </form>
        </ScrollArea>

        <SheetFooter className="border-t border-border p-4 px-5">
          <Button type="button" size="sm" variant="secondary" onClick={() => handleOpenChange(false)}>
            Cancel
          </Button>
          <form.Subscribe
            selector={(state) => [state.canSubmit, state.isSubmitting]}
            children={([canSubmit, isSubmitting]) => (
              <Button type="submit" size="sm" form={formId} variant="default" disabled={!canSubmit || isSubmitting}>
                {isSubmitting && <Spinner />}
                {submitLabel}
              </Button>
            )}
          />
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
