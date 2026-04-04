<template>
  <div class="space-y-6">
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 lg:grid-cols-4">
      <div v-for="item in 4" :key="`core-${item}`" class="card p-4">
        <div class="flex items-center gap-3">
          <div class="h-10 w-10 animate-pulse rounded-lg bg-gray-200 dark:bg-dark-700"></div>
          <div class="flex-1 space-y-2">
            <div class="h-3 w-16 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
            <div class="h-6 w-24 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
          </div>
        </div>
      </div>
    </div>
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 lg:grid-cols-4">
      <div v-for="item in 4" :key="`token-${item}`" class="card p-4">
        <div class="flex items-center gap-3">
          <div class="h-10 w-10 animate-pulse rounded-lg bg-gray-200 dark:bg-dark-700"></div>
          <div class="flex-1 space-y-2">
            <div class="h-3 w-16 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
            <div class="h-6 w-24 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
          </div>
        </div>
      </div>
    </div>
    <div class="card p-6">
      <div class="mb-4 h-5 w-40 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
      <div class="h-72 animate-pulse rounded-xl bg-gray-200 dark:bg-dark-700"></div>
    </div>
  </div>
</template>
