<template>
  <div class="setup-stepper">
    <div class="flex items-center justify-center">
      <template v-for="(step, index) in steps" :key="step.id">
        <div class="setup-stepper__group">
          <div
            :class="[
              'setup-stepper__node',
              currentStep > index
                ? 'setup-stepper__node--done'
                : currentStep === index
                  ? 'setup-stepper__node--active'
                  : 'setup-stepper__node--idle'
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
            class="setup-stepper__label"
            :class="
              currentStep >= index
                ? 'setup-stepper__label--active'
                : 'setup-stepper__label--idle'
            "
          >
            {{ step.title }}
          </span>
        </div>
        <div
          v-if="index < steps.length - 1"
          class="setup-stepper__connector"
          :class="currentStep > index ? 'setup-stepper__connector--active' : 'setup-stepper__connector--idle'"
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
