export interface AuthTokenRefreshedDetail {
  access_token: string
  refresh_token?: string
  expires_at?: number
}

let authTokenRefreshHandler: ((detail: AuthTokenRefreshedDetail) => void) | null = null

export function setAuthTokenRefreshHandler(
  handler: ((detail: AuthTokenRefreshedDetail) => void) | null
): void {
  authTokenRefreshHandler = handler
}

export function emitAuthTokenRefreshed(detail: AuthTokenRefreshedDetail): void {
  authTokenRefreshHandler?.(detail)
}
