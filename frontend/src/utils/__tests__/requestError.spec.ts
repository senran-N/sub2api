import { describe, expect, it } from 'vitest'
import { hasResponseStatus, isAbortError, resolveRequestErrorMessage } from '../requestError'

describe('requestError utils', () => {
  it('detects supported abort error shapes', () => {
    expect(isAbortError({ name: 'AbortError' })).toBe(true)
    expect(isAbortError({ name: 'CanceledError' })).toBe(true)
    expect(isAbortError({ code: 'ERR_CANCELED' })).toBe(true)
    expect(isAbortError(new Error('boom'))).toBe(false)
  })

  it('resolves detail, response message, error message, then fallback', () => {
    expect(
      resolveRequestErrorMessage(
        { response: { data: { detail: 'detail-message', message: 'response-message' } } },
        'fallback'
      )
    ).toBe('detail-message')

    expect(
      resolveRequestErrorMessage(
        { response: { data: { message: 'response-message' } } },
        'fallback'
      )
    ).toBe('response-message')

    expect(resolveRequestErrorMessage(new Error('plain-error'), 'fallback')).toBe('plain-error')
    expect(resolveRequestErrorMessage(null, 'fallback')).toBe('fallback')
  })

  it('matches response status codes safely', () => {
    expect(hasResponseStatus({ response: { status: 409 } }, 409)).toBe(true)
    expect(hasResponseStatus({ response: { status: 400 } }, 409)).toBe(false)
    expect(hasResponseStatus(null, 409)).toBe(false)
  })
})
