/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Logo, LogoText } from '@/assets/Logo'
import { OrganizationPicker } from '@/components/Organizations/OrganizationPicker'
import {
  Sidebar as SidebarComponent,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarSeparator,
  SidebarTrigger,
  useSidebar,
} from '@/components/ui/sidebar'
import { DAYTONA_DOCS_URL, DAYTONA_SLACK_URL } from '@/constants/ExternalLinks'
import { useTheme } from '@/contexts/ThemeContext'
import { RoutePath } from '@/enums/RoutePath'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useUserOrganizationInvitations } from '@/hooks/useUserOrganizationInvitations'
import { useWebhooks } from '@/hooks/useWebhooks'
import { cn } from '@/lib/utils'
import { OrganizationRolePermissionsEnum, OrganizationUserRoleEnum } from '@daytonaio/api-client'
import { addHours, formatRelative } from 'date-fns'
import {
  BookOpen,
  Box,
  ChartColumn,
  ChevronsUpDown,
  Container,
  CreditCard,
  HardDrive,
  KeyRound,
  ListChecks,
  LockKeyhole,
  LogOut,
  Mail,
  Moon,
  PackageOpen,
  Settings,
  Slack,
  SquareUserRound,
  Sun,
  TextSearch,
  TriangleAlert,
  Users,
} from 'lucide-react'
import { useMemo } from 'react'
import { useAuth } from 'react-oidc-context'
import { Link, useLocation } from 'react-router-dom'
import { Button } from './ui/button'
import { Card, CardHeader, CardTitle } from './ui/card'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from './ui/dropdown-menu'
import { ScrollArea } from './ui/scroll-area'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './ui/tooltip'
import { useConfig } from '@/hooks/useConfig'
interface SidebarProps {
  isBannerVisible: boolean
  billingEnabled: boolean
  version: string
}

interface SidebarItem {
  icon: React.ReactElement
  label: string
  path: RoutePath | string
  onClick?: () => void
}

