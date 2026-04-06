<template>
  <div class="sticky top-0 z-10 overflow-x-auto settings-tabs-scroll">
    <nav class="settings-tabs">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        type="button"
        :class="['settings-tab', activeTab === tab.key && 'settings-tab-active']"
        @click="emit('update:activeTab', tab.key)"
      >
        <span class="settings-tab-icon">
          <Icon :name="tab.icon as any" size="sm" />
        </span>
        <span>{{ t(`admin.settings.tabs.${tab.key}`) }}</span>
      </button>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  activeTab: string
  tabs: Array<{ key: string; icon: string }>
}>()

const emit = defineEmits<{
  'update:activeTab': [value: string]
}>()

const { t } = useI18n()
</script>

<style scoped>
.settings-tabs-scroll {
  scrollbar-width: thin;
  scrollbar-color: transparent transparent;
}
.settings-tabs-scroll:hover {
  scrollbar-color: color-mix(in srgb, var(--theme-page-muted) 24%, transparent) transparent;
}
.settings-tabs-scroll::-webkit-scrollbar {
  height: 3px;
}
.settings-tabs-scroll::-webkit-scrollbar-track {
  background: transparent;
}
.settings-tabs-scroll::-webkit-scrollbar-thumb {
  background: transparent;
  border-radius: 3px;
}
.settings-tabs-scroll:hover::-webkit-scrollbar-thumb {
  background: color-mix(in srgb, var(--theme-page-muted) 24%, transparent);
}

.settings-tabs {
  @apply inline-flex min-w-full border backdrop-blur-sm;
  gap: var(--theme-settings-tabs-nav-gap);
  border-radius: var(--theme-settings-tabs-nav-radius);
  padding: var(--theme-settings-tabs-nav-padding);
  border-color: color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  background: color-mix(in srgb, var(--theme-surface) 80%, transparent);
  box-shadow: var(--theme-card-shadow);
}

@media (min-width: 640px) {
  .settings-tabs {
    @apply flex;
  }
}

.settings-tab {
  @apply relative flex flex-1 items-center justify-center gap-1.5
         whitespace-nowrap
         text-sm font-medium
         transition-all duration-200 ease-out;
  border-radius: var(--theme-settings-tab-radius);
  padding: var(--theme-settings-tab-padding-y) var(--theme-settings-tab-padding-x);
  color: var(--theme-page-muted);
}

.settings-tab:hover:not(.settings-tab-active) {
  color: var(--theme-page-text);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.settings-tab-active {
  color: var(--theme-accent);
  background: linear-gradient(
    135deg,
    color-mix(in srgb, var(--theme-accent-soft) 92%, var(--theme-surface)),
    color-mix(in srgb, var(--theme-accent-soft) 60%, var(--theme-surface))
  );
  box-shadow: 0 1px 2px color-mix(in srgb, var(--theme-accent) 14%, transparent);
}

.settings-tab-icon {
  @apply flex h-6 w-6 items-center justify-center
         transition-all duration-200;
  border-radius: var(--theme-settings-tab-icon-radius);
}

.settings-tab-active .settings-tab-icon {
  background: color-mix(in srgb, var(--theme-accent-soft) 88%, var(--theme-surface));
  color: var(--theme-accent);
}
</style>
