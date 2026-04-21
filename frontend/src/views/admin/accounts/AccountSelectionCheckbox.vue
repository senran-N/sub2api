<template>
  <input
    :id="id"
    :name="name"
    type="checkbox"
    class="theme-checkbox cursor-pointer"
    :checked="checked"
    :aria-label="ariaLabel"
    @click.stop="stopPropagation"
    @change="handleChange"
  />
</template>

<script setup lang="ts">
defineProps<{
  checked: boolean
  id?: string
  name?: string
  ariaLabel?: string
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
