<template>
  <BaseDialog
    :show="show"
    :title="t('admin.groups.sortOrder')"
    width="normal"
    @close="$emit('close')"
  >
    <div class="space-y-4">
      <p class="group-sort-order-dialog__hint text-sm">
        {{ t('admin.groups.sortOrderHint') }}
      </p>
      <VueDraggable
        v-model="draggableGroups"
        :animation="200"
        class="space-y-2"
      >
        <div
          v-for="group in groups"
          :key="group.id"
          class="group-sort-order-dialog__item flex cursor-grab items-center gap-3 active:cursor-grabbing"
        >
          <div class="group-sort-order-dialog__handle">
            <Icon name="menu" size="md" />
          </div>
          <div class="flex-1">
            <div class="group-sort-order-dialog__name font-medium">{{ group.name }}</div>
            <div class="group-sort-order-dialog__meta text-xs">
              <GroupPlatformBadge
                :platform="group.platform"
                :show-icon="false"
              />
            </div>
          </div>
          <div class="group-sort-order-dialog__id text-sm">
            #{{ group.id }}
          </div>
        </div>
      </VueDraggable>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3 pt-4">
        <button type="button" class="btn btn-secondary" @click="$emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button
          class="btn btn-primary"
          :disabled="submitting"
          @click="$emit('save')"
        >
          <svg
            v-if="submitting"
            class="-ml-1 mr-2 h-4 w-4 animate-spin"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          {{ submitting ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AdminGroup } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { VueDraggable } from 'vue-draggable-plus'
import GroupPlatformBadge from './GroupPlatformBadge.vue'

const props = defineProps<{
  show: boolean
  groups: AdminGroup[]
  submitting: boolean
}>()

const emit = defineEmits<{
  close: []
  save: []
  'update:groups': [groups: AdminGroup[]]
}>()

const { t } = useI18n()

const draggableGroups = computed({
  get: () => props.groups,
  set: (groups) => emit('update:groups', groups)
})
</script>

<style scoped>
.group-sort-order-dialog__hint,
.group-sort-order-dialog__meta,
.group-sort-order-dialog__id {
  color: var(--theme-page-muted);
}

.group-sort-order-dialog__item {
  border-radius: var(--theme-button-radius);
  padding: var(--theme-group-replace-card-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 88%, transparent);
  background: var(--theme-surface);
  transition: box-shadow 0.2s ease, border-color 0.2s ease, transform 0.2s ease;
}

.group-sort-order-dialog__item:hover {
  border-color: color-mix(in srgb, var(--theme-card-border) 72%, var(--theme-accent));
  box-shadow: var(--theme-card-shadow);
}

.group-sort-order-dialog__handle {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.group-sort-order-dialog__name {
  color: var(--theme-page-text);
}
</style>
