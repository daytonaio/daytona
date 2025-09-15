/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import SandboxCodeSnippetsResponse from '@/components/Playground/Sandbox/CodeSnippetsResponse'
import SandboxParameters from '@/components/Playground/Sandbox/Parameters'
import { PlaygroundCategories, playgroundCategoriesData } from '@/enums/Playground'
import { PlaygroundSandboxParamsProvider } from '@/components/Playground/Sandbox/provider'
import { useState } from 'react'

const Playground: React.FC = () => {
  const [playgroundCategory, setPlaygroundCategory] = useState<PlaygroundCategories>(PlaygroundCategories.SANDBOX)

  return (
    <div className="px-6 py-2">
      <div className="mb-2 h-12 flex items-center justify-between">
        <h1 className="text-2xl font-medium">Playground</h1>
      </div>
      <PlaygroundSandboxParamsProvider>
        <div className="w-full flex flex-col space-y-2 md:flex-row md:space-x-2 md:space-y-0">
          <Card className="basis-full md:basis-1/3 md:max-w-[33.33%] flex-shrink-0">
            <CardHeader>
              <div className="w-full flex items-center justify-center overflow-x-auto space-x-4">
                {playgroundCategoriesData.map((category) => (
                  <Button
                    variant={category.value === playgroundCategory ? 'default' : 'secondary'}
                    className="text-md"
                    onClick={() => setPlaygroundCategory(category.value)}
                  >
                    {category.label}
                  </Button>
                ))}
              </div>
            </CardHeader>
            <CardContent>{playgroundCategory === PlaygroundCategories.SANDBOX && <SandboxParameters />}</CardContent>
          </Card>
          <div className="flex-1 min-w-0 flex flex-col space-y-2">
            {playgroundCategory === PlaygroundCategories.SANDBOX && <SandboxCodeSnippetsResponse />}
          </div>
        </div>
      </PlaygroundSandboxParamsProvider>
    </div>
  )
}

export default Playground
