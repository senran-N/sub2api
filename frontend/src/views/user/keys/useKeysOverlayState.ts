import { computed, ref, type ComponentPublicInstance, type ComputedRef, type Ref } from 'vue'
import { useClipboard } from '@/composables/useClipboard'
import type { ApiKey } from '@/types'
import type { UserKeyGroupOption } from './keysForm'
import {
  buildKeysGroupDropdownPosition,
  buildKeysMoreMenuPosition,
  type KeysOverlayPosition
} from './keysOverlays'

interface KeysOverlayStateOptions {
  apiKeys: Ref<ApiKey[]>
  groupOptions: ComputedRef<UserKeyGroupOption[]>
  filterGroupOptions: (options: UserKeyGroupOption[], query: string) => UserKeyGroupOption[]
  copiedMessage: string
}

export function useKeysOverlayState(options: KeysOverlayStateOptions) {
  const { copyToClipboard: clipboardCopy } = useClipboard()

  const copiedKeyId = ref<number | null>(null)
  const groupSelectorKeyId = ref<number | null>(null)
  const dropdownPosition = ref<KeysOverlayPosition | null>(null)
  const moreMenuKeyId = ref<number | null>(null)
  const moreMenuPosition = ref<{ top: number; left: number } | null>(null)
  const groupSearchQuery = ref('')
  const groupButtonRefs = ref<Map<number, HTMLElement>>(new Map())
  const moreMenuButtonRefs = ref<Map<number, HTMLElement>>(new Map())

  const moreMenuRow = computed(() => {
    if (moreMenuKeyId.value === null) return null
    return options.apiKeys.value.find((key) => key.id === moreMenuKeyId.value) || null
  })

  const selectedKeyForGroup = computed(() => {
    if (groupSelectorKeyId.value === null) return null
    return options.apiKeys.value.find((key) => key.id === groupSelectorKeyId.value) || null
  })

  const filteredGroupOptions = computed(() =>
    options.filterGroupOptions(options.groupOptions.value, groupSearchQuery.value)
  )

  async function copyToClipboard(text: string, keyId: number) {
    const success = await clipboardCopy(text, options.copiedMessage)
    if (!success) {
      return
    }

    copiedKeyId.value = keyId
    setTimeout(() => {
      copiedKeyId.value = null
    }, 800)
  }

  function setMoreMenuRef(keyId: number, el: Element | ComponentPublicInstance | null) {
    if (el instanceof HTMLElement) {
      moreMenuButtonRefs.value.set(keyId, el)
    } else {
      moreMenuButtonRefs.value.delete(keyId)
    }
  }

  function toggleMoreMenu(keyId: number) {
    if (moreMenuKeyId.value === keyId) {
      closeMoreMenu()
      return
    }

    const buttonEl = moreMenuButtonRefs.value.get(keyId)
    if (buttonEl) {
      moreMenuPosition.value = buildKeysMoreMenuPosition(buttonEl)
    }

    moreMenuKeyId.value = keyId
  }

  function closeMoreMenu() {
    moreMenuKeyId.value = null
    moreMenuPosition.value = null
  }

  function setGroupButtonRef(keyId: number, el: Element | ComponentPublicInstance | null) {
    if (el instanceof HTMLElement) {
      groupButtonRefs.value.set(keyId, el)
    } else {
      groupButtonRefs.value.delete(keyId)
    }
  }

  function openGroupSelector(key: ApiKey) {
    if (groupSelectorKeyId.value === key.id) {
      closeGroupSelector()
      return
    }

    const buttonEl = groupButtonRefs.value.get(key.id)
    if (buttonEl) {
      dropdownPosition.value = buildKeysGroupDropdownPosition(buttonEl)
    }
    groupSelectorKeyId.value = key.id
    groupSearchQuery.value = ''
  }

  function closeGroupSelector() {
    groupSelectorKeyId.value = null
    dropdownPosition.value = null
  }

  return {
    copiedKeyId,
    groupSelectorKeyId,
    dropdownPosition,
    moreMenuKeyId,
    moreMenuPosition,
    groupSearchQuery,
    moreMenuRow,
    selectedKeyForGroup,
    filteredGroupOptions,
    copyToClipboard,
    setMoreMenuRef,
    toggleMoreMenu,
    closeMoreMenu,
    setGroupButtonRef,
    openGroupSelector,
    closeGroupSelector
  }
}
