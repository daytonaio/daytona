/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { AlertTriangle, ArrowRight, Box, Container, HardDrive, LockKeyhole, Server, Users } from 'lucide-react'
import { ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'

const DeletePrerequisiteScope = {
  USERS: 'USERS',
  RUNNERS: 'RUNNERS',
  SANDBOXES: 'SANDBOXES',
  SNAPSHOTS: 'SNAPSHOTS',
  VOLUMES: 'VOLUMES',
  SUSPENDED: 'SUSPENDED',
  ORGANIZATIONS: 'ORGANIZATIONS',
  UNKNOWN: 'UNKNOWN',
} as const

type DeletePrerequisiteScopeType = (typeof DeletePrerequisiteScope)[keyof typeof DeletePrerequisiteScope]

export interface DeletePrerequisite {
  id: string
  scope: DeletePrerequisiteScopeType
  type: 'count' | 'error' | 'critical'
  icon: ReactNode
  message: string
  action?: ReactNode
}

const LinkAction = ({ to, label }: { to: string; label: string }) => {
  const navigate = useNavigate()
  return (
    <Button variant="outline" size="sm" className="shrink-0" onClick={() => navigate(to)}>
      {label}
      <ArrowRight size={12} strokeWidth={1.5} />
    </Button>
  )
}

const SupportAction = ({ label }: { label: string }) => {
  const handleClick = () => {
    const pylon = (window as any).Pylon
    if (typeof pylon === 'function') pylon('show')
  }
  return (
    <Button variant="outline" size="sm" className="shrink-0" onClick={handleClick}>
      {label}
      <ArrowRight size={12} strokeWidth={1.5} />
    </Button>
  )
}

const PrereqIcon = ({ icon, type }: { icon: ReactNode; type: DeletePrerequisite['type'] }) => (
  <div
    className={cn('flex h-9 w-9 shrink-0 items-center justify-center rounded-md', {
      'bg-destructive-background text-destructive-foreground': type === 'critical',
      'bg-warning-background text-warning-foreground': type === 'error',
      'bg-muted text-muted-foreground': type === 'count',
    })}
  >
    {icon}
  </div>
)

interface ScopeConfig {
  icon: ReactNode
  action?: ReactNode
}

const SCOPE_CONFIG: Record<DeletePrerequisiteScopeType, ScopeConfig> = {
  [DeletePrerequisiteScope.USERS]: {
    icon: <Users size={16} strokeWidth={1.5} />,
    action: <LinkAction to="/dashboard/members" label="Manage Members" />,
  },
  [DeletePrerequisiteScope.RUNNERS]: {
    icon: <Server size={16} strokeWidth={1.5} />,
    action: <LinkAction to="/dashboard/runners" label="Check Runners" />,
  },
  [DeletePrerequisiteScope.SANDBOXES]: {
    icon: <Container size={16} strokeWidth={1.5} />,
    action: <LinkAction to="/dashboard/sandboxes" label="View Sandboxes" />,
  },
  [DeletePrerequisiteScope.SNAPSHOTS]: {
    icon: <Box size={16} strokeWidth={1.5} />,
    action: <LinkAction to="/dashboard/snapshots" label="View Snapshots" />,
  },
  [DeletePrerequisiteScope.VOLUMES]: {
    icon: <HardDrive size={16} strokeWidth={1.5} />,
    action: <LinkAction to="/dashboard/volumes" label="View Volumes" />,
  },
  [DeletePrerequisiteScope.SUSPENDED]: {
    icon: <LockKeyhole size={16} strokeWidth={1.5} />,
    action: <SupportAction label="Contact Support" />,
  },
  [DeletePrerequisiteScope.ORGANIZATIONS]: { icon: <Users size={16} strokeWidth={1.5} /> },
  [DeletePrerequisiteScope.UNKNOWN]: {
    icon: <AlertTriangle size={16} strokeWidth={1.5} />,
    action: <LinkAction to="#" label="View Details" />,
  },
}

export const parseDeleteErrors = (errors: string[]): DeletePrerequisite[] => {
  if (!Array.isArray(errors)) return []

  return errors.map((msg, index) => {
    const lowerMsg = msg.toLowerCase()
    let scope: DeletePrerequisiteScopeType = DeletePrerequisiteScope.UNKNOWN

    if (lowerMsg.includes('suspended')) scope = DeletePrerequisiteScope.SUSPENDED
    else if (lowerMsg.includes('organization') && lowerMsg.includes('member'))
      scope = DeletePrerequisiteScope.ORGANIZATIONS
    else if (lowerMsg.includes('user')) scope = DeletePrerequisiteScope.USERS
    else if (lowerMsg.includes('runner')) scope = DeletePrerequisiteScope.RUNNERS
    else if (lowerMsg.includes('sandbox')) scope = DeletePrerequisiteScope.SANDBOXES
    else if (lowerMsg.includes('snapshot')) scope = DeletePrerequisiteScope.SNAPSHOTS
    else if (lowerMsg.includes('volume')) scope = DeletePrerequisiteScope.VOLUMES
    else if (lowerMsg.includes('organization')) scope = DeletePrerequisiteScope.ORGANIZATIONS

    let type: 'count' | 'error' | 'critical' = 'count'
    if (lowerMsg.includes('failed to check')) type = 'error'
    if (scope === DeletePrerequisiteScope.SUSPENDED || scope === DeletePrerequisiteScope.ORGANIZATIONS)
      type = 'critical'

    const config = SCOPE_CONFIG[scope]

    return {
      id: `blocker-${scope}-${index}`,
      scope,
      type,
      icon: config.icon,
      message: msg,
      action: config.action,
    }
  })
}

export const DeletePrerequisiteItem = ({ prereq }: { prereq: DeletePrerequisite }) => {
  return (
    <div className="flex items-center gap-3 px-6 py-4">
      <PrereqIcon icon={prereq.icon} type={prereq.type} />
      <p className="flex-1 text-sm">{prereq.message}</p>
      {prereq.action}
    </div>
  )
}
