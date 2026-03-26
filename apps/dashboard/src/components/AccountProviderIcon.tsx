/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ComponentType } from 'react'
import { Github, Link2, Mail, LucideProps } from 'lucide-react'

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

const ICON: { [x: string]: ComponentType<LucideProps> } = {
  github: Github,
  'google-oauth2': Mail,
}
