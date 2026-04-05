import type { ResetPasswordRequest } from '@/api/auth'

export interface ResetPasswordFormData {
  password: string
  confirmPassword: string
}

export interface ResetPasswordFormErrors {
  password: string
  confirmPassword: string
}

export interface ResetPasswordRouteState {
  email: string
  token: string
}

type Translate = (key: string) => string

interface ResetPasswordErrorLike {
  message?: string
  response?: {
    data?: {
      detail?: string
      code?: string
    }
  }
}

export function createResetPasswordFormData(): ResetPasswordFormData {
  return {
    password: '',
    confirmPassword: ''
  }
}

export function createResetPasswordFormErrors(): ResetPasswordFormErrors {
  return {
    password: '',
    confirmPassword: ''
  }
}

export function resolveResetPasswordRouteState(query: Record<string, unknown>): ResetPasswordRouteState {
  return {
    email: typeof query.email === 'string' ? query.email : '',
    token: typeof query.token === 'string' ? query.token : ''
  }
}

export function isResetPasswordLinkInvalid(state: ResetPasswordRouteState): boolean {
  return !state.email || !state.token
}

export function validateResetPasswordForm(
  formData: ResetPasswordFormData,
  t: Translate
): ResetPasswordFormErrors {
  const errors = createResetPasswordFormErrors()

  if (!formData.password) {
    errors.password = t('auth.passwordRequired')
  } else if (formData.password.length < 6) {
    errors.password = t('auth.passwordMinLength')
  }

  if (!formData.confirmPassword) {
    errors.confirmPassword = t('auth.confirmPasswordRequired')
  } else if (formData.password !== formData.confirmPassword) {
    errors.confirmPassword = t('auth.passwordsDoNotMatch')
  }

  return errors
}

export function hasResetPasswordFormErrors(errors: ResetPasswordFormErrors): boolean {
  return Object.values(errors).some(Boolean)
}

export function buildResetPasswordSubmitPayload(
  routeState: ResetPasswordRouteState,
  formData: ResetPasswordFormData
): ResetPasswordRequest {
  return {
    email: routeState.email,
    token: routeState.token,
    new_password: formData.password
  }
}

export function resolveResetPasswordErrorMessage(
  error: unknown,
  t: Translate
): string {
  const resetPasswordError = error as ResetPasswordErrorLike | null

  if (resetPasswordError?.response?.data?.code === 'INVALID_RESET_TOKEN') {
    return t('auth.invalidOrExpiredToken')
  }

  return (
    resetPasswordError?.response?.data?.detail ||
    resetPasswordError?.message ||
    t('auth.resetPasswordFailed')
  )
}
