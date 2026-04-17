/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Chart, parseChart } from '../types/Charts'
import { ExecutionArtifacts } from '../types/ExecuteResponse'

/**
 * Utility class for parsing artifacts from command output
 */
export class ArtifactParser {
  /**
   * Parses artifacts from command output text
   *
   * @param output - Raw output from command execution
   * @returns Parsed artifacts including stdout and charts
   */
  public static parseArtifacts(output: string): ExecutionArtifacts {
    const charts: Chart[] = []
    let stdout = output

    // Split output by lines to find artifact markers
    const lines = output.split('\n')
    const artifactLines: string[] = []

    for (const line of lines) {
      // Look for the artifact marker pattern
      if (line.startsWith('dtn_artifact_k39fd2:')) {
        artifactLines.push(line)

        try {
          const artifactJson = line.substring('dtn_artifact_k39fd2:'.length).trim()
          const artifactData = JSON.parse(artifactJson)

          if (artifactData.type === 'chart' && artifactData.value) {
            const chartData = artifactData.value
            charts.push(parseChart(chartData))
          }
        } catch (error) {
          // Skip invalid artifacts
          console.warn('Failed to parse artifact:', error)
        }
      }
    }

    // Remove artifact lines from stdout along with their following newlines
    for (const line of artifactLines) {
      stdout = stdout.replace(line + '\n', '')
      stdout = stdout.replace(line, '')
    }

    return {
      stdout,
      charts: charts.length > 0 ? charts : undefined,
    }
  }
}
