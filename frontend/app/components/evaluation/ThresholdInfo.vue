<script setup lang="ts">
// Penjelasan 3 ambang "Perlu Ditinjau" — agar tanda bisa dipahami.
const open = ref(false)
const ITEMS = [
  { title: 'Median LOS melebihi p75 target', body: 'Pasien rata-rata dirawat lebih lama dari yang wajar.' },
  { title: '≥ 25% episode melebihi p75 LOS', body: 'Banyak pasien yang lama rawatnya jauh di atas target.' },
  { title: 'Porsi severity III naik ≥ 10 poin persen', body: 'Pergeseran ke kasus lebih berat — periksa apakah pengkodean sesuai.' },
]
</script>

<template>
  <div class="rounded-lg border border-default bg-default">
    <button
      type="button"
      class="flex w-full items-center gap-2 px-4 py-3 text-left text-sm"
      :aria-expanded="open"
      @click="open = !open"
    >
      <UIcon name="i-lucide-info" class="size-4 text-primary" />
      <span class="flex-1 font-medium">Kapan CP ditandai "Perlu Ditinjau"?</span>
      <span class="text-xs text-muted">Pembanding: klaim sesudah tanggal CP disahkan</span>
      <UIcon :name="open ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" class="size-4 text-muted" />
    </button>
    <div v-if="open" class="grid gap-3 border-t border-default px-4 py-3 md:grid-cols-3">
      <div v-for="(it, i) in ITEMS" :key="i" class="flex gap-2 text-sm">
        <span class="grid size-5 shrink-0 place-items-center rounded-full bg-warning/15 text-xs font-semibold text-warning">{{ i + 1 }}</span>
        <div>
          <p class="font-medium">{{ it.title }}</p>
          <p class="text-xs text-muted">{{ it.body }}</p>
        </div>
      </div>
      <p class="text-xs text-muted md:col-span-3">
        Klaim sebelum CP disahkan adalah bahan pembentuk target, sehingga tidak dipakai sebagai pembanding.
        Biaya riil rumah sakit tidak dimuat — evaluasi ini tidak menjawab "tercover atau tidak" dalam rupiah.
      </p>
    </div>
  </div>
</template>
