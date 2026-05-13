/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { TypedConfigService } from './typed-config.service'

describe('TypedConfigService', () => {
  const createTypedConfigService = (entries: [string, any][] = []) => {
    const mockConfig = new Map<string, any>([
      ['redis.host', 'localhost'],
      ['redis.port', 6379],
      ['redis.username', 'default-user'],
      ['redis.password', 'default-password'],
      ['redis.tls', undefined],
      ['skipConnections', false],
      ['redis.mode', 'single'],
      ...entries,
    ])

    const configService = {
      get: (key: string) => mockConfig.get(key),
    } as any

    return new TypedConfigService(configService)
  }

  describe('getRedisConfig', () => {
    it('still returns the single-node redis config unchanged', () => {
      const typedConfig = createTypedConfigService()

      expect(typedConfig.getRedisConfig()).toEqual({
        host: 'localhost',
        port: 6379,
        username: 'default-user',
        password: 'default-password',
        tls: undefined,
        lazyConnect: false,
      })
    })
  })

  describe('getRedisModuleOptions', () => {
    it('returns single-mode redis module options by default', () => {
      const typedConfig = createTypedConfigService([['redis.mode', undefined]])

      expect(typedConfig.getRedisModuleOptions()).toEqual({
        type: 'single',
        options: {
          host: 'localhost',
          port: 6379,
          username: 'default-user',
          password: 'default-password',
          tls: undefined,
          lazyConnect: false,
        },
      })
    })

    it('returns cluster-mode redis module options with parsed nodes', () => {
      const typedConfig = createTypedConfigService([
        ['redis.mode', 'cluster'],
        ['redis.clusterNodes', 'host1:7000,host2:7001,host3:7002'],
      ])

      expect(typedConfig.getRedisModuleOptions()).toEqual({
        type: 'cluster',
        nodes: [
          { host: 'host1', port: 7000 },
          { host: 'host2', port: 7001 },
          { host: 'host3', port: 7002 },
        ],
        options: {
          keyPrefix: undefined,
          redisOptions: {
            username: 'default-user',
            password: 'default-password',
            tls: undefined,
            lazyConnect: false,
          },
        },
      })
    })

    it('passes keyPrefix to ClusterOptions level in cluster mode', () => {
      const typedConfig = createTypedConfigService([
        ['redis.mode', 'cluster'],
        ['redis.clusterNodes', 'host1:7000'],
      ])

      const result = typedConfig.getRedisModuleOptions({ keyPrefix: 'throttler:' })

      expect(result).toEqual({
        type: 'cluster',
        nodes: [{ host: 'host1', port: 7000 }],
        options: {
          keyPrefix: 'throttler:',
          redisOptions: {
            username: 'default-user',
            password: 'default-password',
            tls: undefined,
            lazyConnect: false,
          },
        },
      })
    })

    it('passes overrides through in single mode', () => {
      const typedConfig = createTypedConfigService()

      expect(typedConfig.getRedisModuleOptions({ keyPrefix: 'throttler:' })).toEqual({
        type: 'single',
        options: {
          host: 'localhost',
          port: 6379,
          username: 'default-user',
          password: 'default-password',
          tls: undefined,
          lazyConnect: false,
          keyPrefix: 'throttler:',
        },
      })
    })

    it('throws when cluster mode is enabled without cluster nodes', () => {
      const typedConfig = createTypedConfigService([['redis.mode', 'cluster']])

      expect(() => typedConfig.getRedisModuleOptions()).toThrow('Configuration key "redis.clusterNodes" is undefined')
    })

    it('throws on trailing-comma-only cluster nodes', () => {
      const typedConfig = createTypedConfigService([
        ['redis.mode', 'cluster'],
        ['redis.clusterNodes', ',,'],
      ])

      expect(() => typedConfig.getRedisModuleOptions()).toThrow(
        'redis.clusterNodes must contain at least one valid host',
      )
    })

    it('throws on invalid port in cluster node', () => {
      const typedConfig = createTypedConfigService([
        ['redis.mode', 'cluster'],
        ['redis.clusterNodes', 'host1:abc'],
      ])

      expect(() => typedConfig.getRedisModuleOptions()).toThrow('Invalid Redis cluster node: "host1:abc"')
    })
  })

  describe('parseClusterNodes', () => {
    it('parses a single cluster node with explicit port', () => {
      const typedConfig = createTypedConfigService()
      const parseClusterNodes = Reflect.get(typedConfig, 'parseClusterNodes').bind(typedConfig)

      expect(parseClusterNodes('host1:7000')).toEqual([{ host: 'host1', port: 7000 }])
    })

    it('parses multiple cluster nodes', () => {
      const typedConfig = createTypedConfigService()
      const parseClusterNodes = Reflect.get(typedConfig, 'parseClusterNodes').bind(typedConfig)

      expect(parseClusterNodes('host1:7000,host2:7001,host3:7002')).toEqual([
        { host: 'host1', port: 7000 },
        { host: 'host2', port: 7001 },
        { host: 'host3', port: 7002 },
      ])
    })

    it('uses the default port when one is not provided', () => {
      const typedConfig = createTypedConfigService()
      const parseClusterNodes = Reflect.get(typedConfig, 'parseClusterNodes').bind(typedConfig)

      expect(parseClusterNodes('host1')).toEqual([{ host: 'host1', port: 6379 }])
    })

    it('filters empty entries from trailing commas', () => {
      const typedConfig = createTypedConfigService()
      const parseClusterNodes = Reflect.get(typedConfig, 'parseClusterNodes').bind(typedConfig)

      expect(parseClusterNodes('host1:7000,,host2:7001,')).toEqual([
        { host: 'host1', port: 7000 },
        { host: 'host2', port: 7001 },
      ])
    })
  })
})
