<template>
  <input
    type="checkbox"
    class="theme-checkbox cursor-pointer"
    :checked="checked"
    @click.stop="stopPropagation"
    @change="handleChange"
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

const handleChange = (event: Event) => {
  emit('change', event)
}
</script>
