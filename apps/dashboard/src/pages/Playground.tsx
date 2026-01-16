/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { PageContent, PageHeader, PageLayout, PageTitle } from '@/components/PageLayout'
import {
  PlaygroundLayout,
  PlaygroundLayoutContent,
  PlaygroundLayoutSidebar,
} from '@/components/Playground/PlaygroundLayout'
import SandboxCodeSnippetsResponse from '@/components/Playground/Sandbox/CodeSnippetsResponse'
import SandboxParameters from '@/components/Playground/Sandbox/Parameters'
import TerminalDescription from '@/components/Playground/Terminal/Description'
import WebTerminal from '@/components/Playground/Terminal/WebTerminal'
import VNCDesktopWindowResponse from '@/components/Playground/VNC/DesktopWindowResponse'
import VNCInteractionOptions from '@/components/Playground/VNC/Interaction'
import { Button } from '@/components/ui/button'
import { Drawer, DrawerContent } from '@/components/ui/drawer'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { PlaygroundCategories, playgroundCategoriesData } from '@/enums/Playground'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { PlaygroundProvider } from '@/providers/PlaygroundProvider'
import { AnimatePresence, motion } from 'framer-motion'
import { SettingsIcon } from 'lucide-react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useResizeObserver } from 'usehooks-ts'

const SlideLeftRight = ({ children, direction }: { children: React.ReactNode; direction: 'left' | 'right' }) => {
  return (
    <motion.div
      initial={{ opacity: 0, filter: 'blur(2px)', x: direction === 'left' ? -20 : 20 }}
      animate={{ opacity: 1, filter: 'blur(0px)', x: 0 }}
      exit={{ opacity: 0, filter: 'blur(2px)', x: direction === 'left' ? 20 : -20 }}
      transition={{ duration: 0.2 }}
    >
      {children}
    </motion.div>
  )
}

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

  const [drawerOpen, setDrawerOpen] = useState<PlaygroundCategories | null>(null)
  const handleDrawerOpenChange = (open: boolean) => {
    if (!open) {
      setDrawerOpen(null)
    }
  }

  const pageContentRef = useRef<HTMLDivElement>(null)

  useResizeObserver({
    ref: pageContentRef,
    onResize: () => {
      if (pageContentRef.current) {
        const { width } = pageContentRef.current.getBoundingClientRect()
        if (width < 1024) {
          setDrawerOpen(null)
        }
      }
    },
  })

  const prevCategory = useRef<PlaygroundCategories>(playgroundCategory)
  useEffect(() => {
    prevCategory.current = playgroundCategory
  }, [playgroundCategory])

  const direction = useMemo(() => {
    const currentIndex = playgroundCategoriesData.findIndex((category) => category.value === playgroundCategory)
    const prevIndex = playgroundCategoriesData.findIndex((category) => category.value === prevCategory.current)
    return currentIndex > prevIndex ? 'right' : 'left'
  }, [playgroundCategory])

  return (
    <PageLayout>
      <PageHeader>
        <PageTitle>Playground</PageTitle>
      </PageHeader>

      <PageContent
        size="full"
        className="!p-0 h-full flex flex-col flex-1 max-h-[calc(100vh-111px)]"
        ref={pageContentRef}
      >
        <PlaygroundProvider>
          <Tabs
            value={playgroundCategory}
            onValueChange={(value) => setPlaygroundCategory(value as PlaygroundCategories)}
            className="h-full"
          >
            <div className="flex items-center justify-between shadow-[inset_0_-1px] shadow-border pr-4">
              <TabsList className="px-2 w-full shadow-none">
                {playgroundCategoriesData.map((category) => (
                  <TabsTrigger
                    value={category.value}
                    key={category.value}
                    className="data-[state=inactive]:border-b-transparent"
                  >
                    {category.label}
                  </TabsTrigger>
                ))}
              </TabsList>
              <Button onClick={() => setDrawerOpen(playgroundCategory)} variant="ghost" size="sm" className="lg:hidden">
                <SettingsIcon className="size-4 mr-2" /> Configure
              </Button>
            </div>
            <TabsContent
              value={playgroundCategory}
              key={playgroundCategory}
              className="mt-0 data-[state=inactive]:hidden"
              asChild
            >
              <PlaygroundLayout>
                <PlaygroundLayoutSidebar>
                  <AnimatePresence mode="popLayout">
                    {playgroundCategory === PlaygroundCategories.SANDBOX && (
                      <SlideLeftRight direction={direction} key="sandbox-parameters">
                        <SandboxParameters />
                      </SlideLeftRight>
                    )}
                    {playgroundCategory === PlaygroundCategories.TERMINAL && (
                      <SlideLeftRight direction={direction} key="terminal-description">
                        <TerminalDescription />
                      </SlideLeftRight>
                    )}
                    {playgroundCategory === PlaygroundCategories.VNC && (
                      <SlideLeftRight direction={direction} key="vnc-interaction-options">
                        <VNCInteractionOptions />
                      </SlideLeftRight>
                    )}
                  </AnimatePresence>
                </PlaygroundLayoutSidebar>

                <Drawer open={drawerOpen === playgroundCategory} onOpenChange={handleDrawerOpenChange}>
                  <DrawerContent>
                    <div className="p-4 overflow-auto">
                      {playgroundCategory === PlaygroundCategories.SANDBOX && <SandboxParameters />}
                      {playgroundCategory === PlaygroundCategories.TERMINAL && <TerminalDescription />}
                      {playgroundCategory === PlaygroundCategories.VNC && <VNCInteractionOptions />}
                    </div>
                  </DrawerContent>
                </Drawer>
                <PlaygroundLayoutContent>
                  {playgroundCategory === PlaygroundCategories.SANDBOX && (
                    <SandboxCodeSnippetsResponse className="w-full max-w-[90%]" />
                  )}
                  {playgroundCategory === PlaygroundCategories.TERMINAL && (
                    <WebTerminal getPortPreviewUrl={getPortPreviewUrl} className="w-full max-w-[90%]" />
                  )}
                  {playgroundCategory === PlaygroundCategories.VNC && (
                    <VNCDesktopWindowResponse getPortPreviewUrl={getPortPreviewUrl} className="w-full max-w-[90%]" />
                  )}
                </PlaygroundLayoutContent>
              </PlaygroundLayout>
            </TabsContent>
          </Tabs>
        </PlaygroundProvider>
      </PageContent>
    </PageLayout>
  )
}

export default Playground
