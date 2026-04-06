<template>
  <Teleport to="body">
    <div v-if="show && position && row">
      <div class="fixed inset-0 z-[9998]" @click="emit('close')"></div>
      <div
        class="keys-more-actions-menu fixed z-[9999] overflow-hidden"
        :style="{ top: `${position.top}px`, left: `${position.left}px` }"
      >
        <div class="keys-more-actions-menu__content">
          <button
            v-if="!hideCcsImportButton"
            class="keys-more-actions-menu__button flex w-full items-center gap-2 text-sm"
            @click="handleImport"
          >
            <Icon name="upload" size="sm" class="keys-more-actions-menu__icon keys-more-actions-menu__icon--info" />
            {{ t('keys.importToCcSwitch') }}
          </button>
          <button
            class="keys-more-actions-menu__button flex w-full items-center gap-2 text-sm"
            @click="handleToggleStatus"
          >
            <Icon
              :name="row.status === 'active' ? 'ban' : 'checkCircle'"
              size="sm"
              :class="row.status === 'active' ? 'keys-more-actions-menu__icon keys-more-actions-menu__icon--warning' : 'keys-more-actions-menu__icon keys-more-actions-menu__icon--success'"
            />
            {{ row.status === 'active' ? t('keys.disable') : t('keys.enable') }}
          </button>
          <div class="keys-more-actions-menu__divider my-1 border-t"></div>
          <button
            class="keys-more-actions-menu__button keys-more-actions-menu__button--danger flex w-full items-center gap-2 text-sm"
            @click="handleDelete"
          >
            <Icon name="trash" size="sm" />
            {{ t('common.delete') }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { ApiKey } from '@/types'

defineProps<{
  show: boolean
  position: { top: number; left: number } | null
  row: ApiKey | null
  hideCcsImportButton: boolean
}>()

const emit = defineEmits<{
  close: []
  import: []
  'toggle-status': []
  delete: []
}>()

const { t } = useI18n()

function handleImport() {
  emit('import')
  emit('close')
}

function handleToggleStatus() {
  emit('toggle-status')
  emit('close')
}

function handleDelete() {
  emit('delete')
  emit('close')
}
</script>

<style scoped>
.keys-more-actions-menu {
  width: var(--theme-keys-more-actions-menu-width);
  border-radius: var(--theme-keys-more-actions-menu-radius);
  background: var(--theme-dropdown-bg);
  box-shadow: var(--theme-dropdown-shadow);
  border: 1px solid color-mix(in srgb, var(--theme-dropdown-border) 88%, transparent);
}

.keys-more-actions-menu__content {
  padding-block: var(--theme-keys-more-actions-menu-padding-y);
}

.keys-more-actions-menu__button {
  padding: var(--theme-keys-more-actions-menu-item-padding-y)
    var(--theme-keys-more-actions-menu-item-padding-x);
  color: var(--theme-page-text);
  transition: background-color 0.2s ease, color 0.2s ease;
}

.keys-more-actions-menu__button:hover {
  background: var(--theme-dropdown-item-hover-bg);
}

.keys-more-actions-menu__button--danger {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.keys-more-actions-menu__button--danger:hover {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
}

.keys-more-actions-menu__icon--info {
  color: rgb(var(--theme-info-rgb));
}

.keys-more-actions-menu__icon--warning {
  color: rgb(var(--theme-warning-rgb));
}

.keys-more-actions-menu__icon--success {
  color: rgb(var(--theme-success-rgb));
}

.keys-more-actions-menu__divider {
  border-color: color-mix(in srgb, var(--theme-card-border) 76%, transparent);
}
</style>
