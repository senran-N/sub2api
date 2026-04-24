export interface AccountModalPlatformRequestContext<TPlatform extends string> {
  platform: TPlatform
  requestSequence: number
}

export function createPlatformRequestGuard<TPlatform extends string>(
  isActivePlatform: (platform: TPlatform) => boolean
) {
  let requestSequence = 0

  return {
    begin(platform: TPlatform): AccountModalPlatformRequestContext<TPlatform> {
      requestSequence += 1
      return { platform, requestSequence }
    },
    invalidate() {
      requestSequence += 1
    },
    isActive(requestContext: AccountModalPlatformRequestContext<TPlatform>) {
      return requestContext.requestSequence === requestSequence &&
        isActivePlatform(requestContext.platform)
    },
    isCurrentSequence(requestContext: AccountModalPlatformRequestContext<TPlatform>) {
      return requestContext.requestSequence === requestSequence
    },
    currentSequence() {
      return requestSequence
    }
  }
}

export function createSequenceRequestGuard(isActiveScope: () => boolean) {
  let requestSequence = 0

  return {
    begin() {
      requestSequence += 1
      return requestSequence
    },
    invalidate() {
      requestSequence += 1
    },
    isActive(sequence: number) {
      return sequence === requestSequence && isActiveScope()
    }
  }
}
