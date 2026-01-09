/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useEffect, useMemo, useState } from 'react'
import { Outlet } from 'react-router-dom'

import { AnnouncementBanner } from '@/components/AnnouncementBanner'
import { CommandPalette, useRegisterCommands, type CommandConfig } from '@/components/CommandPalette'
import { Sidebar } from '@/components/Sidebar'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
import { Toaster } from '@/components/ui/sonner'
import { VerifyEmailDialog } from '@/components/VerifyEmailDialog'
import { DAYTONA_DOCS_URL, DAYTONA_SLACK_URL } from '@/constants/ExternalLinks'
import { useTheme } from '@/contexts/ThemeContext'
import { LocalStorageKey } from '@/enums/LocalStorageKey'
import { RoutePath } from '@/enums/RoutePath'
import { useOwnerWalletQuery } from '@/hooks/queries/billingQueries'
import { useConfig } from '@/hooks/useConfig'
import { useDocsSearchCommands } from '@/hooks/useDocsSearchCommands'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { cn } from '@/lib/utils'
import { BookOpen, BookSearchIcon, SlackIcon, SunMoon } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

function useDashboardCommands() {
  const { theme, setTheme } = useTheme()

  const helpCommands: CommandConfig[] = useMemo(
    () => [
      {
        id: 'open-slack',
        label: 'Open Slack',
        icon: <SlackIcon className="w-4 h-4" />,
        onSelect: () => window.open(DAYTONA_SLACK_URL, '_blank'),
      },
      {
        id: 'open-docs',
        label: 'Open Docs',
        icon: <BookOpen className="w-4 h-4" />,
        onSelect: () => window.open(DAYTONA_DOCS_URL, '_blank'),
      },
      {
        id: 'search-docs',
        label: 'Search Docs',
        icon: <BookSearchIcon className="w-4 h-4" />,
        page: 'search-docs',
      },
    ],
    [],
  )
  useRegisterCommands(helpCommands, { groupId: 'help', groupLabel: 'Help', groupOrder: 2 })

  const globalCommands: CommandConfig[] = useMemo(
    () => [
      {
        id: 'toggle-theme',
        label: 'Toggle Theme',
        icon: <SunMoon className="w-4 h-4" />,
        onSelect: () => setTheme(theme === 'dark' ? 'light' : 'dark'),
      },
    ],
    [theme, setTheme],
  )
  useRegisterCommands(globalCommands, { groupId: 'global', groupLabel: 'Global', groupOrder: 5 })
}

const Dashboard: React.FC = () => {
  const { selectedOrganization } = useSelectedOrganization()
  const [showVerifyEmailDialog, setShowVerifyEmailDialog] = useState(false)
  const config = useConfig()
  useOwnerWalletQuery() // prefetch wallet

  useDashboardCommands()
  useDocsSearchCommands()

  const navigate = useNavigate()

  useEffect(() => {
    if (
      selectedOrganization?.suspended &&
      selectedOrganization.suspensionReason === 'Please verify your email address'
    ) {
      setShowVerifyEmailDialog(true)
    }
  }, [selectedOrganization])

  useEffect(() => {
    if (!config.billingApiUrl) {
      return
    }

    if (!selectedOrganization) {
      return
    }

    if (!selectedOrganization.defaultRegionId) {
      navigate(RoutePath.SETTINGS)
      return
    }
  }, [config.billingApiUrl, selectedOrganization]) // Do not depend on navigate to avoid infinite loops

  const [bannerText, bannerLearnMoreUrl] = useMemo(() => {
    if (!config.announcements || Object.entries(config.announcements).length === 0) {
      return [null, null]
    }

    return [Object.values(config.announcements)[0].text, Object.values(config.announcements)[0].learnMoreUrl]
  }, [config.announcements])
  const [isBannerVisible, setIsBannerVisible] = useState(false)

  useEffect(() => {
    if (!bannerText) {
      setIsBannerVisible(false)
      return
    }

    // Check if this announcement has been dismissed
    const dismissedBanners = JSON.parse(localStorage.getItem(LocalStorageKey.AnnouncementBannerDismissed) || '[]')
    const isDismissed = dismissedBanners.includes(bannerText)

    setIsBannerVisible(!isDismissed)
  }, [bannerText])

  const handleDismissBanner = () => {
    // Add this announcement to the dismissed list
    const dismissedBanners = JSON.parse(localStorage.getItem(LocalStorageKey.AnnouncementBannerDismissed) || '[]')
    localStorage.setItem(LocalStorageKey.AnnouncementBannerDismissed, JSON.stringify([...dismissedBanners, bannerText]))

    setIsBannerVisible(false)
  }

  return (
    <div className="relative w-full">
      {isBannerVisible && bannerText && (
        <AnnouncementBanner text={bannerText} onDismiss={handleDismissBanner} learnMoreUrl={bannerLearnMoreUrl} />
      )}
      <SidebarProvider isBannerVisible={isBannerVisible} defaultOpen={true}>
        <Sidebar isBannerVisible={isBannerVisible} billingEnabled={!!config.billingApiUrl} version={config.version} />
        <SidebarInset className="overflow-y-auto">
          <div className={cn('w-full min-h-screen overscroll-none', isBannerVisible ? 'md:pt-12' : '')}>
            <Outlet />
            <CommandPalette />
          </div>
        </SidebarInset>
        <Toaster />
        <VerifyEmailDialog open={showVerifyEmailDialog} onOpenChange={setShowVerifyEmailDialog} />
      </SidebarProvider>
    </div>
  )
}

export default Dashboard
