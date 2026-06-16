/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { GithubLogoIcon } from '@phosphor-icons/react'
import { ComponentType } from 'react'
import { Link2, Mail } from 'lucide-react'

type Props = {
  provider: string
  className?: string
}

export function AccountProviderIcon(props: Props) {
  return getIcon(props.provider, props.className)
}

const getIcon = (provider: string, className?: string) => {
  const IconComponent = ICON[provider]

  if (!IconComponent) {
    return <Link2 className={className} />
  }

  return <IconComponent className={className} />
}

const ICON: { [x: string]: ComponentType<{ className?: string }> } = {
  github: GithubLogoIcon,
  'google-oauth2': Mail,
}
