import { Organization } from '@daytonaio/api-client'
import { createContext } from 'react'

export interface IOrganizationsContext {
  organizations: Organization[]
  refreshOrganizations: () => Promise<Organization[]>
}

export const OrganizationsContext = createContext<IOrganizationsContext | undefined>(undefined)
