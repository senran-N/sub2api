import { describe, expect, it } from 'vitest'
import {
  buildProxyAuthPart,
  buildProxyCopyFormats,
  buildProxyUrl,
  parseBatchProxyInput,
  parseProxyUrl
} from '../proxyUtils'

describe('proxyUtils', () => {
  it('parses proxy urls with optional credentials', () => {
    expect(parseProxyUrl(' socks5://alice:secret@example.com:1080 ')).toEqual({
      protocol: 'socks5',
      host: 'example.com',
      port: 1080,
      username: 'alice',
      password: 'secret'
    })

    expect(parseProxyUrl('https://proxy.local:443')).toEqual({
      protocol: 'https',
      host: 'proxy.local',
      port: 443,
      username: '',
      password: ''
    })
  })

  it('rejects invalid proxy urls and invalid ports', () => {
    expect(parseProxyUrl('')).toBeNull()
    expect(parseProxyUrl('ftp://proxy.local:21')).toBeNull()
    expect(parseProxyUrl('http://proxy.local:70000')).toBeNull()
    expect(parseProxyUrl('http://missing-port')).toBeNull()
  })

  it('summarizes batch input with duplicate and invalid counts', () => {
    expect(
      parseBatchProxyInput([
        'http://one.local:80',
        'http://one.local:80',
        'socks5://alice:secret@two.local:1080',
        'invalid'
      ].join('\n'))
    ).toEqual({
      total: 4,
      valid: 2,
      invalid: 1,
      duplicate: 1,
      proxies: [
        {
          protocol: 'http',
          host: 'one.local',
          port: 80,
          username: '',
          password: ''
        },
        {
          protocol: 'socks5',
          host: 'two.local',
          port: 1080,
          username: 'alice',
          password: 'secret'
        }
      ]
    })
  })

  it('builds proxy auth and copy formats consistently', () => {
    const proxy = {
      protocol: 'http' as const,
      host: 'proxy.local',
      port: 8080,
      username: 'alice@example.com',
      password: 'p@ss word'
    }

    expect(buildProxyAuthPart(proxy)).toBe('alice%40example.com:p%40ss%20word@')
    expect(buildProxyUrl(proxy)).toBe(
      'http://alice%40example.com:p%40ss%20word@proxy.local:8080'
    )
    expect(buildProxyCopyFormats(proxy)).toEqual([
      {
        label: 'http://alice%40example.com:p%40ss%20word@proxy.local:8080',
        value: 'http://alice%40example.com:p%40ss%20word@proxy.local:8080'
      },
      {
        label: 'alice%40example.com:p%40ss%20word@proxy.local:8080',
        value: 'alice%40example.com:p%40ss%20word@proxy.local:8080'
      },
      {
        label: 'proxy.local:8080',
        value: 'proxy.local:8080'
      }
    ])
  })
})
