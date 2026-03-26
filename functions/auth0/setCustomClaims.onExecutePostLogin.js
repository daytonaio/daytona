/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Handler that will be called during the execution of a PostLogin flow.
 *
 * @param {Event} event - Details about the user and the context in which they are logging in.
 * @param {PostLoginAPI} api - Interface whose methods can be used to change the behavior of the login.
 */
exports.onExecutePostLogin = async (event, api) => {
  api.accessToken.setCustomClaim('email', event.user.email)
  api.accessToken.setCustomClaim('name', event.user.name)
  api.accessToken.setCustomClaim('email_verified', event.user.email_verified)
  api.accessToken.setCustomClaim(
    'phone_verified',
    event.user.enrolledFactors && event.user.enrolledFactors.some((f) => f.type === 'phone' && f.method === 'sms'),
  )
  api.accessToken.setCustomClaim('identities', event.user.identities)
  api.idToken.setCustomClaim('identities', event.user.identities)
  api.idToken.setCustomClaim(
    'phone_verified',
    event.user.enrolledFactors && event.user.enrolledFactors.some((f) => f.type === 'phone' && f.method === 'sms'),
  )
}
