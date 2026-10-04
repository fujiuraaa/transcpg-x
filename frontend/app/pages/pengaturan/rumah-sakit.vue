<script setup lang="ts">
// Pengaturan › Setup Rumah Sakit. Lengkapi sedini mungkin — sebelum diisi,
// tarif memakai nilai bawaan sistem.
import type { Hospital } from '~/types/api'

definePageMeta({ title: 'Setup Rumah Sakit', middleware: 'admin' })

const route = useRoute()
const router = useRouter()
const { user } = useAuth()
const { updateHospital } = useAdminApi()
const { hospitals, load } = useHospitalProfile()
const toast = useToast()

await load()

// Admin RS hanya mengubah RS-nya sendiri; System/Super Admin boleh semua profil.
const canPickAny = computed(() => user.value?.role === 'SYSTEM_ADMIN' || user.value?.role === 'SUPER_ADMIN')
const selectedId = computed(() => Number(route.query.id ?? user.value?.hospital_id ?? hospitals.value[0]?.id))
const original = computed(() => hospitals.value.find(h => h.id === selectedId.value) ?? null)

const form = ref<Hospital | null>(null)
watch(original, (h) => { form.value = h ? structuredClone(toRaw(h)) : null }, { immediate: true })

const dirty = computed(() => !!form.value && !!original.value && JSON.stringify(form.value) !== JSON.stringify(original.value))

// Validasi sisi klien (server memeriksa ulang).
const errors = computed<Partial<Record<keyof Hospital, string>>>(() => {
  const f = form.value
  if (!f) return {}
  const e: Partial<Record<keyof Hospital, string>> = {}
  if (!f.name.trim()) e.name = 'Nama rumah sakit wajib diisi'
  if (f.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(f.email.trim())) e.email = 'Format email tidak valid'
  if (f.bed_capacity != null && f.bed_capacity < 0) e.bed_capacity = 'Tidak boleh negatif'
  if (f.icu_capacity != null && f.icu_capacity < 0) e.icu_capacity = 'Tidak boleh negatif'
  return e
})
const valid = computed(() => Object.keys(errors.value).length === 0)

const busy = ref(false)
async function save() {
  if (!form.value || !valid.value) return
  busy.value = true
  try {
    await updateHospital(form.value.id, form.value)
    await load()
    toast.add({ title: 'Profil rumah sakit disimpan', description: 'Tarif INA-CBG di seluruh halaman mengikuti data terbaru.', color: 'success', icon: 'i-lucide-circle-check' })
  }
  catch (err) {
    toast.add({ title: apiErrorMessage(err), color: 'error' })
  }
  finally {
    busy.value = false
  }
}
function reset() {
  if (original.value) form.value = structuredClone(toRaw(original.value))
}

// Cegah perubahan hilang saat berpindah halaman atau menutup tab.
onBeforeRouteLeave(() => (dirty.value ? window.confirm('Perubahan belum disimpan. Tinggalkan halaman ini?') : true))
// Ganti profil RS di pemilih (query berubah, halaman sama) juga harus dikonfirmasi.
onBeforeRouteUpdate(() => (dirty.value ? window.confirm('Perubahan belum disimpan. Pindah ke profil RS lain?') : true))
const beforeUnload = (e: BeforeUnloadEvent) => { if (dirty.value) e.preventDefault() }
onMounted(() => window.addEventListener('beforeunload', beforeUnload))
onBeforeUnmount(() => window.removeEventListener('beforeunload', beforeUnload))

const isIncomplete = computed(() => !!original.value && (!original.value.bpjs_provider_code || !original.value.accreditation))
</script>

<template>
  <div class="mx-auto max-w-[1100px] space-y-6">
    <header class="bg-brand-soft flex flex-wrap items-end justify-between gap-3 rounded-xl border border-primary/15 px-5 py-4">
      <div>
        <NuxtLink to="/pengaturan" class="mb-1 inline-flex items-center gap-1 text-sm text-muted hover:text-default">
          <UIcon name="i-lucide-arrow-left" class="size-4" /> Pengaturan
        </NuxtLink>
        <h2 class="text-2xl font-semibold text-highlighted">Setup Rumah Sakit</h2>
        <p class="text-sm text-muted">Profil RS menentukan tarif INA-CBG yang ditampilkan di seluruh aplikasi.</p>
      </div>
      <USelect
        v-if="canPickAny && hospitals.length > 1"
        :model-value="selectedId"
        :items="hospitals.map(h => ({ label: h.name, value: h.id }))"
        icon="i-lucide-hospital"
        class="w-80"
        aria-label="Profil yang disunting"
        @update:model-value="(v) => router.replace({ query: { id: String(v) } })"
      />
    </header>

    <UAlert
      v-if="isIncomplete"
      color="warning"
      variant="subtle"
      icon="i-lucide-triangle-alert"
      title="Lengkapi setup sedini mungkin"
      description="Kode provider BPJS dan akreditasi belum diisi. Pastikan regional, tipe, kepemilikan, dan kelas rawat sesuai — tarif di seluruh halaman bergantung pada data ini."
    />

    <UEmpty v-if="!form" icon="i-lucide-hospital" title="Profil rumah sakit tidak ditemukan" variant="outline" />
    <AdminHospitalForm v-else v-model="form" :errors="errors" />

    <!-- Bilah simpan: hanya muncul bila ada perubahan -->
    <Transition
      enter-from-class="translate-y-full opacity-0"
      leave-to-class="translate-y-full opacity-0"
      enter-active-class="transition"
      leave-active-class="transition"
    >
      <div v-if="dirty" class="sticky bottom-3 z-20 rounded-xl border border-default bg-default/95 shadow-lg backdrop-blur">
        <div class="flex items-center justify-between gap-3 px-5 py-3">
          <p class="flex items-center gap-2 text-sm">
            <UIcon name="i-lucide-pencil-line" class="size-4 text-warning" />
            Ada perubahan yang belum disimpan
          </p>
          <div class="flex gap-2">
            <UButton label="Batalkan" color="neutral" variant="ghost" @click="reset" />
            <UButton label="Simpan" icon="i-lucide-save" :loading="busy" :disabled="!valid" @click="save" />
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>
