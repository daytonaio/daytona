/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import SandboxCodeSnippetsResponse from '@/components/Playground/Sandbox/CodeSnippetsResponse'
import SandboxParameters from '@/components/Playground/Sandbox/Parameters'
import TerminalDescription from '@/components/Playground/Terminal/Description'
import WebTerminal from '@/components/Playground/Terminal/WebTerminal'
import VNCInteractionOptions from '@/components/Playground/VNC/Interaction'
import VNCDesktopWindowResponse from '@/components/Playground/VNC/DesktopWindowResponse'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { PlaygroundCategories, playgroundCategoriesData } from '@/enums/Playground'
import { PlaygroundProvider } from '@/providers/PlaygroundProvider'
import { useState, useCallback } from 'react'

const Playground: React.FC = () => {
  const [playgroundCategory, setPlaygroundCategory] = useState<PlaygroundCategories>(PlaygroundCategories.SANDBOX)

  const { sandboxApi } = useApi()

  const { selectedOrganization } = useSelectedOrganization()

  const getPortPreviewUrl = useCallback(
    async (sandboxId: string, port: number): Promise<string> => {
      return (await sandboxApi.getPortPreviewUrl(sandboxId, port, selectedOrganization?.id)).data.url
    },
    [sandboxApi, selectedOrganization],
  )

  return (
    <div className="flex flex-col min-h-dvh px-6 py-2">
      <div className="mb-2 h-12 flex items-center justify-between">
        <h1 className="text-2xl font-medium">Playground</h1>
      </div>
      <PlaygroundProvider>
        <div className="w-full flex flex-1 flex-col space-y-2 md:flex-row md:space-x-2 md:space-y-0">
          <Card className="basis-full md:basis-1/3 md:max-w-[33.33%] flex-shrink-0 min-h-full">
            <CardHeader>
              <div className="w-full flex items-center justify-center overflow-x-auto space-x-4">
                {playgroundCategoriesData.map((category) => (
                  <Button
                    key={category.value}
                    variant={category.value === playgroundCategory ? 'default' : 'secondary'}
                    className="text-md"
                    onClick={() => setPlaygroundCategory(category.value)}
                  >
                    {category.label}
                  </Button>
                ))}
              </div>
            </CardHeader>
            <CardContent>
              {playgroundCategory === PlaygroundCategories.SANDBOX && <SandboxParameters />}
              {playgroundCategory === PlaygroundCategories.TERMINAL && <TerminalDescription />}
              {playgroundCategory === PlaygroundCategories.VNC && <VNCInteractionOptions />}
            </CardContent>
          </Card>
          <div className="flex-1 min-w-0 flex flex-col space-y-2">
            {playgroundCategory === PlaygroundCategories.SANDBOX && <SandboxCodeSnippetsResponse />}
            {playgroundCategory === PlaygroundCategories.TERMINAL && (
              <WebTerminal sandboxId="93eb77da-54a2-4f34-872e-6c9843be0228" getPortPreviewUrl={getPortPreviewUrl} />
            )}
            {playgroundCategory === PlaygroundCategories.VNC && (
              <VNCDesktopWindowResponse
                sandboxId="dda297bb-6afc-4439-a3ac-1c2b42ec5f2c"
                getPortPreviewUrl={getPortPreviewUrl}
              />
            )}
          </div>
        </div>
      </PlaygroundProvider>
    </div>
  )
}

export default Playground
