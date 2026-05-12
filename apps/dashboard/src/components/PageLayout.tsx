/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DAYTONA_SLACK_URL } from '@/constants/ExternalLinks'
import { useTheme } from '@/contexts/ThemeContext'
import { RoutePath } from '@/enums/RoutePath'
import { useUserOrganizationInvitations } from '@/hooks/useUserOrganizationInvitations'
import { cn } from '@/lib/utils'
import {
  ChevronsUpDown,
  LifeBuoyIcon,
  ListChecks,
  LogOut,
  Mail,
  MoonIcon,
  Settings,
  SlackIcon,
  SquareUserRound,
  SunIcon,
} from 'lucide-react'
import { usePostHog } from 'posthog-js/react'
import { type ComponentProps, type ReactNode, useLayoutEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useAuth } from 'react-oidc-context'
import { Link } from 'react-router-dom'
import { BannerStack } from './Banner'
import { FeedbackSheet } from './FeedbackSheet'
import { Button } from './ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from './ui/dropdown-menu'
import { SidebarTrigger } from './ui/sidebar'

function PageLayout({ className, contained = false, ...props }: ComponentProps<'div'> & { contained?: boolean }) {
  return (
    <div
      className={cn('flex h-full flex-col group/page', { 'max-h-screen overflow-hidden': contained }, className)}
      {...props}
    />
  )
}

function PageHeader({ className, children, ...props }: ComponentProps<'header'>) {
  return (
    <header
      className={cn(
        'flex gap-2 sm:gap-4 items-center border-b border-border px-4 py-[15px] bg-background z-10 group-[:has([data-slot=page-banner]:not(:empty))]/page:border-b-transparent min-h-[55px]',
        className,
      )}
      {...props}
    >
      <SidebarTrigger className="shrink-0 [&_svg]:size-5 md:hidden" />
      <div className="flex min-w-0 flex-1 items-center gap-2 sm:gap-4">{children}</div>
      <PageHeaderGlobalActions />
    </header>
  )
}

function PageHeaderGlobalActions() {
  return (
    <div className="flex shrink-0 items-center">
      <FeedbackSheet />
      <PageHeaderExternalAction label="Slack" href={DAYTONA_SLACK_URL} icon={<SlackIcon className="size-4" />} />
      <PageHeaderMockAction label="Support" icon={<LifeBuoyIcon className="size-4" />} />
      <PageHeaderProfileMenu />
    </div>
  )
}

function PageHeaderExternalAction({ label, href, icon }: { label: string; href: string; icon: ReactNode }) {
  return (
    <Button
      variant="ghost"
      size="sm"
      className="px-2 text-muted-foreground hover:text-foreground md:px-3"
      aria-label={label}
      asChild
    >
      <a href={href} target="_blank" rel="noopener noreferrer">
        {icon}
        <span className="hidden md:inline">{label}</span>
      </a>
    </Button>
  )
}

function PageHeaderMockAction({ label, icon }: { label: string; icon: ReactNode }) {
  return (
    <Button
      type="button"
      variant="ghost"
      size="sm"
      className="px-2 text-muted-foreground hover:text-foreground md:px-3"
      aria-label={label}
    >
      {icon}
      <span className="hidden md:inline">{label}</span>
    </Button>
  )
}

