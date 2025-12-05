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
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarTrigger,
} from '@/components/ui/sidebar'
import { DAYTONA_DOCS_URL, DAYTONA_SLACK_URL } from '@/constants/ExternalLinks'
import { useTheme } from '@/contexts/ThemeContext'
import { RoutePath } from '@/enums/RoutePath'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useUserOrganizationInvitations } from '@/hooks/useUserOrganizationInvitations'
import { useWebhooks } from '@/hooks/useWebhooks'
import { OrganizationRolePermissionsEnum, OrganizationUserRoleEnum } from '@daytonaio/api-client'
import { addHours, formatRelative } from 'date-fns'
import {
  ArrowRightIcon,
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
  TriangleAlertIcon,
  Users,
} from 'lucide-react'
import { useMemo } from 'react'
import { useAuth } from 'react-oidc-context'
import { Link, useLocation } from 'react-router-dom'
import { Alert, AlertDescription, AlertTitle } from './ui/alert'
import { Button } from './ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from './ui/dropdown-menu'
interface SidebarProps {
  isBannerVisible: boolean
  billingEnabled: boolean
  version: string
}

export function Sidebar({ isBannerVisible, billingEnabled, version }: SidebarProps) {
  const { theme, setTheme } = useTheme()
  const { user, signoutRedirect } = useAuth()
  const { pathname } = useLocation()
  const { selectedOrganization, authenticatedUserOrganizationMember, authenticatedUserHasPermission } =
    useSelectedOrganization()
  const { count: organizationInvitationsCount } = useUserOrganizationInvitations()
  const { isInitialized: webhooksInitialized, openAppPortal } = useWebhooks()
  const sidebarItems = useMemo(() => {
    const arr: Array<{
      icon: React.ReactElement
      label: string
      path: RoutePath | string
      onClick?: () => void
    }> = [
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

    // Add Webhooks link if webhooks are initialized
    if (webhooksInitialized) {
      arr.push({
        icon: <Mail size={16} strokeWidth={1.5} />,
        label: 'Webhooks',
        path: '#webhooks' as any, // This will be handled by onClick
        onClick: () => openAppPortal(),
      })
    }

    return arr
  }, [authenticatedUserHasPermission, webhooksInitialized, openAppPortal])

  const settingsItems = useMemo(() => {
    const arr = [
      {
        icon: <Settings size={16} strokeWidth={1.5} />,
        label: 'General',
        path: RoutePath.SETTINGS,
      },
      { icon: <KeyRound size={16} strokeWidth={1.5} />, label: 'API Keys', path: RoutePath.KEYS },
    ]

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
  }, [authenticatedUserOrganizationMember?.role, selectedOrganization?.personal])

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
      <SidebarContent>
        <SidebarGroup>
          <div className="flex justify-between items-center gap-2 px-2 mb-2 h-12">
            <div className="flex items-center gap-2 group-data-[state=collapsed]:hidden text-primary">
              <Logo />
              <LogoText />
            </div>
            <SidebarTrigger className="p-2 [&_svg]:size-5" />
          </div>
          <OrganizationPicker />
          <SidebarGroupContent>
            <SidebarMenu>
              {sidebarItems.map((item) => (
                <SidebarMenuItem key={item.label}>
                  <SidebarMenuButton asChild isActive={pathname.startsWith(item.path)} className="text-sm">
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
        <SidebarGroup>
          <SidebarGroupLabel>Settings</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {settingsItems.map((item) => (
                <SidebarMenuItem key={item.label}>
                  <SidebarMenuButton asChild isActive={pathname.startsWith(item.path)} className="text-sm">
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
        <SidebarGroup>
          {billingItems.length > 0 && <SidebarGroupLabel>Billing</SidebarGroupLabel>}
          <SidebarGroupContent>
            <SidebarMenu>
              {billingItems.map((item) => (
                <SidebarMenuItem key={item.label}>
                  <SidebarMenuButton asChild isActive={pathname.startsWith(item.path)} className="text-sm">
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
      </SidebarContent>
      <SidebarFooter>
        <SidebarMenu>
          {selectedOrganization?.suspended && (
            <SidebarMenuItem key="suspended">
              <SuspensionBanner suspension={selectedOrganization} />
            </SidebarMenuItem>
          )}
          <SidebarMenuItem key="theme-toggle">
            <SidebarMenuButton
              onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
              className="h-8 py-0"
              title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
            >
              {theme === 'dark' ? <Sun size={16} /> : <Moon size={16} />}
              <span>{theme === 'dark' ? 'Light mode' : 'Dark mode'}</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem key="slack">
            <SidebarMenuButton asChild>
              <a href={DAYTONA_SLACK_URL} className=" h-8 py-0" target="_blank" rel="noopener noreferrer">
                <Slack size={16} strokeWidth={1.5} />
                <span>Slack</span>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem key="docs">
            <SidebarMenuButton asChild>
              <a href={DAYTONA_DOCS_URL} className=" h-8 py-0" target="_blank" rel="noopener noreferrer">
                <BookOpen size={16} strokeWidth={1.5} />
                <span>Docs</span>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton className="h-12 flex flex-shrink-0 items-center outline outline-1 outline-border outline-offset-0 bg-muted font-medium mt-2">
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
                <DropdownMenuItem asChild>
                  <Button variant="ghost" className="w-full cursor-pointer justify-start" asChild>
                    <Link to={RoutePath.ACCOUNT_SETTINGS}>
                      <Settings className="w-4 h-4" />
                      Account Settings
                    </Link>
                  </Button>
                </DropdownMenuItem>
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
            <div className="flex items-center w-full justify-center gap-2 pt-2">
              <span className="text-xs text-muted-foreground">Version {version}</span>
            </div>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </SidebarComponent>
  )
}

interface Suspension {
  suspensionReason: string
  suspendedUntil: Date
  suspendedAt: Date
  suspensionCleanupGracePeriodHours: number
}

function SetupRequiredBanner({ suspension }: { suspension: Suspension }) {
  const isPaymentMethodRequired = suspension.suspensionReason === 'Payment method required'

  const config = isPaymentMethodRequired
    ? {
        icon: <CreditCard size={16} strokeWidth={1.5} />,
        title: 'Setup Required',
        description: 'Add a payment method to start creating sandboxes.',
        action: (
          <Button size="sm" className="mt-2" asChild>
            <Link to={RoutePath.BILLING_WALLET}>
              Go to Billing <ArrowRightIcon size={16} strokeWidth={1.5} />
            </Link>
          </Button>
        ),
      }
    : {
        icon: <Mail size={16} strokeWidth={1.5} />,
        title: 'Verification Required',
        description: 'Please verify your email address to access all features.',
      }

  return (
    <Alert variant="info">
      {config.icon}
      <AlertTitle> {config.title}</AlertTitle>
      <AlertDescription>
        {config.description}
        {config.action}
      </AlertDescription>
    </Alert>
  )
}

// todo: enumerate
function isSetupRequiredSuspension(reason: string) {
  return reason === 'Payment method required' || reason === 'Please verify your email address'
}

function SuspensionBanner({ suspension }: { suspension: Suspension }) {
  if (isSetupRequiredSuspension(suspension.suspensionReason)) {
    return <SetupRequiredBanner suspension={suspension} />
  }

  return (
    <Alert variant="destructive">
      <TriangleAlertIcon />
      <AlertTitle>Organization suspended</AlertTitle>
      <AlertDescription>
        {suspension.suspensionReason && (
          <div className="text-ellipsis overflow-hidden">{suspension.suspensionReason}</div>
        )}
        <div className="text-xs">
          Sandboxes will be stopped{' '}
          {formatRelative(
            addHours(new Date(suspension.suspendedAt), suspension.suspensionCleanupGracePeriodHours),
            new Date(suspension.suspendedAt),
          )}
        </div>
      </AlertDescription>
    </Alert>
  )
}
