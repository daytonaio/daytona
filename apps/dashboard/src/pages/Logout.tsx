import LoadingFallback from '@/components/LoadingFallback'
import { useEffect } from 'react'
import { useAuth } from 'react-oidc-context'

const Logout = () => {
  const { signoutRedirect } = useAuth()

  useEffect(() => {
    void signoutRedirect()
  }, [signoutRedirect])

  return <LoadingFallback />
}

export default Logout
