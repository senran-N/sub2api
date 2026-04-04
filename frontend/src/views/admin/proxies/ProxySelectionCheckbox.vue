<template>
  <input
    type="checkbox"
    class="h-4 w-4 cursor-pointer rounded border-gray-300 text-primary-600 focus:ring-primary-500"
    :checked="checked"
    @click.stop="stopPropagation"
    @change="emit('change', $event)"
  />
</template>

<script setup lang="ts">
defineProps<{
  checked: boolean
}>()

const emit = defineEmits<{
  change: [event: Event]
}>()

const stopPropagation = (event: Event) => {
  event.stopPropagation()
}
</script>
