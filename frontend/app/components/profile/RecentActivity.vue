<script setup lang="ts">
// 20 aktivitas terakhir akun ini, termasuk masuk aplikasi — membantu
// pengguna mengenali akses yang bukan dirinya.
const { myActivity } = useAuth()
const { data, status, refresh } = await useAsyncData('profile:activity', myActivity)
defineExpose({ refresh })
</script>

<template>
  <section class="rounded-xl border border-default bg-default">
    <header class="flex items-center justify-between gap-2 border-b border-default px-5 py-4">
      <div>
        <h3 class="font-semibold text-highlighted">Aktivitas terakhir</h3>
        <p class="text-sm text-muted">20 kejadian terbaru dari akun Anda.</p>
      </div>
      <UButton icon="i-lucide-refresh-cw" size="sm" color="neutral" variant="ghost" aria-label="Muat ulang aktivitas" :loading="status === 'pending'" @click="refresh()" />
    </header>
    <div v-if="status === 'pending' && !data" class="space-y-3 p-5">
      <USkeleton v-for="i in 4" :key="i" class="h-10 w-full" />
    </div>
    <p v-else-if="!data?.length" class="p-5 text-sm text-muted">Belum ada aktivitas tercatat.</p>
    <ol v-else class="divide-y divide-default">
      <li v-for="e in data" :key="e.uid" class="flex items-start gap-3 px-5 py-3">
        <span
          class="mt-0.5 grid size-7 shrink-0 place-items-center rounded-full"
          :class="isAttention(e.action) ? 'bg-warning/10 text-warning' : 'bg-elevated text-muted'"
        >
          <UIcon :name="isAttention(e.action) ? 'i-lucide-triangle-alert' : AUDIT_CATEGORY[e.category].icon" class="size-3.5" />
        </span>
        <div class="min-w-0 flex-1">
          <p class="text-sm" :class="isAttention(e.action) && 'text-warning'">{{ e.summary }}</p>
          <p class="truncate text-xs text-muted">
            <template v-if="e.target_code || (e.target_label && e.target_type !== 'USER')">
              {{ [e.target_code, e.target_label].filter(Boolean).join(' · ') }} ·
            </template>
            <time :datetime="e.created_at" :title="formatDate(e.created_at, true)">{{ relativeTime(e.created_at) }}</time>
            <template v-if="e.ip"> · IP {{ e.ip }}</template>
          </p>
        </div>
      </li>
    </ol>
    <p class="border-t border-default px-5 py-3 text-xs text-muted">
      Ada aktivitas yang bukan Anda? Segera ganti kata sandi dan beri tahu Admin RS.
    </p>
  </section>
</template>
