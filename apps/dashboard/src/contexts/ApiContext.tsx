import { ApiClient } from '@/api/apiClient'
import { createContext } from 'react'

export const ApiContext = createContext<ApiClient | null>(null)
