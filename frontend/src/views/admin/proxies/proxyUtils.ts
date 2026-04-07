import type { Proxy, ProxyProtocol } from '@/types'

export interface ProxyBatchEntry {
  protocol: ProxyProtocol
  host: string
  port: number
  username: string
  password: string
}

export type ProxyCopyTarget = Pick<Proxy, 'protocol' | 'host' | 'port' | 'username' | 'password'>

export interface ProxyCopyFormat {
  label: string
  value: string
}

export interface ProxyBatchParseResult {
  total: number
  valid: number
  invalid: number
  duplicate: number
  proxies: ProxyBatchEntry[]
}

const PROXY_URL_REGEX = /^(https?|socks5h?):\/\/(?:([^:@]+):([^@]+)@)?([^:]+):(\d+)$/i

export function parseProxyUrl(line: string): ProxyBatchEntry | null {
  const trimmed = line.trim()
  if (!trimmed) {
    return null
  }

  const match = trimmed.match(PROXY_URL_REGEX)
  if (!match) {
    return null
  }

  const [, protocol, username, password, host, port] = match
  const parsedPort = Number.parseInt(port, 10)

  if (parsedPort < 1 || parsedPort > 65535) {
    return null
  }

  return {
    protocol: protocol.toLowerCase() as ProxyProtocol,
    host: host.trim(),
    port: parsedPort,
    username: username?.trim() || '',
    password: password?.trim() || ''
  }
}

export function parseBatchProxyInput(input: string): ProxyBatchParseResult {
  const lines = input
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)

  const seen = new Set<string>()
  const proxies: ProxyBatchEntry[] = []
  let invalid = 0
  let duplicate = 0

  for (const line of lines) {
    const parsed = parseProxyUrl(line)
    if (!parsed) {
      invalid += 1
      continue
    }

    const duplicateKey = `${parsed.host}:${parsed.port}:${parsed.username}:${parsed.password}`
    if (seen.has(duplicateKey)) {
      duplicate += 1
      continue
    }

    seen.add(duplicateKey)
    proxies.push(parsed)
  }

  return {
    total: lines.length,
    valid: proxies.length,
    invalid,
    duplicate,
    proxies
  }
}

export function buildProxyAuthPart(proxy: ProxyCopyTarget): string {
  const username = proxy.username ? encodeURIComponent(proxy.username) : ''
  const password = proxy.password ? encodeURIComponent(proxy.password) : ''

  if (username && password) {
    return `${username}:${password}@`
  }
  if (username) {
    return `${username}@`
  }
  if (password) {
    return `:${password}@`
  }

  return ''
}

export function buildProxyUrl(proxy: ProxyCopyTarget): string {
  return `${proxy.protocol}://${buildProxyAuthPart(proxy)}${proxy.host}:${proxy.port}`
}

export function buildProxyCopyFormats(proxy: ProxyCopyTarget): ProxyCopyFormat[] {
  const fullUrl = buildProxyUrl(proxy)
  const formats: ProxyCopyFormat[] = [{ label: fullUrl, value: fullUrl }]

  if (proxy.username || proxy.password) {
    const withoutProtocol = fullUrl.replace(/^[^:]+:\/\//, '')
    formats.push({ label: withoutProtocol, value: withoutProtocol })
  }

  formats.push({
    label: `${proxy.host}:${proxy.port}`,
    value: `${proxy.host}:${proxy.port}`
  })

  return formats
}
