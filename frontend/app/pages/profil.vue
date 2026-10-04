<script setup lang="ts">
// Profil Saya — semua peran: data diri, kata sandi, dan aktivitas terakhir.
definePageMeta({ title: 'Profil Saya' })

const { user } = useAuth()
const { hospitals, load } = useHospitalProfile()
onMounted(() => {
  if (!hospitals.value.length) load()
})
</script>

<template>
  <div class="mx-auto max-w-[1200px] space-y-6">
    <header class="bg-brand-soft rounded-xl border border-primary/15 px-5 py-4">
      <h2 class="text-2xl font-semibold text-highlighted">Profil Saya</h2>
      <p class="text-sm text-muted">Kelola data diri dan keamanan akun Anda.</p>
    </header>

    <div v-if="user" class="grid gap-6 lg:grid-cols-[minmax(0,20rem)_minmax(0,1fr)]">
      <div class="space-y-6 lg:sticky lg:top-20 lg:self-start">
        <ProfileIdentityCard :user="user" />
      </div>
      <div class="space-y-6">
        <ProfileDetailsForm :user="user" />
        <ProfilePasswordForm />
        <ProfileSessionsCard />
        <ProfileRecentActivity />
      </div>
    </div>
  </div>
</template>
