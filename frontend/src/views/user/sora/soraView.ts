import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import soraAPI, { type QuotaInfo } from '@/api/sora'
import { useAppStore, useAuthStore } from '@/stores'

export type SoraTabKey = 'generate' | 'library'

export interface SoraTabOption {
  key: SoraTabKey
  label: string
}

type Translate = (key: string) => string

export function resolveSoraDashboardPath(isAdmin: boolean): string {
  return isAdmin ? '/admin/dashboard' : '/dashboard'
}

export function buildSoraTabs(t: Translate): SoraTabOption[] {
  return [
    { key: 'generate', label: t('sora.tabGenerate') },
    { key: 'library', label: t('sora.tabLibrary') }
  ]
}

export function useSoraViewModel() {
  const { t } = useI18n()
  const authStore = useAuthStore()
  const appStore = useAppStore()

  const soraEnabled = computed(() => appStore.cachedPublicSettings?.sora_client_enabled ?? false)
  const activeTab = ref<SoraTabKey>('generate')
  const quota = ref<QuotaInfo | null>(null)
  const activeTaskCount = ref(0)
  const hasGeneratingTask = ref(false)

  const dashboardPath = computed(() => resolveSoraDashboardPath(authStore.isAdmin))
  const tabs = computed(() => buildSoraTabs(t))

  function updateTaskCounts(counts: { active: number; generating: boolean }) {
    activeTaskCount.value = counts.active
    hasGeneratingTask.value = counts.generating
  }

  async function loadQuota() {
    if (!soraEnabled.value) {
      return
    }

    try {
      quota.value = await soraAPI.getQuota()
    } catch {
      // 配额查询失败不阻塞页面
    }
  }

  return {
    soraEnabled,
    activeTab,
    quota,
    activeTaskCount,
    hasGeneratingTask,
    dashboardPath,
    tabs,
    updateTaskCounts,
    loadQuota
  }
}
