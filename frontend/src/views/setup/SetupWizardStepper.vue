<template>
  <div class="mb-8">
    <div class="flex items-center justify-center">
      <template v-for="(step, index) in steps" :key="step.id">
        <div class="flex items-center">
          <div
            :class="[
              'flex h-10 w-10 items-center justify-center rounded-full text-sm font-semibold transition-all',
              currentStep > index
                ? 'bg-primary-500 text-white'
                : currentStep === index
                  ? 'bg-primary-500 text-white ring-4 ring-primary-100 dark:ring-primary-900'
                  : 'bg-gray-200 text-gray-500 dark:bg-dark-700 dark:text-dark-400'
            ]"
          >
            <Icon
              v-if="currentStep > index"
              name="check"
              size="md"
              :stroke-width="2"
            />
            <span v-else>{{ index + 1 }}</span>
          </div>
          <span
            class="ml-2 text-sm font-medium"
            :class="
              currentStep >= index
                ? 'text-gray-900 dark:text-white'
                : 'text-gray-400 dark:text-dark-500'
            "
          >
            {{ step.title }}
          </span>
        </div>
        <div
          v-if="index < steps.length - 1"
          class="mx-3 h-0.5 w-12"
          :class="currentStep > index ? 'bg-primary-500' : 'bg-gray-200 dark:bg-dark-700'"
        ></div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'
import type { SetupStep } from './setupWizardView'

defineProps<{
  currentStep: number
  steps: SetupStep[]
}>()
</script>
