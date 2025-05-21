/**
 Unstable SemConv
 Because the "incubating" entry-point may include breaking changes in minor versions,
 it is recommended that instrumentation libraries not import @opentelemetry/semantic-conventions/incubating in runtime code,
 but instead copy relevant definitions into their own code base. (This is the same recommendation as for other languages.)
 
 See: https://www.npmjs.com/package/@opentelemetry/semantic-conventions#:~:text=%7D)%3B-,Unstable%20SemConv,-Because%20the%20%22incubating
 */

export const ATTR_DEPLOYMENT_ENVIRONMENT_NAME = 'deployment.environment.name'
