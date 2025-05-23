/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export { CodeLanguage, Daytona } from './Daytona'
export type { CreateSandboxParams, DaytonaConfig, SandboxResources, VolumeMount } from './Daytona'
export { FileSystem } from './FileSystem'
export { Git } from './Git'
export { LspLanguageId } from './LspServer'
export { Process } from './Process'
// export { LspServer } from './LspServer'
// export type { LspLanguageId, Position } from './LspServer'
export { DaytonaError } from './errors/DaytonaError'
export { Sandbox } from './Sandbox'
export type { SandboxCodeToolbox } from './Sandbox'

// Chart and artifact types
export { ChartType } from './types/Charts'
export type {
  BarChart,
  BoxAndWhiskerChart,
  Chart,
  CompositeChart,
  LineChart,
  PieChart,
  ScatterChart,
} from './types/Charts'

export { WorkspaceState as SandboxState, CreateWorkspaceTargetEnum as SandboxTargetRegion } from '@daytonaio/api-client'
export type {
  FileInfo,
  GitStatus,
  ListBranchResponse,
  Match,
  ReplaceResult,
  SearchFilesResponse,
} from '@daytonaio/api-client'

// Re-export necessary Workspace-related types for backward compatibility
import { CreateWorkspaceTargetEnum, WorkspaceState as WS } from '@daytonaio/api-client'
import type { CreateSandboxParams, SandboxResources } from './Daytona'
import type { SandboxCodeToolbox } from './Sandbox'
import { Sandbox } from './Sandbox'

/** @deprecated `CreateWorkspaceParams` is deprecated. Please use `CreateSandboxParams` instead. This will be removed in a future version. */
export type CreateWorkspaceParams = CreateSandboxParams

/** @deprecated `Workspace` is deprecated. Please use `Sandbox` instead. This will be removed in a future version. */
export const Workspace = Sandbox
/** @deprecated `Workspace` is deprecated. Please use `Sandbox` instead. This will be removed in a future version. */
export type Workspace = Sandbox

/** @deprecated `WorkspaceCodeToolbox` is deprecated. Please use `SandboxCodeToolbox` instead. This will be removed in a future version. */
export type WorkspaceCodeToolbox = SandboxCodeToolbox

/** @deprecated `WorkspaceResources` is deprecated. Please use `SandboxResources` instead. This will be removed in a future version. */
export type WorkspaceResources = SandboxResources

/** @deprecated `WorkspaceState` is deprecated. Please use `SandboxState` instead. This will be removed in a future version. */
export type WorkspaceState = WS
/** @deprecated `WorkspaceState` is deprecated. Please use `SandboxState` instead. This will be removed in a future version. */
export const WorkspaceState = WS

/** @deprecated `WorkspaceTargetRegion` is deprecated. Please use `SandboxTargetRegion` instead. This will be removed in a future version. */
export const WorkspaceTargetRegion = CreateWorkspaceTargetEnum
/** @deprecated `WorkspaceTargetRegion` is deprecated. Please use `SandboxTargetRegion` instead. This will be removed in a future version. */
export type WorkspaceTargetRegion = CreateWorkspaceTargetEnum
