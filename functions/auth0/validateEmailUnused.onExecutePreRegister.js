/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Handler that will be called during the execution of a PreUserRegistration flow.
 *
 * @param {Event} event - Details about the context and user that is attempting to register.
 * @param {PreUserRegistrationAPI} api - Interface whose methods can be used to change the behavior of the signup.
 */
exports.onExecutePreUserRegistration = async (event, api) => {
  const ManagementClient = require('auth0').ManagementClient

  const management = new ManagementClient({
    domain: event.secrets.DOMAIN,
    clientId: event.secrets.CLIENT_ID,
    clientSecret: event.secrets.CLIENT_SECRET,
    scope: 'read:users update:users',
  })

  try {
    // Search for users with the same email
    const users = await management.getUsersByEmail(event.user.email)

    if (users.length >= 1) {
      return api.access.deny('Email already used', 'Something went wrong, please try again later')
    }
  } catch (error) {
    console.error('Failed to fetch users:', error)
    return api.access.deny('Could not fetch users', 'Something went wrong, please try again later')
  }
}
