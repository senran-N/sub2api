<template>
  <input
    type="checkbox"
    class="theme-checkbox cursor-pointer"
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
