<template>
  <BaseDialog
    :show="show"
    :title="t('admin.groups.sortOrder')"
    width="normal"
    @close="$emit('close')"
  >
    <div class="space-y-4">
      <p class="text-sm text-gray-500 dark:text-gray-400">
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
          class="flex cursor-grab items-center gap-3 rounded-lg border border-gray-200 bg-white p-3 transition-shadow hover:shadow-md active:cursor-grabbing dark:border-dark-600 dark:bg-dark-700"
        >
          <div class="text-gray-400">
            <Icon name="menu" size="md" />
          </div>
          <div class="flex-1">
            <div class="font-medium text-gray-900 dark:text-white">{{ group.name }}</div>
            <div class="text-xs text-gray-500 dark:text-gray-400">
              <GroupPlatformBadge
                :platform="group.platform"
                :show-icon="false"
                badge-class="gap-1 px-2 py-0.5"
              />
            </div>
          </div>
          <div class="text-sm text-gray-400">
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
