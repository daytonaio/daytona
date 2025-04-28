import { ReactNode, useCallback, useMemo, useState } from 'react'
import { suspend } from 'suspend-react'
import { useApi } from '@/hooks/useApi'
import { OrganizationsContext, IOrganizationsContext } from '@/contexts/OrganizationsContext'
import { Organization } from '@daytonaio/api-client'
import { handleApiError } from '@/lib/error-handling'

type Props = {
  children: ReactNode
}

export function OrganizationsProvider(props: Props) {
  const { organizationsApi } = useApi()

  const getOrganizations = useCallback(async () => {
    try {
      return (await organizationsApi.listOrganizations()).data
    } catch (error) {
      handleApiError(error, 'Failed to fetch your organizations')
      return []
    }
  }, [organizationsApi])

  const [organizations, setOrganizations] = useState<Organization[]>(
    suspend(getOrganizations, [organizationsApi, 'organizations']),
  )

  const refreshOrganizations = useCallback(async () => {
    const orgs = await getOrganizations()
    setOrganizations(orgs)
    return orgs
  }, [getOrganizations])

  const contextValue: IOrganizationsContext = useMemo(() => {
    return {
      organizations,
      setOrganizations,
      refreshOrganizations,
    }
  }, [organizations, refreshOrganizations])

  return <OrganizationsContext.Provider value={contextValue}>{props.children}</OrganizationsContext.Provider>
}
