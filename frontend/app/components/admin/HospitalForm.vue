<script setup lang="ts">
// Formulir Setup Rumah Sakit — 3 kelompok isian. Kombinasi regional + tipe +
// kepemilikan + kelas rawat menentukan tarif INA-CBG di seluruh aplikasi.
import type { Hospital } from '~/types/api'

const form = defineModel<Hospital>({ required: true })
defineProps<{ errors: Partial<Record<keyof Hospital, string>> }>()

const TYPES = [
  { label: 'Tipe A', value: 'A' }, { label: 'Tipe B', value: 'B' }, { label: 'Tipe C', value: 'C' }, { label: 'Tipe D', value: 'D' },
  { label: 'Klinik Utama', value: 'KLINIK_UTAMA' }, { label: 'Klinik Pratama', value: 'KLINIK_PRATAMA' }, { label: 'Puskesmas', value: 'PUSKESMAS' },
]
const OWNERSHIP = [{ label: 'Pemerintah', value: 'PEMERINTAH' }, { label: 'Swasta', value: 'SWASTA' }]
const REGIONAL = [1, 2, 3, 4, 5].map(r => ({ label: `Regional ${r}`, value: r }))
const CLASSES = [1, 2, 3].map(c => ({ label: `Kelas ${c}`, value: c }))
const ACCREDITATION = ['Paripurna', 'Utama', 'Madya', 'Dasar', 'Belum terakreditasi'].map(a => ({ label: a, value: a }))

// Bidang teks opsional: string kosong disimpan sebagai null.
function text(key: 'bpjs_provider_code' | 'city' | 'province' | 'address' | 'phone' | 'email') {
  return computed({
    get: () => form.value[key] ?? '',
    set: (v: string) => { form.value[key] = v.trim() === '' ? null : v },
  })
}
function num(key: 'bed_capacity' | 'icu_capacity') {
  return computed({
    get: () => form.value[key] ?? undefined,
    set: (v: number | undefined | null) => { form.value[key] = v == null || Number.isNaN(v) ? null : v },
  })
}
const city = text('city')
const province = text('province')
const address = text('address')
const provider = text('bpjs_provider_code')
const phone = text('phone')
const email = text('email')
const beds = num('bed_capacity')
const icu = num('icu_capacity')

const typeLabel = computed(() => TYPES.find(t => t.value === form.value.hospital_type)?.label ?? form.value.hospital_type)
</script>

<template>
  <div class="space-y-6">
    <!-- Identitas -->
    <section class="rounded-xl border border-default bg-default">
      <header class="flex items-center gap-2 border-b border-default px-5 py-3">
        <UIcon name="i-lucide-hospital" class="size-5 text-primary" />
        <div>
          <h3 class="font-semibold">Identitas rumah sakit</h3>
          <p class="text-xs text-muted">Nama dan lokasi yang tampil di pemilih profil RS.</p>
        </div>
      </header>
      <div class="grid gap-4 p-5 md:grid-cols-2">
        <UFormField label="Nama rumah sakit" required :error="errors.name" class="md:col-span-2">
          <UInput v-model="form.name" class="w-full" placeholder="Mis. RSUD Kota Contoh" />
        </UFormField>
        <UFormField label="Tipe RS" required :error="errors.hospital_type">
          <USelect v-model="form.hospital_type" :items="TYPES" class="w-full" />
        </UFormField>
        <UFormField label="Kepemilikan" required :error="errors.ownership">
          <URadioGroup v-model="form.ownership" :items="OWNERSHIP" orientation="horizontal" variant="table" />
        </UFormField>
        <UFormField label="Kota/Kabupaten">
          <UInput v-model="city" class="w-full" />
        </UFormField>
        <UFormField label="Provinsi">
          <UInput v-model="province" class="w-full" />
        </UFormField>
        <UFormField label="Alamat" class="md:col-span-2">
          <UTextarea v-model="address" :rows="2" class="w-full" />
        </UFormField>
      </div>
    </section>

    <!-- BPJS & INA-CBG -->
    <section class="rounded-xl border border-default bg-default">
      <header class="flex items-center gap-2 border-b border-default px-5 py-3">
        <UIcon name="i-lucide-wallet" class="size-5 text-primary" />
        <div>
          <h3 class="font-semibold">Konfigurasi BPJS & INA-CBG</h3>
          <p class="text-xs text-muted">Menentukan tarif INA-CBG di CP Library, Detail CP, CP Aktif, dan Evaluasi.</p>
        </div>
      </header>
      <div class="grid gap-4 p-5 md:grid-cols-2">
        <UFormField label="Regional BPJS" required :error="errors.bpjs_regional" description="Sesuai Permenkes tarif INA-CBG yang berlaku.">
          <USelect v-model="form.bpjs_regional" :items="REGIONAL" class="w-full" />
        </UFormField>
        <UFormField label="Kelas rawat tertinggi" required :error="errors.max_care_class">
          <USelect v-model="form.max_care_class" :items="CLASSES" class="w-full" />
        </UFormField>
        <UFormField label="Kode provider BPJS">
          <UInput v-model="provider" class="w-full font-mono" placeholder="Mis. 0123R001" />
        </UFormField>
        <UFormField label="Akreditasi">
          <USelect
            :model-value="form.accreditation ?? undefined"
            :items="ACCREDITATION"
            placeholder="Pilih status akreditasi"
            class="w-full"
            @update:model-value="(v) => form.accreditation = (v as string) ?? null"
          />
        </UFormField>
        <div class="bg-brand-soft flex flex-wrap items-center gap-2 rounded-lg border border-primary/15 px-4 py-3 text-sm md:col-span-2">
          <UIcon name="i-lucide-calculator" class="size-4 text-primary" />
          <span class="text-muted">Kombinasi tarif:</span>
          <UBadge :label="`Regional ${form.bpjs_regional}`" color="neutral" variant="outline" />
          <UBadge :label="typeLabel" color="neutral" variant="outline" />
          <UBadge :label="form.ownership === 'PEMERINTAH' ? 'Pemerintah' : 'Swasta'" color="neutral" variant="outline" />
          <span class="text-muted">× kelas rawat 1–3 × severity I–III</span>
        </div>
      </div>
    </section>

    <!-- Kontak & kapasitas -->
    <section class="rounded-xl border border-default bg-default">
      <header class="flex items-center gap-2 border-b border-default px-5 py-3">
        <UIcon name="i-lucide-contact" class="size-5 text-primary" />
        <h3 class="font-semibold">Kontak & kapasitas</h3>
      </header>
      <div class="grid gap-4 p-5 md:grid-cols-2 xl:grid-cols-4">
        <UFormField label="Telepon">
          <UInput v-model="phone" type="tel" class="w-full" />
        </UFormField>
        <UFormField label="Email RS" :error="errors.email">
          <UInput v-model="email" type="email" class="w-full" />
        </UFormField>
        <UFormField label="Kapasitas tempat tidur" :error="errors.bed_capacity">
          <UInputNumber v-model="beds" :min="0" class="w-full" />
        </UFormField>
        <UFormField label="Kapasitas ICU" :error="errors.icu_capacity">
          <UInputNumber v-model="icu" :min="0" class="w-full" />
        </UFormField>
      </div>
    </section>
  </div>
</template>
