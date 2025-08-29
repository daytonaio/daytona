/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Track job execution in activeJobs set.
 * @returns A decorator function that tracks execution of a job.
 */
export function TrackJobExecution() {
  return function (target: any, propertyKey: string, descriptor: PropertyDescriptor) {
    const original = descriptor.value

    descriptor.value = async function (...args: any[]) {
      if (!this.activeJobs) {
        throw new Error(`@TrackExecution requires 'activeJobs' property on ${target.constructor.name}`)
      }

      this.activeJobs.add(propertyKey)
      try {
        return await original.apply(this, args)
      } finally {
        this.activeJobs.delete(propertyKey)
      }
    }
  }
}
