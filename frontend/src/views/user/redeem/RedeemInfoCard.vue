<template>
  <div
    class="card border-primary-200 bg-primary-50 dark:border-primary-800/50 dark:bg-primary-900/20"
  >
    <div class="p-6">
      <div class="flex items-start gap-4">
        <div
          class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-primary-100 dark:bg-primary-900/30"
        >
          <Icon name="infoCircle" size="md" class="text-primary-600 dark:text-primary-400" />
        </div>
        <div class="flex-1">
          <h3 class="text-sm font-semibold text-primary-800 dark:text-primary-300">
            {{ aboutCodesLabel }}
          </h3>
          <ul
            class="mt-2 list-inside list-disc space-y-1 text-sm text-primary-700 dark:text-primary-400"
          >
            <li>{{ codeRule1 }}</li>
            <li>{{ codeRule2 }}</li>
            <li>
              {{ codeRule3 }}
              <span
                v-if="contactInfo"
                class="ml-1.5 inline-flex items-center rounded-md bg-primary-200/50 px-2 py-0.5 text-xs font-medium text-primary-800 dark:bg-primary-800/40 dark:text-primary-200"
              >
                {{ contactInfo }}
              </span>
            </li>
            <li>{{ codeRule4 }}</li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  aboutCodesLabel: string
  codeRule1: string
  codeRule2: string
  codeRule3: string
  codeRule4: string
  contactInfo: string
}>()
</script>
