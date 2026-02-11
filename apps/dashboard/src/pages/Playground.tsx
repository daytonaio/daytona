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
import { PlaygroundCategories } from '@/enums/Playground'
import { PlaygroundProvider } from '@/providers/PlaygroundProvider'
import { AnimatePresence, motion } from 'framer-motion'
import { SettingsIcon } from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useResizeObserver } from 'usehooks-ts'

const playgroundCategoriesData = [
  { value: PlaygroundCategories.SANDBOX, label: 'Sandbox' },
  { value: PlaygroundCategories.TERMINAL, label: 'Terminal' },
  { value: PlaygroundCategories.VNC, label: 'VNC' },
]

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

  const [drawerOpen, setDrawerOpen] = useState<PlaygroundCategories | null>(null)
  const handleDrawerOpenChange = (open: boolean) => {
    if (!open) {
      setDrawerOpen(null)
    }
  }

  const pageContentRef = useRef<HTMLDivElement>(null)

  useResizeObserver({
    ref: pageContentRef as React.RefObject<HTMLElement>,
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

  const sidePanel = useMemo(() => {
    if (playgroundCategory === PlaygroundCategories.SANDBOX) return <SandboxParameters />
    if (playgroundCategory === PlaygroundCategories.TERMINAL) return <TerminalDescription />
    if (playgroundCategory === PlaygroundCategories.VNC) return <VNCInteractionOptions />
    return null
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
              <TabsList className="px-2 shadow-none bg-transparent w-auto pb-0">
                {playgroundCategoriesData.map((category) => (
                  <TabsTrigger
                    value={category.value}
                    key={category.value}
                    className="data-[state=inactive]:border-b-transparent data-[state=active]:border-b-foreground border-b rounded-none !shadow-none -mb-0.5"
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
                    <SlideLeftRight direction={direction} key={playgroundCategory}>
                      {sidePanel}
                    </SlideLeftRight>
                  </AnimatePresence>
                </PlaygroundLayoutSidebar>

                <Drawer open={drawerOpen === playgroundCategory} onOpenChange={handleDrawerOpenChange}>
                  <DrawerContent>
                    <div className="p-4 overflow-auto">{sidePanel}</div>
                  </DrawerContent>
                </Drawer>
                <PlaygroundLayoutContent className="[&>*]:w-full [&>*]:max-w-[min(90%,1024px)]">
                  {playgroundCategory === PlaygroundCategories.SANDBOX && <SandboxCodeSnippetsResponse />}
                  {playgroundCategory === PlaygroundCategories.TERMINAL && <WebTerminal />}
                  {playgroundCategory === PlaygroundCategories.VNC && <VNCDesktopWindowResponse />}
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
