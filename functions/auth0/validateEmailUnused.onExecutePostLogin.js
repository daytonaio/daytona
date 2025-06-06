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
  // Skip validation when performing account linking
  if (event.request.query?.accountLinking === 'true') {
    return
  }

  // User object must have an email address
  if (!event.user.email) {
    return api.access.deny('Please ensure your email address is public')
  }

  const ManagementClient = require('auth0').ManagementClient

  const management = new ManagementClient({
    domain: event.secrets.DOMAIN,
    clientId: event.secrets.CLIENT_ID,
    clientSecret: event.secrets.CLIENT_SECRET,
    scope: 'read:users',
  })

  // Fetch users with this email address in the Auth0 database
  let users = []

  try {
    users = await management.getUsersByEmail(event.user.email)
  } catch (error) {
    console.error('Failed to fetch Auth0 users:', error)
    return api.access.deny('Something went wrong, please try again later')
  }

  // Skip validation if this is the only user with this email address
  if (users.length <= 1) {
    return
  }

  // Deny access if this user doesn't exist in the Daytona database
  try {
    const response = await fetch(`${event.secrets.DAYTONA_API_URL}/users/${event.user.user_id}`, {
      headers: {
        Authorization: `Bearer ${event.secrets.DAYTONA_API_KEY}`,
      },
    })

    if (!response.ok) {
      return api.access.deny('Something went wrong, please try again later')
    }
  } catch (error) {
    console.error('Failed to fetch Daytona users:', error)
    return api.access.deny('Something went wrong, please try again later')
  }
}
