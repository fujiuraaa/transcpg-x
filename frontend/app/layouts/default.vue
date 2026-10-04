<script setup lang="ts">
// Kerangka halaman setelah login: bilah samping + bilah atas + konten.
// Layar < lg: bilah samping disembunyikan, diganti laci navigasi (tombol ☰).
const { load } = useHospitalProfile()
const collapsed = useCookie<boolean>('transcpg_sidebar_collapsed', { default: () => false })
const navOpen = ref(false)

onMounted(load)
</script>

<template>
  <div class="flex min-h-screen bg-default">
    <LayoutAppSidebar v-model:collapsed="collapsed" class="hidden lg:flex" />
    <USlideover v-model:open="navOpen" side="left" :ui="{ content: 'max-w-72' }" title="Menu navigasi" :close="false">
      <template #content>
        <LayoutAppSidebar drawer @navigate="navOpen = false" />
      </template>
    </USlideover>
    <div class="flex min-w-0 flex-1 flex-col">
      <LayoutAppTopbar @open-nav="navOpen = true" />
      <main class="flex-1 p-4 sm:p-6">
        <slot />
      </main>
    </div>
  </div>
</template>
