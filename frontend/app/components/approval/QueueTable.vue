<script setup lang="ts">
// Tabel antrean pengesahan. Tidak ada tombol setuju/tolak di sini —
// "Tinjau"/"Buka" membawa ke Halaman Detail CP agar isi CP dibaca dulu.
import type { ApprovalRow } from '~/types/api'

defineProps<{ rows: ApprovalRow[], showReadiness: boolean }>()
</script>

<template>
  <div class="table-wrap">
    <table class="data-table">
      <thead>
        <tr>
          <th>Clinical Pathway</th>
          <th class="w-64">Tahap</th>
          <th class="w-36">{{ showReadiness ? 'Di tahap ini' : 'Disahkan' }}</th>
          <th v-if="showReadiness" class="w-72">Syarat aktif</th>
          <th class="w-28"><span class="sr-only">Aksi</span></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in rows" :key="r.id" :class="r.can_act ? 'bg-primary/[0.03]' : ''">
          <td>
            <p class="flex items-center gap-2 text-xs text-muted">
              <span class="font-mono font-medium text-default">{{ r.cbg_code }}</span>
              <span aria-hidden="true">·</span>
              <span class="truncate">{{ r.mdc_name }}</span>
            </p>
            <NuxtLink :to="`/library/${r.cbg_code}`" class="font-medium text-highlighted hover:text-primary">{{ r.name }}</NuxtLink>
            <p class="text-xs text-muted">{{ formatNumber(r.episodes) }} episode rawat inap</p>
          </td>
          <td>
            <PathwayStageBadge :stage="r.stage" />
            <p v-if="r.stage !== 'AKTIF'" class="mt-1 flex items-center gap-1 text-xs" :class="r.can_act ? 'font-medium text-primary' : 'text-muted'">
              <UIcon :name="r.can_act ? 'i-lucide-hand' : 'i-lucide-hourglass'" class="size-3.5" />
              {{ r.can_act ? 'Menunggu tindakan Anda' : `Menunggu ${r.waiting_on}` }}
            </p>
          </td>
          <td>
            <p class="tabular-nums">{{ sinceLabel(r.stage_changed_at) }}</p>
            <p class="text-xs text-muted">{{ formatDate(r.stage_changed_at) }}</p>
          </td>
          <td v-if="showReadiness">
            <template v-if="r.syarat_aktif_lengkap !== undefined">
              <UBadge
                :label="r.syarat_aktif_lengkap ? 'LENGKAP' : 'BELUM LENGKAP'"
                :icon="r.syarat_aktif_lengkap ? 'i-lucide-shield-check' : 'i-lucide-shield-alert'"
                :color="r.syarat_aktif_lengkap ? 'success' : 'warning'"
                variant="subtle"
                size="sm"
              />
              <ul v-if="r.syarat_aktif_kurang?.length" class="mt-1 space-y-0.5 text-xs text-muted">
                <li v-for="k in r.syarat_aktif_kurang" :key="k" class="flex gap-1">
                  <span aria-hidden="true">–</span>{{ k.split(' (')[0] }}
                </li>
              </ul>
            </template>
          </td>
          <td class="text-right">
            <UButton
              :to="`/library/${r.cbg_code}`"
              :label="r.can_act ? 'Tinjau' : 'Buka'"
              :icon="r.can_act ? 'i-lucide-clipboard-check' : undefined"
              :variant="r.can_act ? 'solid' : 'outline'"
              :color="r.can_act ? 'primary' : 'neutral'"
              size="sm"
            />
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
