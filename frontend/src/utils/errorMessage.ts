type ErrorWithMessage = {
  message?: string
  response?: {
    data?: {
      message?: string
      detail?: string
    }
  }
}

export function resolveErrorMessage(error: unknown, fallback: string): string {
  if (typeof error === 'string' && error.trim().length > 0) {
    return error
  }

  const normalizedError = error as ErrorWithMessage | null

  return (
    normalizedError?.response?.data?.detail ||
    normalizedError?.response?.data?.message ||
    normalizedError?.message ||
    fallback
  )
}
