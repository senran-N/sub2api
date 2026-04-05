<template>
  <div class="sora-root">
    <div class="sora-page">
      <SoraNotEnabledState
        v-if="!soraEnabled"
        :title="t('sora.notEnabled')"
        :description="t('sora.notEnabledDesc')"
      />

      <template v-else>
        <SoraHeader
          :tabs="tabs"
          :active-tab="activeTab"
          :dashboard-path="dashboardPath"
          :back-title="t('common.back')"
          :queue-tasks-label="t('sora.queueTasks')"
          :quota="quota"
          :active-task-count="activeTaskCount"
          :has-generating-task="hasGeneratingTask"
          @update:active-tab="activeTab = $event"
        />

        <main class="sora-main">
          <SoraGeneratePage v-show="activeTab === 'generate'" @task-count-change="updateTaskCounts" />
          <SoraLibraryPage v-show="activeTab === 'library'" @switch-to-generate="activeTab = 'generate'" />
        </main>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import SoraGeneratePage from '@/components/sora/SoraGeneratePage.vue'
import SoraLibraryPage from '@/components/sora/SoraLibraryPage.vue'
import SoraHeader from './sora/SoraHeader.vue'
import SoraNotEnabledState from './sora/SoraNotEnabledState.vue'
import { useSoraViewModel } from './sora/soraView'
import './sora/soraTheme.css'

const { t } = useI18n()
const {
  soraEnabled,
  activeTab,
  quota,
  activeTaskCount,
  hasGeneratingTask,
  dashboardPath,
  tabs,
  updateTaskCounts,
  loadQuota
} = useSoraViewModel()

onMounted(async () => {
  await loadQuota()
})
</script>
