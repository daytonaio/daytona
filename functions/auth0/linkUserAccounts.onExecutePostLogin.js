/**
 * Handler that will be called during the execution of a PostLogin flow.
 *
 * @param {Event} event - Details about the user and the context in which they are logging in.
 * @param {PostLoginAPI} api - Interface whose methods can be used to change the behavior of the login.
 */
exports.onExecutePostLogin = async (event, api) => {
  // Early exit if not signup
  if (!event.authorization || event.stats.logins_count != 1) {
    console.log('POST LOGIN EARLY EXIT')
    return
  }

  if (!event.user.email) {
    return api.access.deny('Please ensure your email address is public')
  }

  const ManagementClient = require('auth0').ManagementClient

  const management = new ManagementClient({
    domain: event.secrets.DOMAIN,
    clientId: event.secrets.CLIENT_ID,
    clientSecret: event.secrets.CLIENT_SECRET,
    scope: 'read:users update:users',
  })

  // Search for users with the same email
  let users = []

  try {
    users = await management.getUsersByEmail(event.user.email)
  } catch (error) {
    console.error('Failed to fetch users:', error)
    return api.access.deny('Something went wrong, please try again later')
  }

  try {
    // Early exit if first account with this email address
    if (users.length <= 1) {
      return
    }

    const primaryUser = users.find((u) => u.user_id !== event.user.user_id)

    if (!primaryUser) {
      throw new Error('Primary user not found')
    }

    if (!event.user.email_verified) {
      throw new Error('Secondary user email is not verified')
    }

    // Link current user to primary user if both have verified email
    if (primaryUser.email_verified) {
      // Extract provider from event.user.user_id
      // Format is typically "provider|..." (e.g., "google-oauth2|123456")
      const [provider] = event.user.user_id.split('|')

      await management.linkUsers(primaryUser.user_id, {
        provider,
        user_id: event.user.user_id,
        connection_id: event.connection.id,
      })

      // Switch to the primary user for the login
      api.authentication.setPrimaryUser(primaryUser.user_id)
    }
  } catch (error) {
    console.error('Account linking failed:', error)
    return api.access.deny('Something went wrong, please try again later')
  }
}
