<template>
  <div class="relative" ref="dropdownRef">
    <button
      class="btn btn-secondary px-2 md:px-3"
      :title="t('admin.users.columnSettings')"
      @click="toggleDropdown"
    >
      <Icon name="grid" size="sm" class="md:mr-1.5" />
      <span class="hidden md:inline">{{ t('admin.users.columnSettings') }}</span>
    </button>
    <div
      v-if="showDropdown"
      class="menu-panel right-0 top-full max-h-80 w-48 overflow-y-auto"
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
          class="text-primary-500"
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
