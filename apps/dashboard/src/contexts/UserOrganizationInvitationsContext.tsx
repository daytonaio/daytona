import { createContext } from 'react'

export interface IUserOrganizationInvitationsContext {
  count: number
  setCount(count: number): void
}

export const UserOrganizationInvitationsContext = createContext<IUserOrganizationInvitationsContext | undefined>(
  undefined,
)