export function Sidebar({ isBannerVisible, billingEnabled, version }: SidebarProps) {
  const config = useConfig()
  const { theme, setTheme } = useTheme()
  const { user, signoutRedirect } = useAuth()
  const { pathname } = useLocation()
  const sidebar = useSidebar()
  const { selectedOrganization, authenticatedUserOrganizationMember, authenticatedUserHasPermission } =
    useSelectedOrganization()
  const { count: organizationInvitationsCount } = useUserOrganizationInvitations()
  const { isInitialized: webhooksInitialized, openAppPortal } = useWebhooks()
  const sidebarItems = useMemo(() => {
    const arr: SidebarItem[] = [
      {
        icon: <Container size={16} strokeWidth={1.5} />,
        label: 'Sandboxes',
        path: RoutePath.SANDBOXES,
      },
      {
        icon: <Box size={16} strokeWidth={1.5} />,
        label: 'Snapshots',
        path: RoutePath.SNAPSHOTS,
      },
      {
        icon: <PackageOpen size={16} strokeWidth={1.5} />,
        label: 'Registries',
        path: RoutePath.REGISTRIES,
      },
    ]
    if (authenticatedUserHasPermission(OrganizationRolePermissionsEnum.READ_VOLUMES)) {
      arr.push({
        icon: <HardDrive size={16} strokeWidth={1.5} />,
        label: 'Volumes',
        path: RoutePath.VOLUMES,
      })
    }

    if (authenticatedUserHasPermission(OrganizationRolePermissionsEnum.READ_AUDIT_LOGS)) {
      arr.push({
        icon: <TextSearch size={16} strokeWidth={1.5} />,
        label: 'Audit Logs',
        path: RoutePath.AUDIT_LOGS,
      })
    }

    return arr
  }, [authenticatedUserHasPermission])

  const settingsItems = useMemo(() => {
    const arr: SidebarItem[] = [
      {
        icon: <Settings size={16} strokeWidth={1.5} />,
        label: 'Settings',
        path: RoutePath.SETTINGS,
      },
      { icon: <KeyRound size={16} strokeWidth={1.5} />, label: 'API Keys', path: RoutePath.KEYS },
    ]

    // Add Webhooks link if webhooks are initialized
    if (webhooksInitialized) {
      arr.push({
        icon: <Mail size={16} strokeWidth={1.5} />,
        label: 'Webhooks',
        path: '#webhooks' as any, // This will be handled by onClick
        onClick: () => openAppPortal(),
      })
    }

    if (authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER) {
      arr.push({
        icon: <LockKeyhole size={16} strokeWidth={1.5} />,
        label: 'Limits',
        path: RoutePath.LIMITS,
      })
    }
    if (!selectedOrganization?.personal) {
      arr.push({
        icon: <Users size={16} strokeWidth={1.5} />,
        label: 'Members',
        path: RoutePath.MEMBERS,
      })
      // TODO: uncomment when we allow creating custom roles
      // if (authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER) {
      //   arr.push({ icon: <UserCog className="w-5 h-5" />, label: 'Roles', path: RoutePath.ROLES })
      // }
    }

    return arr
  }, [authenticatedUserOrganizationMember?.role, selectedOrganization?.personal, webhooksInitialized, openAppPortal])

  const billingItems = useMemo(() => {
    if (!billingEnabled || authenticatedUserOrganizationMember?.role !== OrganizationUserRoleEnum.OWNER) {
      return []
    }

    return [
      {
        icon: <ChartColumn size={16} strokeWidth={1.5} />,
        label: 'Spending',
        path: RoutePath.BILLING_SPENDING,
      },
      {
        icon: <CreditCard size={16} strokeWidth={1.5} />,
        label: 'Wallet',
        path: RoutePath.BILLING_WALLET,
      },
    ]
  }, [billingEnabled, authenticatedUserOrganizationMember?.role])

  const handleSignOut = () => {
    signoutRedirect()
  }

  return (
    <SidebarComponent isBannerVisible={isBannerVisible} collapsible="icon">
      <SidebarHeader>
        <div
          className={cn('flex justify-between items-center gap-2 px-2 mb-2 h-12', {
            'justify-center px-0': !sidebar.open,
          })}
        >
          <div className="flex items-center gap-2 group-data-[state=collapsed]:hidden text-primary">
            <Logo />
            <LogoText />
          </div>
          <SidebarTrigger className="p-2 [&_svg]:size-5" />
        </div>
        <SidebarMenu>
          <OrganizationPicker />
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <ScrollArea fade="shadow" className="overflow-auto flex-1">
          <SidebarGroup>
            <SidebarGroupContent>
              <SidebarMenu>
                {sidebarItems.map((item) => (
                  <SidebarMenuItem key={item.label}>
                    <SidebarMenuButton
                      asChild
                      isActive={pathname.startsWith(item.path)}
                      className="text-sm"
                      tooltip={item.label}
                    >
                      {item.onClick ? (
                        <button onClick={() => item.onClick?.()}>
                          {item.icon}
                          <span>{item.label}</span>
                        </button>
                      ) : (
                        <Link to={item.path}>
                          {item.icon}
                          <span>{item.label}</span>
                        </Link>
                      )}
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
          <SidebarSeparator />

          <SidebarGroup>
            <SidebarGroupContent>
              <SidebarMenu>
                {settingsItems.map((item) => (
                  <SidebarMenuItem key={item.label}>
                    <SidebarMenuButton
                      asChild
                      isActive={pathname.startsWith(item.path)}
                      className="text-sm"
                      tooltip={item.label}
                    >
                      <Link to={item.path}>
                        {item.icon}
                        <span>{item.label}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
          <SidebarSeparator />
          <SidebarGroup>
            <SidebarGroupContent>
              <SidebarMenu>
                {billingItems.map((item) => (
                  <SidebarMenuItem key={item.label}>
                    <SidebarMenuButton
                      asChild
                      isActive={pathname.startsWith(item.path)}
                      className="text-sm"
                      tooltip={item.label}
                    >
                      <Link to={item.path}>
                        {item.icon}
                        <span>{item.label}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </ScrollArea>
      </SidebarContent>
      <SidebarFooter className="pb-4">
        <SidebarMenu>
          {selectedOrganization?.suspended && (
            <SidebarMenuItem key="suspended">
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Card className="w-full m-0 p-0 mb-4 cursor-pointer border-red-600 bg-red-100/80 text-red-800 dark:border-red-800 dark:bg-red-900/20 dark:text-red-300">
                      <CardHeader className="py-2 pl-2">
                        <CardTitle className="text-sm flex items-center gap-2">
                          <TriangleAlert className="w-4 h-4 flex-shrink-0" />
                          <div className="overflow-hidden">
                            Organization suspended
                            {selectedOrganization.suspensionReason && (
                              <div className="text-xs text-muted-foreground text-ellipsis overflow-hidden">
                                ({selectedOrganization.suspensionReason})
                              </div>
                            )}
                          </div>
                        </CardTitle>
                      </CardHeader>
                    </Card>
                  </TooltipTrigger>
                  <TooltipContent className="mb-2 flex flex-col gap-2 max-w-[400px]" side="right">
                    <p>
                      <strong>Organization suspended</strong>
                      <br />
                      Starting and creating sandboxes is disabled.
                    </p>
                    {selectedOrganization.suspensionReason && (
                      <p>
                        <strong>Suspension reason:</strong> <br />
                        {selectedOrganization.suspensionReason}
                      </p>
                    )}
                    {selectedOrganization.suspendedAt && (
                      <p>
                        <strong>Suspended at:</strong> <br />
                        {new Date(selectedOrganization.suspendedAt).toLocaleString('en-US', {
                          timeZone: 'UTC',
                          year: 'numeric',
                          month: 'long',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit',
                          second: '2-digit',
                        })}
                        <br />
                        Started sandboxes will be stopped{' '}
                        {formatRelative(
                          addHours(
                            new Date(selectedOrganization.suspendedAt),
                            selectedOrganization.suspensionCleanupGracePeriodHours,
                          ),
                          new Date(selectedOrganization.suspendedAt),
                        )}
                      </p>
                    )}
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </SidebarMenuItem>
          )}
          <SidebarMenuItem key="theme-toggle">
            <SidebarMenuButton
              onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
              className="h-8 py-0"
              title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
              tooltip="Toggle theme"
            >
              {theme === 'dark' ? <Sun size={16} /> : <Moon size={16} />}
              <span>{theme === 'dark' ? 'Light mode' : 'Dark mode'}</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem key="slack">
            <SidebarMenuButton asChild tooltip="Slack">
              <a href={DAYTONA_SLACK_URL} className=" h-8 py-0" target="_blank" rel="noopener noreferrer">
                <Slack size={16} strokeWidth={1.5} />
                <span>Slack</span>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem key="docs">
            <SidebarMenuButton asChild tooltip="Docs">
              <a href={DAYTONA_DOCS_URL} className=" h-8 py-0" target="_blank" rel="noopener noreferrer">
                <BookOpen size={16} strokeWidth={1.5} />
                <span>Docs</span>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton
                  className={cn(
                    'flex flex-shrink-0 items-center outline outline-1 outline-border outline-offset-0 bg-muted font-medium mt-2',
                    {
                      'h-12': sidebar.open,
                    },
                  )}
                  tooltip="Profile"
                >
                  {user?.profile.picture ? (
                    <img
                      src={user.profile.picture}
                      alt={user.profile.name || 'Profile picture'}
                      className="h-4 w-4 rounded-sm flex-shrink-0"
                    />
                  ) : (
                    <SquareUserRound className="!w-4 !h-4  flex-shrink-0" />
                  )}
                  <div className="flex-1 min-w-0">
                    <span className="truncate block">{user?.profile.name || ''}</span>
                    <span className="truncate block text-muted-foreground text-xs">{user?.profile.email || ''}</span>
                  </div>
                  <ChevronsUpDown className="w-4 h-4 opacity-50 flex-shrink-0" />
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent side="top" align="start" className="w-[--radix-popper-anchor-width] min-w-[12rem]">
                {config.linkedAccountsEnabled && (
                  <DropdownMenuItem asChild>
                    <Button variant="ghost" className="w-full cursor-pointer justify-start" asChild>
                      <Link to={RoutePath.ACCOUNT_SETTINGS}>
                        <Settings className="w-4 h-4" />
                        Account Settings
                      </Link>
                    </Button>
                  </DropdownMenuItem>
                )}
                <DropdownMenuItem asChild>
                  <Button variant="ghost" className="w-full cursor-pointer justify-start" asChild>
                    <Link to={RoutePath.USER_INVITATIONS}>
                      <Mail className="w-4 h-4" />
                      Invitations
                      {organizationInvitationsCount > 0 && (
                        <span className="ml-auto px-2 py-0.5 text-xs font-medium bg-secondary rounded-full">
                          {organizationInvitationsCount}
                        </span>
                      )}
                    </Link>
                  </Button>
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                  <Button variant="ghost" className="w-full cursor-pointer justify-start" asChild>
                    <Link to={RoutePath.ONBOARDING}>
                      <ListChecks className="w-4 h-4" />
                      Onboarding
                    </Link>
                  </Button>
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                  <Button variant="ghost" className="w-full cursor-pointer justify-start" onClick={handleSignOut}>
                    <LogOut className="w-4 h-4" />
                    Sign out
                  </Button>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
          <SidebarMenuItem key="version">
            <div
              className={cn(
                'flex items-center w-full justify-center gap-2 mt-2 overflow-auto min-h-4 whitespace-nowrap',
              )}
            >
              {sidebar.open && <span className="text-xs text-muted-foreground">Version {version}</span>}
            </div>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </SidebarComponent>
  )
}
