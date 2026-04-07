/**
 * Auth token refresh synchronisation bridge.
 *
 * The HTTP client (api/client.ts) must notify the auth store when it silently
 * refreshes a token, so the store's in-memory state stays consistent with
 * localStorage. A direct store import from client.ts would create a circular
 * dependency (store → api/auth → api/client → store), so we use a
 * lightweight handler pattern as a one-way bridge instead.
 */

export interface AuthTokenRefreshedDetail {
  access_token: string
  refresh_token?: string
  expires_at?: number
}

let authTokenRefreshHandler: ((detail: AuthTokenRefreshedDetail) => void) | null = null

/** Registered by the auth store on initialisation. */
export function setAuthTokenRefreshHandler(
  handler: ((detail: AuthTokenRefreshedDetail) => void) | null
): void {
  authTokenRefreshHandler = handler
}

/** Called by the HTTP interceptor after a successful silent token refresh. */
export function emitAuthTokenRefreshed(detail: AuthTokenRefreshedDetail): void {
  authTokenRefreshHandler?.(detail)
}