function PageHeaderProfileMenu() {
  const posthog = usePostHog()
  const { theme, setTheme } = useTheme()
  const { user, signoutRedirect } = useAuth()
  const { count: organizationInvitationsCount } = useUserOrganizationInvitations()

  const handleSignOut = () => {
    posthog?.reset()
    signoutRedirect()
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          type="button"
          variant="outline"
          size="sm"
          className="ml-1 h-8 max-w-44 gap-2 bg-input/50 px-2 text-muted-foreground hover:text-foreground md:px-2.5"
          aria-label="Profile"
        >
          {user?.profile.picture ? (
            <img
              src={user.profile.picture}
              alt={user.profile.name || 'Profile picture'}
              className="size-4 shrink-0 rounded-sm"
            />
          ) : (
            <SquareUserRound className="size-4 shrink-0" />
          )}
          <span className="hidden min-w-0 truncate md:block">
            {user?.profile.name || user?.profile.email || 'Profile'}
          </span>
          <ChevronsUpDown className="hidden size-4 shrink-0 opacity-50 md:block" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent side="bottom" align="end" className="w-64">
        <div className="px-2 py-1.5">
          <div className="truncate text-sm font-medium">{user?.profile.name || 'Profile'}</div>
          <div className="truncate text-xs text-muted-foreground">{user?.profile.email || ''}</div>
        </div>
        <DropdownMenuSeparator />
        <DropdownMenuItem asChild>
          <Link to={RoutePath.ACCOUNT_SETTINGS}>
            <Settings className="size-4" />
            Account Settings
          </Link>
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}>
          {theme === 'dark' ? <SunIcon className="size-4" /> : <MoonIcon className="size-4" />}
          {theme === 'dark' ? 'Light mode' : 'Dark mode'}
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem asChild>
          <Link to={RoutePath.USER_INVITATIONS}>
            <Mail className="size-4" />
            Invitations
            {organizationInvitationsCount > 0 && (
              <span className="ml-auto rounded-full bg-secondary px-2 py-0.5 text-xs font-medium">
                {organizationInvitationsCount}
              </span>
            )}
          </Link>
        </DropdownMenuItem>
        <DropdownMenuItem asChild>
          <Link to={RoutePath.ONBOARDING}>
            <ListChecks className="size-4" />
            Onboarding
          </Link>
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={handleSignOut}>
          <LogOut className="size-4" />
          Sign out
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

function PageTitle({ className, children, ...props }: ComponentProps<'h1'>) {
  return (
    <h1 className={cn('text-2xl font-medium tracking-tight', className)} {...props}>
      {children}
    </h1>
  )
}

function PageIntro({
  title,
  desc,
  actions,
  className,
}: {
  title: ReactNode
  desc?: ReactNode
  actions?: ReactNode
  className?: string
}) {
  return (
    <div className={cn('mb-8 shrink-0 flex flex-col gap-1', className)}>
      <div className="flex min-w-0 flex-wrap items-start justify-between gap-x-4 gap-y-3">
        <div className="flex min-w-0 flex-1 flex-col gap-1">
          <PageTitle>{title}</PageTitle>
          {desc ? <div className="text-sm text-muted-foreground">{desc}</div> : null}
        </div>
        {actions ? (
          <div className="ml-auto flex shrink-0 flex-wrap items-center justify-end gap-x-3 gap-y-2">{actions}</div>
        ) : null}
      </div>
    </div>
  )
}

function PageBanner({ className, children, ...props }: ComponentProps<'div'>) {
  return (
    <div data-slot="page-banner" className={cn('w-full relative z-30 empty:hidden', className)} {...props}>
      {children}
    </div>
  )
}

function PageContent({
  className,
  size = 'default',
  ...props
}: ComponentProps<'main'> & { size?: 'default' | 'full' }) {
  return (
    <>
      <PageBanner>
        <BannerStack bannerClassName={cn({ 'max-w-5xl mx-auto': size === 'default' })} />
      </PageBanner>
      <main
        className={cn(
          'flex flex-col gap-4 p-4 w-full flex-1 min-h-0 overflow-auto',
          {
            'max-w-5xl mx-auto': size === 'default',
          },
          className,
        )}
        {...props}
      />
    </>
  )
}

function PageFooterPortal({ children }: { children: ReactNode }): ReactNode {
  const [container, setContainer] = useState<Element | null>(null)

  useLayoutEffect(() => {
    setContainer(document.querySelector('[data-slot="page-footer"]'))
  }, [])

  if (!container) return children

  return <>{createPortal(children, container)}</>
}

function PageFooter({ className, children, ...props }: ComponentProps<'footer'>) {
  return (
    <footer
      data-slot="page-footer"
      className={cn(
        'flex gap-2 sm:gap-4 items-center border-t border-border p-4 bg-background z-10 empty:hidden',
        className,
      )}
      {...props}
    >
      {children}
    </footer>
  )
}

export { PageContent, PageFooter, PageFooterPortal, PageHeader, PageIntro, PageLayout, PageTitle }
