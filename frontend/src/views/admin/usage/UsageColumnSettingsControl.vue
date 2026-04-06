<template>
  <div class="relative" ref="dropdownRef">
    <button
      class="usage-column-settings-control__trigger btn btn-secondary"
      :title="t('admin.users.columnSettings')"
      @click="toggleDropdown"
    >
      <Icon name="grid" size="sm" class="md:mr-1.5" />
      <span class="hidden md:inline">{{ t('admin.users.columnSettings') }}</span>
    </button>
    <div
      v-if="showDropdown"
      class="usage-column-settings-control__dropdown menu-panel right-0 top-full overflow-y-auto"
    >
      <button
        v-for="column in toggleableColumns"
        :key="column.key"
        class="menu-item"
        @click="emit('toggle-column', column.key)"
      >
        <span>{{ column.label }}</span>
        <Icon
          v-if="isColumnVisible(column.key)"
          name="check"
          size="sm"
          class="usage-column-settings-control__check"
          :stroke-width="2"
        />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Column } from '@/components/common/types'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  toggleableColumns: Column[]
  isColumnVisible: (key: string) => boolean
}>()

const emit = defineEmits<{
  'toggle-column': [key: string]
}>()

const { t } = useI18n()

const showDropdown = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

const toggleDropdown = () => {
  showDropdown.value = !showDropdown.value
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement | null
  if (!target) {
    return
  }

  if (dropdownRef.value && !dropdownRef.value.contains(target)) {
    showDropdown.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.usage-column-settings-control__dropdown {
  width: var(--theme-settings-menu-width-sm);
  max-height: var(--theme-settings-menu-max-height);
}

.usage-column-settings-control__trigger {
  padding-inline: var(--theme-settings-code-padding-x);
}

.usage-column-settings-control__check {
  color: var(--theme-accent);
}

@media (min-width: 768px) {
  .usage-column-settings-control__trigger {
    padding-inline: var(--theme-settings-action-padding-x);
  }
}
</style>
