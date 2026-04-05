<template>
  <div class="space-y-6">
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <div v-for="index in 4" :key="index" class="card p-5">
        <div class="flex items-start gap-4">
          <div class="h-12 w-12 animate-pulse rounded-xl bg-gray-200 dark:bg-dark-700"></div>
          <div class="flex-1 space-y-2">
            <div class="h-4 w-20 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
            <div class="h-7 w-28 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
          </div>
        </div>
      </div>
    </div>

    <div class="card p-6">
      <div class="mb-4 flex items-center justify-between">
        <div class="h-5 w-32 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
        <div class="flex gap-2">
          <div class="h-8 w-24 animate-pulse rounded-lg bg-gray-200 dark:bg-dark-700"></div>
          <div class="h-8 w-24 animate-pulse rounded-lg bg-gray-200 dark:bg-dark-700"></div>
        </div>
      </div>
      <div class="h-64 animate-pulse rounded-xl bg-gray-200 dark:bg-dark-700"></div>
    </div>

    <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
      <div class="card p-6 lg:col-span-2">
        <div class="mb-4 h-5 w-40 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
        <div class="space-y-3">
          <div v-for="index in 4" :key="index" class="flex items-center gap-4">
            <div class="h-4 w-full animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
          </div>
        </div>
      </div>

      <div class="card p-6 lg:col-span-1">
        <div class="mb-4 h-5 w-32 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
        <div class="space-y-3">
          <div v-for="index in 3" :key="index" class="h-10 animate-pulse rounded-lg bg-gray-200 dark:bg-dark-700"></div>
        </div>
      </div>
    </div>
  </div>
</template>
