/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { z } from 'zod'

const IMAGE_NAME_REGEX = /^[a-zA-Z0-9_.\-:]+(\/[a-zA-Z0-9_.\-:]+)*(@sha256:[a-f0-9]{64})?$/
const IMAGE_TAG_OR_DIGEST_REGEX = /^[^@]+@sha256:[a-f0-9]{64}$|^(?!.*@sha256:).*:.+$/

export const imageNameSchema = z
  .string()
  .min(1, 'Image name is required')
  .refine((name) => IMAGE_NAME_REGEX.test(name), 'Only letters, digits, dots, colons, slashes and dashes are allowed')
  .refine(
    (name) => IMAGE_TAG_OR_DIGEST_REGEX.test(name),
    'Image must include a tag (e.g., ubuntu:22.04) or digest (@sha256:...)',
  )
  .refine((name) => !name.endsWith(':latest'), 'Images with tag ":latest" are not allowed')
