/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { ReactNode } from 'react'
import awsIcon from '@/assets/aws.svg'
import dockerIcon from '@/assets/docker.svg'
import githubIcon from '@/assets/github.svg'
import googleIcon from '@/assets/google.svg'
import { DAYTONA_DOCS_URL } from './ExternalLinks'

export type RegistryProvider = 'generic' | 'dockerhub' | 'gcp' | 'ghcr' | 'ecr'

export const REGISTRY_PROVIDER_VALUES: readonly RegistryProvider[] = [
  'generic',
  'dockerhub',
  'gcp',
  'ghcr',
  'ecr',
] as const

// Full names — used for aria-label / tooltip / Zod error messages.
export const REGISTRY_PROVIDER_LABELS: Record<RegistryProvider, string> = {
  generic: 'Generic',
  dockerhub: 'Docker Hub',
  gcp: 'Google Artifact Registry',
  ghcr: 'GitHub Container Registry',
  ecr: 'Amazon ECR',
}

// Tab content — text for "generic" (no universal icon), brand SVG for the rest.
// Mono-color marks (GitHub, AWS-orange-on-dark) get `dark:invert` so they stay
// legible across themes; the multi-color marks render as-is.
const ICON_CLASS = 'h-4 w-4'
export const REGISTRY_PROVIDER_TAB_CONTENT: Record<RegistryProvider, ReactNode> = {
  generic: 'Generic',
  dockerhub: <img src={dockerIcon} alt="" className={ICON_CLASS} />,
  gcp: <img src={googleIcon} alt="" className={ICON_CLASS} />,
  ghcr: <img src={githubIcon} alt="" className={`${ICON_CLASS} dark:invert`} />,
  ecr: <img src={awsIcon} alt="" className={ICON_CLASS} />,
}

export interface ProviderFieldSpec {
  label: string
  // Skip rendering. Submit uses `defaultValue` (or empty) as the payload.
  hidden?: boolean
  // Reject empty input on submit. Has no effect when hidden.
  required?: boolean
  placeholder?: string
  // Visible: initial value populated on tab switch. Hidden: value sent on submit.
  defaultValue?: string
  readOnly?: boolean
  multiline?: boolean
  // Render as masked input with an eye toggle. Ignored when multiline.
  secret?: boolean
  helper?: ReactNode
}

export interface ProviderFormSpec {
  url: ProviderFieldSpec
  username: ProviderFieldSpec
  password: ProviderFieldSpec
  project: ProviderFieldSpec
}

export const REGISTRY_PROVIDER_SPECS: Record<RegistryProvider, ProviderFormSpec> = {
  generic: {
    url: {
      label: 'Registry URL',
      placeholder: 'https://registry.example.com',
      helper: 'Defaults to docker.io when left blank.',
    },
    username: { required: true, label: 'Username' },
    password: { required: true, label: 'Password', secret: true },
    project: {
      label: 'Project',
      placeholder: 'my-project',
      helper: 'Leave empty for private Docker Hub entries.',
    },
  },
  dockerhub: {
    // Always docker.io — auto-filled, no input.
    url: { hidden: true, label: 'Registry URL', defaultValue: 'docker.io' },
    username: {
      required: true,
      label: 'Username',
      helper: 'Your Docker Hub username.',
    },
    password: {
      required: true,
      label: 'Personal Access Token',
      secret: true,
      helper: (
        <>
          Use a{' '}
          <a href="https://docs.docker.com/security/access-tokens/" target="_blank" rel="noopener noreferrer">
            Docker Hub PAT
          </a>
          , not your account password.
        </>
      ),
    },
    project: { hidden: true, label: 'Project' },
  },
  gcp: {
    url: {
      required: true,
      label: 'Registry URL',
      placeholder: 'https://us-central1-docker.pkg.dev',
      helper: 'Base URL for your region.',
    },
    // Always _json_key for service-account auth — auto-filled, no input.
    username: { hidden: true, label: 'Username', defaultValue: '_json_key' },
    password: {
      required: true,
      label: 'Service Account JSON Key',
      multiline: true,
      placeholder: '{\n  "type": "service_account",\n  ...\n}',
      helper: 'Paste the full contents of your service account key JSON file.',
    },
    project: {
      label: 'Google Cloud Project ID',
      placeholder: 'my-gcp-project',
      helper: 'Your GCP project ID.',
    },
  },
  ghcr: {
    // Always ghcr.io — auto-filled, no input.
    url: { hidden: true, label: 'Registry URL', defaultValue: 'ghcr.io' },
    username: {
      required: true,
      label: 'GitHub Username',
      helper: 'The account with access to the image.',
    },
    password: {
      required: true,
      label: 'Personal Access Token',
      secret: true,
      helper: (
        <>
          Use a{' '}
          <a
            href="https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens"
            target="_blank"
            rel="noopener noreferrer"
          >
            GitHub PAT
          </a>{' '}
          with <code>read:packages</code> scope.
        </>
      ),
    },
    project: { hidden: true, label: 'Project' },
  },
  ecr: {
    url: {
      required: true,
      label: 'Registry URL',
      placeholder: '123456789012.dkr.ecr.us-east-1.amazonaws.com',
    },
    username: {
      required: true,
      label: 'Role ARN',
      placeholder: 'arn:aws:iam::123456789012:role/daytona-ecr-puller',
      helper: (
        <>
          Daytona will assume this role on every pull.{' '}
          <a
            href={`${DAYTONA_DOCS_URL}/snapshots#amazon-elastic-container-registry-ecr`}
            target="_blank"
            rel="noopener noreferrer"
          >
            Set up the role ↗
          </a>
        </>
      ),
    },
    password: { hidden: true, label: 'Password' },
    project: { hidden: true, label: 'Project' },
  },
}
