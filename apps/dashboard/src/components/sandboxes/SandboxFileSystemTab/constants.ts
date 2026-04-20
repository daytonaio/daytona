/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { SandboxFileSystemNode } from './types'

export const ROOT_PATH = '/'
// Hard cap for any in-browser preview payload.
export const MAX_PREVIEW_BYTES = 10 * 1024 * 1024
// Default point where wrapped plain text becomes too expensive to keep on.
export const LARGE_TEXT_WRAP_THRESHOLD = 256 * 1024
// Plain-text files above this switch to the virtualized line viewer when wrap is off.
export const LARGE_TEXT_VIRTUALIZATION_THRESHOLD = 512 * 1024
// Preferred max width of the files column before we fall back to overlay mode.
export const FILES_COLUMN_MAX_WIDTH = 360
// Minimum usable width for the contents pane in split view.
export const CONTENTS_OVERLAY_MIN_WIDTH = 350
// Combined width threshold below which contents render as an overlay.
export const CONTENTS_OVERLAY_BREAKPOINT = FILES_COLUMN_MAX_WIDTH + CONTENTS_OVERLAY_MIN_WIDTH
// Top/bottom breathing room for the virtualized file list.
export const FILE_TREE_EDGE_PADDING = 8
// Horizontal indent added per tree depth level.
export const FILE_TREE_INDENT = 16
// Base left inset before tree indentation starts.
export const FILE_TREE_BASE_PADDING = 4
// Width reserved for the expand/collapse toggle lane.
export const FILE_TREE_TOGGLE_SIZE = 32
// Center point used to align indentation guides with the toggle chevron.
export const FILE_TREE_TOGGLE_CENTER = FILE_TREE_TOGGLE_SIZE / 2
// Horizontal row padding used when positioning guides.
export const FILE_TREE_ROW_PADDING_X = 8
// Search must reach this many characters before querying.
export const FILE_SEARCH_MIN_CHARS = 3
// Width reserved in search rows for non-label UI so pretext can measure correctly.
export const FILE_SEARCH_RESULT_LABEL_RESERVED_WIDTH = 40

export const ROOT_NODE: SandboxFileSystemNode = {
  group: 'root',
  id: ROOT_PATH,
  isDir: true,
  modTime: '',
  mode: '',
  name: ROOT_PATH,
  owner: 'root',
  path: ROOT_PATH,
  permissions: '',
  size: 0,
}
