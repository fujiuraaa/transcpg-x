<script setup lang="ts">
// Ganti kata sandi sendiri. Kata sandi lama wajib; syarat ditampilkan langsung.
const { changePassword } = useAuth()
const toast = useToast()

const form = reactive({ current: '', next: '', confirm: '' })
const show = ref(false)
const saving = ref(false)
const error = ref('')

const rules = computed(() => [
  { label: 'Minimal 8 karakter', ok: form.next.length >= 8 },
  { label: 'Memuat huruf dan angka', ok: /[a-z]/i.test(form.next) && /\d/.test(form.next) },
  { label: 'Berbeda dari kata sandi lama', ok: !!form.next && form.next !== form.current },
  { label: 'Konfirmasi sama', ok: !!form.next && form.next === form.confirm },
])
// Huruf+angka disarankan, tidak diwajibkan server; yang wajib: panjang, beda, konfirmasi.
const ready = computed(() => !!form.current && rules.value[0]!.ok && rules.value[2]!.ok && rules.value[3]!.ok)

async function save() {
  if (!ready.value) return
  error.value = ''
  saving.value = true
  try {
    await changePassword(form.current, form.next)
    toast.add({ title: 'Kata sandi diganti', description: 'Sesi di perangkat lain otomatis dikeluarkan.', color: 'success', icon: 'i-lucide-shield-check' })
    form.current = form.next = form.confirm = ''
  }
  catch (err) {
    error.value = apiErrorMessage(err)
  }
  finally {
    saving.value = false
  }
}
</script>

<template>
  <form class="space-y-4 rounded-xl border border-default bg-default p-5" @submit.prevent="save">
    <header class="flex items-start justify-between gap-3">
      <div>
        <h3 class="font-semibold text-highlighted">Kata sandi</h3>
        <p class="text-sm text-muted">Untuk masuk dengan email.</p>
      </div>
      <UButton
        :icon="show ? 'i-lucide-eye-off' : 'i-lucide-eye'"
        :label="show ? 'Sembunyikan' : 'Tampilkan'"
        size="xs"
        color="neutral"
        variant="ghost"
        @click="show = !show"
      />
    </header>
    <!-- Nama pengguna tersembunyi membantu pengelola kata sandi peramban. -->
    <input type="text" autocomplete="username" class="sr-only" tabindex="-1" aria-hidden="true">
    <UFormField label="Kata sandi saat ini" required>
      <UInput v-model="form.current" :type="show ? 'text' : 'password'" autocomplete="current-password" class="w-full" />
    </UFormField>
    <div class="grid gap-4 sm:grid-cols-2">
      <UFormField label="Kata sandi baru" required>
        <UInput v-model="form.next" :type="show ? 'text' : 'password'" autocomplete="new-password" maxlength="72" class="w-full" />
      </UFormField>
      <UFormField label="Ulangi kata sandi baru" required>
        <UInput v-model="form.confirm" :type="show ? 'text' : 'password'" autocomplete="new-password" maxlength="72" class="w-full" />
      </UFormField>
    </div>
    <ul class="grid gap-1 text-xs sm:grid-cols-2" aria-label="Syarat kata sandi">
      <li v-for="r in rules" :key="r.label" class="flex items-center gap-1.5 transition-colors" :class="r.ok ? 'text-success' : 'text-muted'">
        <UIcon :name="r.ok ? 'i-lucide-circle-check' : 'i-lucide-circle'" class="size-3.5" />
        {{ r.label }}
      </li>
    </ul>
    <UAlert v-if="error" :title="error" color="error" variant="subtle" icon="i-lucide-circle-alert" />
    <div class="flex justify-end">
      <UButton type="submit" label="Ganti kata sandi" icon="i-lucide-key-round" :loading="saving" :disabled="!ready" />
    </div>
  </form>
</template>
