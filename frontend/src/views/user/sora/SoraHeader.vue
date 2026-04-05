<template>
  <header class="sora-header">
    <div class="sora-header-left">
      <router-link :to="dashboardPath" class="sora-back-btn" :title="backTitle">
        <Icon name="chevronLeft" size="md" :stroke-width="2" />
      </router-link>

      <nav class="sora-nav-tabs">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          :class="['sora-nav-tab', activeTab === tab.key && 'active']"
          @click="$emit('update:activeTab', tab.key)"
        >
          {{ tab.label }}
        </button>
      </nav>
    </div>

    <div class="sora-header-right">
      <SoraQuotaBar v-if="quota" :quota="quota" />
      <div v-if="activeTaskCount > 0" class="sora-queue-indicator">
        <span class="sora-queue-dot" :class="{ busy: hasGeneratingTask }"></span>
        <span>{{ activeTaskCount }} {{ queueTasksLabel }}</span>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import type { QuotaInfo } from '@/api/sora'
import Icon from '@/components/icons/Icon.vue'
import SoraQuotaBar from '@/components/sora/SoraQuotaBar.vue'
import type { SoraTabKey, SoraTabOption } from './soraView'

defineProps<{
  tabs: SoraTabOption[]
  activeTab: SoraTabKey
  dashboardPath: string
  backTitle: string
  queueTasksLabel: string
  quota: QuotaInfo | null
  activeTaskCount: number
  hasGeneratingTask: boolean
}>()

defineEmits<{
  'update:activeTab': [value: SoraTabKey]
}>()
</script>
