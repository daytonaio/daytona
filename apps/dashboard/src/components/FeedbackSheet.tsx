/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { Field, FieldDescription, FieldError, FieldLabel } from '@/components/ui/field'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { Spinner } from '@/components/ui/spinner'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { usePylon } from '@/vendor/pylon'
import { useForm } from '@tanstack/react-form'
import { MessageSquareText } from 'lucide-react'
import { usePostHog } from 'posthog-js/react'
import { type ReactNode, useRef, useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { useLocation } from 'react-router-dom'
import { toast } from 'sonner'
import { z } from 'zod'

const feedbackTypeSchema = z.enum(['issue', 'idea'])

const formSchema = z.object({
  type: feedbackTypeSchema,
  message: z.string().trim().min(1, 'Feedback is required'),
})

type FeedbackType = z.infer<typeof feedbackTypeSchema>
type FormValues = z.infer<typeof formSchema>

const defaultValues: FormValues = {
  type: 'issue',
  message: '',
}

const getFeedbackCopy = (type: FeedbackType) => {
  if (type === 'issue') {
    return {
      label: 'What happened?',
      description: 'Describe the problem, where you saw it, and what you expected instead.',
      placeholder: 'Something is not working when...',
    }
  }

  return {
    label: 'What should we build or improve?',
    description: 'Share the workflow, improvement, or rough idea. Short notes are fine.',
    placeholder: 'It would be helpful if Daytona could...',
  }
}

export function FeedbackSheet({ children }: { children?: ReactNode }) {
  const posthog = usePostHog()
  const { user } = useAuth()
  const location = useLocation()
  const { toggle: togglePylon } = usePylon()
  const [open, setOpen] = useState(false)
  const formRef = useRef<HTMLFormElement>(null)

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
      const message = value.message.trim()

      posthog?.capture('dashboard_feedback_submitted', {
        feedback_type: value.type,
        message,
        path: location.pathname,
        url: window.location.href,
        user_email: user?.profile.email,
        user_name: user?.profile.name,
      })

      toast.success('Feedback sent')
      resetForm(defaultValues)
      setOpen(false)
    },
  })
  const { reset: resetForm } = form

  const handleGetHelpInstead = () => {
    setOpen(false)
    window.setTimeout(() => {
      togglePylon()
    }, 150)
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        {children ?? (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="px-2 text-muted-foreground hover:text-foreground md:px-3"
            aria-label="Feedback"
          >
            <MessageSquareText className="size-4" />
            <span className="hidden md:inline">Feedback</span>
          </Button>
        )}
      </SheetTrigger>
      <SheetContent className="flex w-dvw flex-col gap-0 p-0 sm:w-[440px]">
        <SheetHeader className="border-b border-border p-4 px-5 items-center flex text-left flex-row">
          <SheetTitle>Send feedback</SheetTitle>
          <SheetDescription className="sr-only">Submit product feedback to Daytona.</SheetDescription>
        </SheetHeader>
        <form
          ref={formRef}
          id="feedback-form"
          className="flex min-h-0 flex-1 flex-col"
          onSubmit={(e) => {
            e.preventDefault()
            e.stopPropagation()
            form.handleSubmit()
          }}
        >
          <div className="flex flex-1 flex-col gap-5 p-5">
            <form.Field name="type">
              {(field) => (
                <Tabs value={field.state.value} onValueChange={(value) => field.handleChange(value as FeedbackType)}>
                  <TabsList className="grid w-full grid-cols-2">
                    <TabsTrigger value="issue">Issue</TabsTrigger>
                    <TabsTrigger value="idea">Idea</TabsTrigger>
                  </TabsList>
                </Tabs>
              )}
            </form.Field>

            <form.Subscribe selector={(state) => state.values.type}>
              {(feedbackType) => {
                const copy = getFeedbackCopy(feedbackType)

                return (
                  <form.Field name="message">
                    {(field) => {
                      const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid

                      return (
                        <Field data-invalid={isInvalid}>
                          <FieldLabel htmlFor={field.name}>{copy.label}</FieldLabel>
                          <FieldDescription>{copy.description}</FieldDescription>
                          <Textarea
                            id={field.name}
                            name={field.name}
                            value={field.state.value}
                            onBlur={field.handleBlur}
                            onChange={(event) => field.handleChange(event.target.value)}
                            placeholder={copy.placeholder}
                            aria-invalid={isInvalid}
                            className="min-h-36 resize-none"
                          />
                          {field.state.meta.errors.length > 0 && field.state.meta.isTouched && (
                            <FieldError errors={field.state.meta.errors} />
                          )}
                          {feedbackType === 'issue' ? (
                            <div className="flex justify-start">
                              <button
                                type="button"
                                className="w-fit text-sm font-medium text-primary underline-offset-4 hover:underline"
                                onClick={handleGetHelpInstead}
                              >
                                Get help instead
                              </button>
                            </div>
                          ) : null}
                        </Field>
                      )
                    }}
                  </form.Field>
                )
              }}
            </form.Subscribe>
          </div>
        </form>
        <SheetFooter className="border-t border-border p-4 px-5">
          <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <form.Subscribe
            selector={(state) => [state.canSubmit, state.isSubmitting] as const}
            children={([canSubmit, isSubmitting]) => (
              <Button type="submit" form="feedback-form" variant="default" disabled={!canSubmit || isSubmitting}>
                {isSubmitting && <Spinner />}
                Submit
              </Button>
            )}
          />
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
