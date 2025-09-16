import { PlaygroundSandboxParamsContext, PlaygroundSandboxParams, SetPlaygroundSandboxParamsValue } from './context'
import { useState } from 'react'

export const PlaygroundSandboxParamsProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [playgroundSandboxParametersState, setPlaygroundSandboxParametersState] = useState<PlaygroundSandboxParams>({
    // Default values
    resources: {
      cpu: 2,
      // gpu: 0,
      memory: 4,
      disk: 8,
    },
    createSandboxBaseParams: {
      autoStopInterval: 15,
      autoArchiveInterval: 7,
      autoDeleteInterval: -1,
    },
  })

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
