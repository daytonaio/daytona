import type { Meta, StoryObj } from '@storybook/react'
import { Field, FieldContent, FieldDescription, FieldError, FieldGroup, FieldLabel } from '../field'
import { Input } from '../input'

const meta: Meta<typeof Field> = {
  title: 'UI/Field',
  component: Field,
}

export default meta
type Story = StoryObj<typeof Field>

export const Vertical: Story = {
  render: () => (
    <FieldGroup className="max-w-sm">
      <Field orientation="vertical">
        <FieldLabel htmlFor="email">Email</FieldLabel>
        <FieldContent>
          <Input id="email" placeholder="you@example.com" />
          <FieldDescription>We'll never share your email.</FieldDescription>
        </FieldContent>
      </Field>
    </FieldGroup>
  ),
}

export const Horizontal: Story = {
  render: () => (
    <FieldGroup className="max-w-md">
      <Field orientation="horizontal">
        <FieldLabel htmlFor="name">Name</FieldLabel>
        <FieldContent>
          <Input id="name" placeholder="John Doe" />
        </FieldContent>
      </Field>
    </FieldGroup>
  ),
}

export const WithError: Story = {
  render: () => (
    <FieldGroup className="max-w-sm">
      <Field orientation="vertical">
        <FieldLabel htmlFor="password">Password</FieldLabel>
        <FieldContent>
          <Input id="password" type="password" aria-invalid="true" />
          <FieldError>Password must be at least 8 characters.</FieldError>
        </FieldContent>
      </Field>
    </FieldGroup>
  ),
}
