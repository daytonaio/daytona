import { Injectable, Logger } from '@nestjs/common'
import { ThrottlerGuard, ThrottlerRequest } from '@nestjs/throttler'
import { Request } from 'express'
import { isAuthContext, BaseAuthContext } from '../interfaces/auth-context.interface'

@Injectable()
export class RateLimitGuard extends ThrottlerGuard {
  private readonly logger = new Logger(RateLimitGuard.name)

  protected async getTracker(req: Request): Promise<string> {
    // Check if user is authenticated
    const user = req.user as BaseAuthContext

    if (user && isAuthContext(user)) {
      // For authenticated users, use ogranizationId as tracker
      return `user:${user.organizationId}`
    }

    // For unauthenticated users, use IP address as tracker
    const ip = req.ips.length ? req.ips[0] : req.ip
    return `ip:${ip}`
  }

  async handleRequest(requestProps: ThrottlerRequest): Promise<boolean> {
    const { context, throttler } = requestProps
    const request = context.switchToHttp().getRequest<Request>()
    const user = request.user as BaseAuthContext

    const isAuthenticated = user && isAuthContext(user)

    this.logger.debug('HANDLE REQUEST CALLED')
    this.logger.debug('throttler.name', throttler.name)
    this.logger.debug('isAuthenticated', isAuthenticated)
    this.logger.debug('request.user', request.user)

    // // Potentially authenticated request
    // if(/^Bearer .+/.test(request.headers.authorization)) {
    //   // skip rate limiter
    // }

    // Skip throttlers that don't match our auth state
    if (
      (throttler.name === 'authenticated' && !isAuthenticated) ||
      (throttler.name === 'anonymous' && isAuthenticated)
    ) {
      return true // Skip this throttler
    }

    // Use parent's logic with our custom tracker
    return super.handleRequest({
      ...requestProps,
      getTracker: () => this.getTracker(request),
    })
  }
}
