/**
 * User API endpoints
 * Handles user profile management and password changes
 */

import { apiClient } from './client'
import type { User, ChangePasswordRequest, NotifyEmailEntry } from '@/types'

/**
 * Get current user profile
 * @returns User profile data
 */
export async function getProfile(): Promise<User> {
  const { data } = await apiClient.get<User>('/user/profile')
  return data
}

/**
 * Update current user profile
 * @param profile - Profile data to update
 * @returns Updated user profile data
 */
export async function updateProfile(profile: {
  username?: string
  balance_notify_enabled?: boolean
  balance_notify_threshold?: number | null
  balance_notify_threshold_type?: string
  balance_notify_extra_emails?: NotifyEmailEntry[]
}): Promise<User> {
  const { data } = await apiClient.put<User>('/user', profile)
  return data
}

/**
 * Change current user password
 * @param passwords - Old and new password
 * @returns Success message
 */
export async function changePassword(
  oldPassword: string,
  newPassword: string
): Promise<{ message: string }> {
  const payload: ChangePasswordRequest = {
    old_password: oldPassword,
    new_password: newPassword
  }

  const { data } = await apiClient.put<{ message: string }>('/user/password', payload)
  return data
}

export async function sendNotifyEmailCode(email: string): Promise<void> {
  await apiClient.post('/user/notify-email/send-code', { email })
}

export async function verifyNotifyEmail(email: string, code: string): Promise<User> {
  const { data } = await apiClient.post<User>('/user/notify-email/verify', { email, code })
  return data
}

export async function removeNotifyEmail(email: string): Promise<User> {
  const { data } = await apiClient.delete<User>('/user/notify-email', { data: { email } })
  return data
}

export async function toggleNotifyEmail(email: string, disabled: boolean): Promise<User> {
  const { data } = await apiClient.put<User>('/user/notify-email/toggle', { email, disabled })
  return data
}


export async function sendEmailBindingCode(email: string): Promise<void> {
  await apiClient.post('/user/account-bindings/email/send-code', { email })
}

export async function bindEmailIdentity(payload: {
  email: string
  verify_code: string
  password: string
}): Promise<User> {
  const { data } = await apiClient.post<User>('/user/account-bindings/email', payload)
  return data
}

export type AuthIdentityProvider = 'linuxdo' | 'oidc' | 'wechat'

export async function unbindAuthIdentity(provider: AuthIdentityProvider): Promise<User> {
  const { data } = await apiClient.delete<User>(`/user/account-bindings/${provider}`)
  return data
}

async function startOAuthBinding(provider: AuthIdentityProvider, redirectTo = '/profile'): Promise<void> {
  if (typeof window === 'undefined') {
    return
  }

  await apiClient.post('/auth/oauth/bind-token')

  const apiBase = (import.meta.env.VITE_API_BASE_URL as string | undefined) || '/api/v1'
  const normalizedApiBase = apiBase.replace(/\/$/, '')
  const params = new URLSearchParams({ redirect: redirectTo, intent: 'bind_current_user' })
  window.location.href = `${normalizedApiBase}/auth/oauth/${provider}/bind/start?${params.toString()}`
}

export async function startLinuxDoBinding(redirectTo = '/profile'): Promise<void> {
  await startOAuthBinding('linuxdo', redirectTo)
}

export async function startOIDCBinding(redirectTo = '/profile'): Promise<void> {
  await startOAuthBinding('oidc', redirectTo)
}

export async function startWeChatBinding(redirectTo = '/profile'): Promise<void> {
  await startOAuthBinding('wechat', redirectTo)
}

export const userAPI = {
  getProfile,
  updateProfile,
  changePassword,
  sendNotifyEmailCode,
  verifyNotifyEmail,
  removeNotifyEmail,
  toggleNotifyEmail,
  sendEmailBindingCode,
  bindEmailIdentity,
  unbindAuthIdentity,
  startLinuxDoBinding,
  startOIDCBinding,
  startWeChatBinding
}

export default userAPI
