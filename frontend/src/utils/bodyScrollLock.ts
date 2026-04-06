let bodyScrollLockCount = 0

export function lockBodyScroll(): void {
  if (typeof document === 'undefined') return
  bodyScrollLockCount += 1
  document.body.classList.add('modal-open')
}

export function unlockBodyScroll(): boolean {
  if (typeof document === 'undefined') return false
  bodyScrollLockCount = Math.max(0, bodyScrollLockCount - 1)
  if (bodyScrollLockCount === 0) {
    document.body.classList.remove('modal-open')
    return true
  }
  return false
}
