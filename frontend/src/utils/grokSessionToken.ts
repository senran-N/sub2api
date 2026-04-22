const GROK_SESSION_COOKIE_SSO = 'sso'
const GROK_SESSION_COOKIE_SSO_RW = 'sso-rw'
const GROK_SESSION_COOKIE_CF_CLEARANCE = 'cf_clearance'
const GROK_SESSION_COOKIE_X_ANON_USER_ID = 'x-anonuserid'
const MIN_GROK_SESSION_TOKEN_LENGTH = 24

interface GrokSessionCookieJar {
  order: string[]
  values: Map<string, string>
}

function trimCookieHeaderPrefix(raw: string): string {
  return raw.trim().replace(/^cookie:\s*/i, '')
}

function createCookieJar(): GrokSessionCookieJar {
  return {
    order: [],
    values: new Map<string, string>()
  }
}

function setCookie(jar: GrokSessionCookieJar, name: string, value: string): void {
  const normalizedName = name.trim().toLowerCase()
  const normalizedValue = value.trim()
  if (!normalizedName || !normalizedValue) {
    return
  }
  if (!jar.values.has(normalizedName)) {
    jar.order.push(normalizedName)
  }
  jar.values.set(normalizedName, normalizedValue)
}

function getCookie(jar: GrokSessionCookieJar, name: string): string {
  return jar.values.get(name.trim().toLowerCase())?.trim() ?? ''
}

function parseCookieHeader(raw: string): GrokSessionCookieJar | null {
  const trimmed = trimCookieHeaderPrefix(raw)
  if (!trimmed) {
    return createCookieJar()
  }

  const jar = createCookieJar()
  for (const part of trimmed.split(';')) {
    const trimmedPart = part.trim()
    if (!trimmedPart) {
      continue
    }
    const separator = trimmedPart.indexOf('=')
    if (separator <= 0) {
      continue
    }
    const name = trimmedPart.slice(0, separator).trim()
    const value = trimmedPart.slice(separator + 1).trim()
    if (!name || !value) {
      continue
    }
    setCookie(jar, name, value)
  }

  return jar.values.size > 0 ? jar : null
}

function buildCookieHeader(jar: GrokSessionCookieJar): string {
  const orderedNames = [
    GROK_SESSION_COOKIE_SSO,
    GROK_SESSION_COOKIE_SSO_RW,
    GROK_SESSION_COOKIE_CF_CLEARANCE,
    GROK_SESSION_COOKIE_X_ANON_USER_ID,
    ...jar.order.filter((name) =>
      ![
        GROK_SESSION_COOKIE_SSO,
        GROK_SESSION_COOKIE_SSO_RW,
        GROK_SESSION_COOKIE_CF_CLEARANCE,
        GROK_SESSION_COOKIE_X_ANON_USER_ID
      ].includes(name)
    )
  ]

  const parts: string[] = []
  for (const name of orderedNames) {
    const value = getCookie(jar, name)
    if (value) {
      parts.push(`${name}=${value}`)
    }
  }
  return parts.join('; ')
}

function isValidPrimarySessionToken(token: string): boolean {
  const normalized = token.trim()
  return (
    normalized.length >= MIN_GROK_SESSION_TOKEN_LENGTH &&
    !/[\s;]/.test(normalized)
  )
}

export function normalizeGrokSessionToken(raw: string): string | null {
  const trimmed = raw.trim()
  if (!trimmed) {
    return null
  }

  if (!trimmed.includes('=')) {
    if (!isValidPrimarySessionToken(trimmed)) {
      return null
    }
    return `${GROK_SESSION_COOKIE_SSO}=${trimmed}; ${GROK_SESSION_COOKIE_SSO_RW}=${trimmed}`
  }

  const jar = parseCookieHeader(trimmed)
  if (!jar) {
    return null
  }

  const sessionToken = getCookie(jar, GROK_SESSION_COOKIE_SSO) || getCookie(jar, GROK_SESSION_COOKIE_SSO_RW)
  if (!isValidPrimarySessionToken(sessionToken)) {
    return null
  }

  setCookie(jar, GROK_SESSION_COOKIE_SSO, sessionToken)
  setCookie(jar, GROK_SESSION_COOKIE_SSO_RW, getCookie(jar, GROK_SESSION_COOKIE_SSO_RW) || sessionToken)
  return buildCookieHeader(jar)
}

export function findInvalidGrokSessionBatchImportLine(rawInput: string): number | null {
  const lines = rawInput.split(/\r?\n/)
  for (let index = 0; index < lines.length; index += 1) {
    const line = lines[index]?.trim() ?? ''
    if (!line) {
      continue
    }
    if (!normalizeGrokSessionToken(line)) {
      return index + 1
    }
  }
  return null
}
