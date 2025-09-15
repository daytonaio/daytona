import { PlaygroundSandboxParamsContext, PlaygroundSandboxParams, SetPlaygroundSandboxParamsValue } from './context'
import { useState } from 'react'

export const PlaygroundSandboxParamsProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [playgroundSandboxParametersState, setPlaygroundSandboxParametersState] = useState<PlaygroundSandboxParams>({})

  const setPlaygroundSandboxParameterValue: SetPlaygroundSandboxParamsValue = (key, value) => {
    setPlaygroundSandboxParametersState((prev) => ({ ...prev, [key]: value }))
  }
  return (
    <PlaygroundSandboxParamsContext.Provider
      value={{ playgroundSandboxParametersState, setPlaygroundSandboxParameterValue }}
    >
      {children}
    </PlaygroundSandboxParamsContext.Provider>
  )
}
