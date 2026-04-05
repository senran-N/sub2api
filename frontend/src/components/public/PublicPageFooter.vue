<template>
  <footer class="relative z-10 border-t border-gray-200/50 px-6 py-8 dark:border-dark-800/50">
    <div
      class="mx-auto flex max-w-6xl flex-col items-center justify-center gap-4 text-center sm:flex-row sm:text-left"
    >
      <p class="text-sm text-gray-500 dark:text-dark-400">
        &copy; {{ currentYear }} {{ siteName }}. {{ rightsLabel }}
      </p>
      <div class="flex items-center gap-4">
        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="text-sm text-gray-500 transition-colors hover:text-gray-700 dark:text-dark-400 dark:hover:text-white"
        >
          {{ docsLabel }}
        </a>
        <a
          :href="githubUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="text-sm text-gray-500 transition-colors hover:text-gray-700 dark:text-dark-400 dark:hover:text-white"
        >
          GitHub
        </a>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
defineProps<{
  currentYear: number
  siteName: string
  docUrl?: string
  githubUrl: string
  rightsLabel: string
  docsLabel: string
}>()
</script>
