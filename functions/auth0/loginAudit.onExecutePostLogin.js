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
  // User object must have an email address
  if (!event.user.email) {
    return api.access.deny('Please ensure your email address is public')
  }

  // Deny access if login action can't be logged in audit system
  try {
    const response = await fetch(`${event.secrets.DAYTONA_API_URL}/audit`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${event.secrets.DAYTONA_API_KEY}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        actorId: event.user.user_id,
        actorEmail: event.user.email,
        action: 'login',
      }),
    })
    if (!response.ok) {
      console.error('Unable to create audit log for login action')
      return api.access.deny('Something went wrong, please try again later')
    }
  } catch (error) {
    console.error(error)
    return api.access.deny('Something went wrong, please try again later')
  }
}
