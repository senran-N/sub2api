/**
 * Shared localStorage key constants for authentication tokens.
 *
 * Centralised here to prevent the HTTP client and the auth store from
 * maintaining independent copies of the same key strings (DRY).
 * Do NOT import auth-store logic from this file — it must remain
 * dependency-free so the HTTP interceptor can import it without
 * creating a circular dependency.
 */

export const AUTH_TOKEN_KEY = 'auth_token'
export const AUTH_USER_KEY = 'auth_user'
export const REFRESH_TOKEN_KEY = 'refresh_token'
export const TOKEN_EXPIRES_AT_KEY = 'token_expires_at'
