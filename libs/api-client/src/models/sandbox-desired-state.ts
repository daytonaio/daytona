/* tslint:disable */

/**
 * Daytona
 * Daytona AI platform API Docs
 *
 * The version of the OpenAPI document: 1.0
 * Contact: support@daytona.com
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

/**
 * The desired state of the sandbox
 * @export
 * @enum {string}
 */

export const SandboxDesiredState = {
  DESTROYED: 'destroyed',
  STARTED: 'started',
  STOPPED: 'stopped',
  RESIZED: 'resized',
  ARCHIVED: 'archived',
} as const

export type SandboxDesiredState = (typeof SandboxDesiredState)[keyof typeof SandboxDesiredState]
