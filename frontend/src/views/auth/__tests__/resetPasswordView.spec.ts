import { describe, expect, it } from 'vitest'
import {
  buildResetPasswordSubmitPayload,
  hasResetPasswordFormErrors,
  isResetPasswordLinkInvalid,
  resolveResetPasswordErrorMessage,
  resolveResetPasswordRouteState,
  validateResetPasswordForm
} from '../reset-password/resetPasswordView'

const t = (key: string) => key

describe('resetPasswordView', () => {
  it('resolves route state from query params', () => {
    expect(
      resolveResetPasswordRouteState({
        email: 'user@example.com',
        token: 'reset-token'
      })
    ).toEqual({
      email: 'user@example.com',
      token: 'reset-token'
    })

    expect(
      resolveResetPasswordRouteState({
        email: ['wrong'],
        token: null
      })
    ).toEqual({
      email: '',
      token: ''
    })
  })

  it('detects invalid reset links', () => {
    expect(isResetPasswordLinkInvalid({ email: '', token: 'token' })).toBe(true)
    expect(isResetPasswordLinkInvalid({ email: 'user@example.com', token: 'token' })).toBe(
      false
    )
  })

  it('validates password fields without layered fallbacks', () => {
    const errors = validateResetPasswordForm(
      {
        password: '123',
        confirmPassword: '456'
      },
      t
    )

    expect(errors).toEqual({
      password: 'auth.passwordMinLength',
      confirmPassword: 'auth.passwordsDoNotMatch'
    })
    expect(hasResetPasswordFormErrors(errors)).toBe(true)
  })

  it('builds the reset password submit payload explicitly', () => {
    expect(
      buildResetPasswordSubmitPayload(
        { email: 'user@example.com', token: 'reset-token' },
        { password: 'password123', confirmPassword: 'password123' }
      )
    ).toEqual({
      email: 'user@example.com',
      token: 'reset-token',
      new_password: 'password123'
    })
  })

  it('maps invalid-token and generic reset errors deterministically', () => {
    expect(
      resolveResetPasswordErrorMessage(
        { response: { data: { code: 'INVALID_RESET_TOKEN' } } },
        t
      )
    ).toBe('auth.invalidOrExpiredToken')

    expect(
      resolveResetPasswordErrorMessage(
        { response: { data: { detail: 'detail' } } },
        t
      )
    ).toBe('detail')
  })
})
