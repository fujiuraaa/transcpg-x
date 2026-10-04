<script setup lang="ts">
// Keamanan sesi: keluarkan akun dari semua perangkat (mis. HP hilang atau
// pernah masuk di komputer bersama).
const { logoutAll } = useAuth()
const toast = useToast()
const confirmOpen = ref(false)
const busy = ref(false)

async function run() {
  busy.value = true
  try {
    await logoutAll()
  }
  catch (err) {
    toast.add({ title: apiErrorMessage(err), color: 'error' })
    busy.value = false
  }
}
</script>

<template>
  <section class="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-default bg-default p-5">
    <div class="min-w-0">
      <h3 class="font-semibold text-highlighted">Sesi masuk</h3>
      <p class="text-sm text-muted">Pernah masuk di komputer bersama atau perangkat hilang? Keluarkan akun dari semua perangkat.</p>
    </div>
    <UButton label="Keluar dari semua perangkat" icon="i-lucide-log-out" color="neutral" variant="outline" @click="confirmOpen = true" />

    <UModal v-model:open="confirmOpen" title="Keluar dari semua perangkat?" description="Termasuk perangkat ini — Anda perlu masuk lagi.">
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton label="Batal" color="neutral" variant="ghost" @click="confirmOpen = false" />
          <UButton label="Keluar dari semua" color="error" :loading="busy" @click="run" />
        </div>
      </template>
    </UModal>
  </section>
</template>
