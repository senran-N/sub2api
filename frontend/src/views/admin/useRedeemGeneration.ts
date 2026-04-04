import { computed, reactive, ref, watch } from 'vue'
import { adminAPI } from '@/api/admin'
import type { Group, RedeemCode } from '@/types'
import {
  buildGeneratedRedeemCodesText,
  buildRedeemSubscriptionGroupOptions,
  createDefaultRedeemGenerationForm,
  getGeneratedRedeemTextareaHeight,
  resetRedeemGenerationSubscriptionFields,
  syncRedeemGenerationFormValue
} from './redeemForm'

interface RedeemGenerationOptions {
  t: (key: string, params?: Record<string, unknown>) => string
  showError: (message: string) => void
  copyToClipboard: (text: string, successMessage?: string) => Promise<boolean>
  reloadCodes: () => Promise<void> | void
}

function downloadGeneratedCodesFile(content: string, filename: string): void {
  const blob = new Blob([content], { type: 'text/plain' })
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

export function useRedeemGeneration(options: RedeemGenerationOptions) {
  const showGenerateDialog = ref(false)
  const showResultDialog = ref(false)
  const generating = ref(false)
  const copiedAll = ref(false)
  const generatedCodes = ref<RedeemCode[]>([])
  const subscriptionGroups = ref<Group[]>([])
  const generateForm = reactive(createDefaultRedeemGenerationForm())

  const subscriptionGroupOptions = computed(() =>
    buildRedeemSubscriptionGroupOptions(subscriptionGroups.value)
  )
  const generatedCodesText = computed(() => buildGeneratedRedeemCodesText(generatedCodes.value))
  const textareaHeight = computed(() => `${getGeneratedRedeemTextareaHeight(generatedCodes.value.length)}px`)

  let copiedAllTimeout: ReturnType<typeof setTimeout> | null = null

  watch(
    () => generateForm.type,
    () => {
      syncRedeemGenerationFormValue(generateForm)
    }
  )

  const loadSubscriptionGroups = async () => {
    try {
      subscriptionGroups.value = await adminAPI.groups.getAll()
    } catch (error) {
      console.error('Error loading subscription groups:', error)
    }
  }

  const handleGenerateCodes = async () => {
    if (generateForm.type === 'subscription' && !generateForm.group_id) {
      options.showError(options.t('admin.redeem.groupRequired'))
      return
    }

    generating.value = true
    try {
      generatedCodes.value = await adminAPI.redeem.generate(
        generateForm.count,
        generateForm.type,
        generateForm.value,
        generateForm.type === 'subscription' ? generateForm.group_id : undefined,
        generateForm.type === 'subscription' ? generateForm.validity_days : undefined
      )
      showGenerateDialog.value = false
      showResultDialog.value = true
      resetRedeemGenerationSubscriptionFields(generateForm)
      await options.reloadCodes()
    } catch (error: any) {
      options.showError(error.response?.data?.detail || options.t('admin.redeem.failedToGenerate'))
      console.error('Error generating codes:', error)
    } finally {
      generating.value = false
    }
  }

  const closeResultDialog = () => {
    showResultDialog.value = false
    generatedCodes.value = []
    copiedAll.value = false
    if (copiedAllTimeout) {
      clearTimeout(copiedAllTimeout)
      copiedAllTimeout = null
    }
  }

  const copyGeneratedCodes = async () => {
    const success = await options.copyToClipboard(
      generatedCodesText.value,
      options.t('admin.redeem.copied')
    )
    if (!success) {
      return
    }

    copiedAll.value = true
    if (copiedAllTimeout) {
      clearTimeout(copiedAllTimeout)
    }
    copiedAllTimeout = setTimeout(() => {
      copiedAll.value = false
    }, 2000)
  }

  const downloadGeneratedCodes = () => {
    const date = new Date().toISOString().split('T')[0]
    downloadGeneratedCodesFile(generatedCodesText.value, `redeem-codes-${date}.txt`)
  }

  const dispose = () => {
    if (copiedAllTimeout) {
      clearTimeout(copiedAllTimeout)
    }
  }

  return {
    showGenerateDialog,
    showResultDialog,
    generating,
    copiedAll,
    generatedCodes,
    subscriptionGroups,
    generateForm,
    subscriptionGroupOptions,
    generatedCodesText,
    textareaHeight,
    loadSubscriptionGroups,
    handleGenerateCodes,
    closeResultDialog,
    copyGeneratedCodes,
    downloadGeneratedCodes,
    dispose
  }
}
