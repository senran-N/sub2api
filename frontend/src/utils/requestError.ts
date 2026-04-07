interface ErrorResponsePayload {
  detail?: string
  message?: string
}

interface RequestErrorLike {
  name?: string
  code?: string
  message?: string
  response?: {
    status?: number
    data?: ErrorResponsePayload
  }
}

export function isAbortError(error: unknown): boolean {
  if (!error || typeof error !== 'object') {
    return false
  }

  const maybeError = error as RequestErrorLike
  return (
    maybeError.name === 'AbortError' ||
    maybeError.name === 'CanceledError' ||
    maybeError.code === 'ERR_CANCELED'
  )
}

export function resolveRequestErrorMessage(error: unknown, fallback: string): string {
  if (!error || typeof error !== 'object') {
    return fallback
  }

  const maybeError = error as RequestErrorLike
  const detail = maybeError.response?.data?.detail
  if (typeof detail === 'string' && detail.trim()) {
    return detail
  }

  const responseMessage = maybeError.response?.data?.message
  if (typeof responseMessage === 'string' && responseMessage.trim()) {
    return responseMessage
  }

  if (typeof maybeError.message === 'string' && maybeError.message.trim()) {
    return maybeError.message
  }

  return fallback
}

export function hasResponseStatus(error: unknown, status: number): boolean {
  if (!error || typeof error !== 'object') {
    return false
  }

  const maybeError = error as RequestErrorLike
  return maybeError.response?.status === status
}
