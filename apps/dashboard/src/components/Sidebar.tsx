/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  BookOpen,
  Box,
  ChartColumn,
  ChevronsUpDown,
  Container,
  CreditCard,
  KeyRound,
  ListChecks,
  LogOut,
  Mail,
  Moon,
  PackageOpen,
  Settings,
  Slack,
  SquareUserRound,
  Sun,
  TriangleAlert,
  //UserCog,
  Users,
} from 'lucide-react'

import {
  Sidebar as SidebarComponent,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar'
import daytonaLogoBlack from '../assets/daytona-full-black.png'
import daytonaLogoWhite from '../assets/daytona-full-white.png'
import { useTheme } from '@/contexts/ThemeContext'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from './ui/dropdown-menu'
import { useAuth } from 'react-oidc-context'
import { Button } from './ui/button'
import { useNavigate } from 'react-router-dom'
import { useMemo } from 'react'
import { OrganizationPicker } from '@/components/Organizations/OrganizationPicker'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useUserOrganizationInvitations } from '@/hooks/useUserOrganizationInvitations'
import { OrganizationUserRoleEnum } from '@daytonaio/api-client'
import { Card, CardDescription, CardHeader, CardTitle } from './ui/card'
import { Tooltip, TooltipContent } from './ui/tooltip'
import { TooltipProvider, TooltipTrigger } from './ui/tooltip'
import { addHours, formatRelative } from 'date-fns'
import { RoutePath } from '@/enums/RoutePath'
import { DAYTONA_DOCS_URL, DAYTONA_SLACK_URL } from '@/constants/ExternalLinks'

export function Sidebar() {
  const { theme, setTheme } = useTheme()
  const { user, signoutRedirect } = useAuth()
  const navigate = useNavigate()

  const { selectedOrganization, authenticatedUserOrganizationMember } = useSelectedOrganization()
  const { count: organizationInvitationsCount } = useUserOrganizationInvitations()

  const sidebarItems = useMemo(() => {
    const arr = [
      { icon: <Container className="w-5 h-5" />, label: 'Sandboxes', path: RoutePath.SANDBOXES },
      { icon: <KeyRound className="w-5 h-5" />, label: 'Keys', path: RoutePath.KEYS },
      {
        icon: <Box className="w-5 h-5" />,
        label: 'Images',
        path: RoutePath.IMAGES,
      },
      { icon: <PackageOpen className="w-5 h-5" />, label: 'Registries', path: RoutePath.REGISTRIES },
      { icon: <ChartColumn className="w-5 h-5" />, label: 'Usage', path: RoutePath.USAGE },
    ]

    if (
      import.meta.env.VITE_BILLING_API_URL &&
      authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER
    ) {
      arr.push({ icon: <CreditCard className="w-5 h-5" />, label: 'Billing', path: RoutePath.BILLING })
    }

    if (!selectedOrganization?.personal) {
      arr.push({ icon: <Users className="w-5 h-5" />, label: 'Members', path: RoutePath.MEMBERS })

      // TODO: uncomment when we allow creating custom roles
      // if (authenticatedUserOrganizationMember?.role === OrganizationUserRoleEnum.OWNER) {
      //   arr.push({ icon: <UserCog className="w-5 h-5" />, label: 'Roles', path: RoutePath.ROLES })
      // }
    }
    arr.push({ icon: <Settings className="w-5 h-5" />, label: 'Settings', path: RoutePath.SETTINGS })
    return arr
  }, [authenticatedUserOrganizationMember?.role, selectedOrganization?.personal])

  const handleSignOut = () => {
    signoutRedirect()
  }

  return (
    <SidebarComponent>
      <SidebarContent>
        <SidebarGroup>
          <div className="p-2 mb-2">
            <img
              src={theme === 'dark' ? daytonaLogoWhite : daytonaLogoBlack}
              alt="Daytona Logo"
              className="h-8 w-auto"
            />
          </div>

          <OrganizationPicker />

          <SidebarGroupContent>
            <SidebarMenu>
              {sidebarItems.map((item) => (
                <SidebarMenuItem key={item.label}>
                  <SidebarMenuButton asChild>
                    <button onClick={() => navigate(item.path)} className="text-sm">
                      {item.icon}
                      <span>{item.label}</span>
                    </button>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
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
                          addHours(new Date(selectedOrganization.suspendedAt), 24),
                          new Date(selectedOrganization.suspendedAt),
                        )}
                      </p>
                    )}
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </SidebarMenuItem>
          )}
          <SidebarMenuItem key="slack">
            <SidebarMenuButton asChild>
              <a href={DAYTONA_SLACK_URL} className="text-xs h-8 py-0" target="_blank" rel="noopener noreferrer">
                <Slack className="!w-4 !h-4 fill-foreground" />
                <span>Slack</span>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem key="docs">
            <SidebarMenuButton asChild>
              <a href={DAYTONA_DOCS_URL} className="text-xs h-8 py-0" target="_blank" rel="noopener noreferrer">
                <BookOpen className="!w-4 !h-4" />
                <span>Docs</span>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton className="h-14 pl-1 flex items-center">
                  {user?.profile.picture ? (
                    <img
                      src={user.profile.picture}
                      alt={user.profile.name || 'Profile picture'}
                      className="h-8 w-8 rounded-sm flex-shrink-0"
                    />
                  ) : (
                    <SquareUserRound className="!w-8 !h-8 flex-shrink-0" />
                  )}
                  <div className="flex-1 min-w-0">
                    <span className="truncate block">{user?.profile.name || ''}</span>
                    <span className="truncate block text-muted-foreground">{user?.profile.email || ''}</span>
                  </div>
                  <ChevronsUpDown className="w-4 h-4 opacity-50 flex-shrink-0" />
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent side="top" className="w-[--radix-popper-anchor-width]">
                <DropdownMenuItem asChild>
                  <Button
                    variant="ghost"
                    className="w-full cursor-pointer justify-start"
                    onClick={() => navigate(RoutePath.USER_INVITATIONS)}
                  >
                    <Mail className="w-4 h-4" />
                    Invitations
                    {organizationInvitationsCount > 0 && (
                      <span className="ml-auto px-2 py-0.5 text-xs font-medium bg-secondary rounded-full">
                        {organizationInvitationsCount}
                      </span>
                    )}
                  </Button>
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                  <Button
                    variant="ghost"
                    className="w-full cursor-pointer justify-start"
                    onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
                  >
                    {theme === 'dark' ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
                    {theme === 'dark' ? 'Light mode' : 'Dark mode'}
                  </Button>
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                  <Button
                    variant="ghost"
                    className="w-full cursor-pointer justify-start"
                    onClick={() => navigate(RoutePath.ONBOARDING)}
                  >
                    <ListChecks className="w-4 h-4" />
                    Onboarding
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
        </SidebarMenu>
      </SidebarFooter>
    </SidebarComponent>
  )
}
