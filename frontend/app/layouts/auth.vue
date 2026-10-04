<template>
  <div class="relative flex min-h-screen w-full items-center justify-center overflow-hidden bg-default p-4">
    <!-- Cahaya gradasi merah lembut di belakang kartu -->
    <div
      class="pointer-events-none absolute -top-40 left-1/2 size-[36rem] -translate-x-1/2 rounded-full bg-(--ui-color-primary-200)/40 blur-3xl dark:bg-(--ui-color-primary-900)/30"
      aria-hidden="true"
    />
    <div class="relative z-10 flex w-full justify-center">
      <slot />
    </div>
  </div>
</template>
