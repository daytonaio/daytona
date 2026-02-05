/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion'
import { SandboxParametersSections, sandboxParametersSectionsData } from '@/enums/Playground'
import { useApi } from '@/hooks/useApi'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { handleApiError } from '@/lib/error-handling'
import { cn } from '@/lib/utils'
import { SnapshotDto } from '@daytonaio/api-client'
import { BoltIcon, FolderIcon, GitBranchIcon, SquareTerminalIcon } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import SandboxFileSystem from './FileSystem'
import SandboxGitOperations from './GitOperations'
import SandboxManagementParameters from './Management'
import SandboxProcessCodeExecution from './ProcessCodeExecution'

const sectionIcons = {
  [SandboxParametersSections.SANDBOX_MANAGEMENT]: <BoltIcon strokeWidth={1.5} />,
  [SandboxParametersSections.GIT_OPERATIONS]: <GitBranchIcon strokeWidth={1.5} />,
  [SandboxParametersSections.FILE_SYSTEM]: <FolderIcon strokeWidth={1.5} />,
  [SandboxParametersSections.PROCESS_CODE_EXECUTION]: <SquareTerminalIcon strokeWidth={1.5} />,
}

const SandboxParameters = ({ className }: { className?: string }) => {
  const [openedParametersSections, setOpenedParametersSections] = useState<SandboxParametersSections[]>([
    SandboxParametersSections.SANDBOX_MANAGEMENT,
  ])

  const [snapshotsData, setSnapshotsData] = useState<Array<SnapshotDto>>([])
  const [snapshotsLoading, setSnapshotsLoading] = useState<boolean>(true)

  const { snapshotApi } = useApi()
  const { selectedOrganization } = useSelectedOrganization()

  const fetchSnapshots = useCallback(async () => {
    if (!selectedOrganization) return
    setSnapshotsLoading(true)
    try {
      const response = (await snapshotApi.getAllSnapshots(selectedOrganization.id)).data
      setSnapshotsData(response.items)
    } catch (error) {
      handleApiError(error, 'Failed to fetch snapshots')
    } finally {
      setSnapshotsLoading(false)
    }
  }, [snapshotApi, selectedOrganization])

  useEffect(() => {
    fetchSnapshots()
  }, [fetchSnapshots])

  return (
    <div className={cn('flex flex-col gap-6', className)}>
      <div>
        <h2>Sandbox Configuration</h2>
        <p className="text-sm text-muted-foreground mt-1">Manage resources, lifecycle policies, and file systems.</p>
      </div>
      <Accordion
        type="multiple"
        value={openedParametersSections}
        onValueChange={(parametersSections) =>
          setOpenedParametersSections(parametersSections as SandboxParametersSections[])
        }
      >
        {sandboxParametersSectionsData.map((section) => {
          const isCollapsed = !openedParametersSections.includes(section.value as SandboxParametersSections)
          return (
            <AccordionItem
              key={section.value}
              value={section.value}
              className="border px-2 last:border-b first:rounded-t-lg last:rounded-b-lg border-t-0 first:border-t"
            >
              <AccordionTrigger className="font-semibold text-muted-foreground hover:no-underline dark:bg-muted/50 bg-muted/80 hover:text-primary py-3 border-b border-b-transparent data-[state=open]:border-b-border -mx-2 px-3">
                <div className="flex items-center gap-2 [&_svg]:size-4 text-sm font-medium">
                  {sectionIcons[section.value]} {section.label}
                </div>
              </AccordionTrigger>
              <AccordionContent className="py-3 px-1">
                {!isCollapsed && (
                  <div className="space-y-4">
                    {section.value === SandboxParametersSections.FILE_SYSTEM && <SandboxFileSystem />}
                    {section.value === SandboxParametersSections.GIT_OPERATIONS && <SandboxGitOperations />}
                    {section.value === SandboxParametersSections.SANDBOX_MANAGEMENT && (
                      <SandboxManagementParameters snapshotsData={snapshotsData} snapshotsLoading={snapshotsLoading} />
                    )}
                    {section.value === SandboxParametersSections.PROCESS_CODE_EXECUTION && (
                      <SandboxProcessCodeExecution />
                    )}
                  </div>
                )}
              </AccordionContent>
            </AccordionItem>
          )
        })}
      </Accordion>
    </div>
  )
}

export default SandboxParameters
