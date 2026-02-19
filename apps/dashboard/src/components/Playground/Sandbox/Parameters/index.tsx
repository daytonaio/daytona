/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion'
import { Switch } from '@/components/ui/switch'
import { SandboxParametersSections } from '@/enums/Playground'
import { usePlayground } from '@/hooks/usePlayground'
import { cn } from '@/lib/utils'
import { BoltIcon, FolderIcon, GitBranchIcon, SquareTerminalIcon } from 'lucide-react'
import SandboxFileSystem from './FileSystem'
import SandboxGitOperations from './GitOperations'
import SandboxManagementParameters from './Management'
import SandboxProcessCodeExecution from './ProcessCodeExecution'

const sandboxParametersSectionsData = [
  { value: SandboxParametersSections.SANDBOX_MANAGEMENT, label: 'Management' },
  { value: SandboxParametersSections.FILE_SYSTEM, label: 'File System' },
  { value: SandboxParametersSections.GIT_OPERATIONS, label: 'Git Operations' },
  { value: SandboxParametersSections.PROCESS_CODE_EXECUTION, label: 'Process & Code Execution' },
]

const sectionIcons = {
  [SandboxParametersSections.SANDBOX_MANAGEMENT]: <BoltIcon strokeWidth={1.5} />,
  [SandboxParametersSections.GIT_OPERATIONS]: <GitBranchIcon strokeWidth={1.5} />,
  [SandboxParametersSections.FILE_SYSTEM]: <FolderIcon strokeWidth={1.5} />,
  [SandboxParametersSections.PROCESS_CODE_EXECUTION]: <SquareTerminalIcon strokeWidth={1.5} />,
}

const SandboxParameters = ({ className }: { className?: string }) => {
  const { openedParametersSections, setOpenedParametersSections, enabledSections, enableSection, disableSection } =
    usePlayground()

  // TODO - Currently, snapshot selection is not supported in the Playground, so we are using empty array and false for loading. We keep to code commented to enable it in future if requested by users.
  // const { snapshotApi } = useApi()
  // const { selectedOrganization } = useSelectedOrganization()

  // const { data: snapshotsData = [], isLoading: snapshotsLoading } = useQuery({
  //   queryKey: ['snapshots', selectedOrganization?.id, 'all'],
  //   queryFn: async () => {
  //     if (!selectedOrganization) return []
  //     const response = await snapshotApi.getAllSnapshots(selectedOrganization.id)
  //     return response.data.items
  //   },
  //   enabled: !!selectedOrganization,
  // })

  return (
    <div className={cn('flex flex-col gap-6', className)}>
      <div>
        <h2>Sandbox Configuration</h2>
        <p className="text-sm text-muted-foreground mt-1">Manage resources, lifecycle policies, and file systems.</p>
      </div>
      <Accordion
        type="multiple"
        value={openedParametersSections}
        onValueChange={(sections) => setOpenedParametersSections(sections as SandboxParametersSections[])}
      >
        {sandboxParametersSectionsData.map((section) => {
          const isManagement = section.value === SandboxParametersSections.SANDBOX_MANAGEMENT
          const isEnabled = enabledSections.includes(section.value as SandboxParametersSections)
          const isExpanded = openedParametersSections.includes(section.value as SandboxParametersSections)
          return (
            <AccordionItem
              key={section.value}
              value={section.value}
              className="border px-2 last:border-b first:rounded-t-lg last:rounded-b-lg border-t-0 first:border-t"
            >
              <AccordionTrigger
                headerClassName={cn(
                  'font-semibold text-muted-foreground dark:bg-muted/50 bg-muted/80 border-b -mx-2 px-3',
                  {
                    'border-b-border': isExpanded,
                    'border-b-transparent': !isExpanded,
                    'opacity-80': !isEnabled && !isManagement,
                  },
                )}
                className="hover:no-underline hover:text-primary py-3"
                right={
                  !isManagement ? (
                    <Switch
                      checked={isEnabled}
                      onCheckedChange={(checked) =>
                        checked
                          ? enableSection(section.value as SandboxParametersSections)
                          : disableSection(section.value as SandboxParametersSections)
                      }
                      size="sm"
                      className="ml-3"
                    />
                  ) : undefined
                }
              >
                <div className="flex items-center gap-2 [&_svg]:size-4 text-sm font-medium">
                  {sectionIcons[section.value]} {section.label}
                </div>
              </AccordionTrigger>
              <AccordionContent className="py-3 px-1">
                {isExpanded && (
                  <div className="space-y-4">
                    {section.value === SandboxParametersSections.FILE_SYSTEM && <SandboxFileSystem />}
                    {section.value === SandboxParametersSections.GIT_OPERATIONS && <SandboxGitOperations />}
                    {section.value === SandboxParametersSections.SANDBOX_MANAGEMENT && (
                      <SandboxManagementParameters snapshotsData={[]} snapshotsLoading={false} />
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
