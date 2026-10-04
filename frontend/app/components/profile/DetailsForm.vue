<script setup lang="ts">
// Ubah nama tampilan dan nomor telepon (dipakai untuk masuk dengan kode OTP).
import type { SessionUser } from '~/types/api'

const props = defineProps<{ user: SessionUser }>()
const { updateProfile } = useAuth()
const toast = useToast()

const form = reactive({ full_name: props.user.full_name, phone: props.user.phone ?? '' })
const dirty = computed(() => form.full_name.trim() !== props.user.full_name || form.phone.trim() !== (props.user.phone ?? ''))
const saving = ref(false)
const error = ref('')

// Nomor tersimpan dalam format +62…; tampilkan apa adanya setelah simpan.
watch(() => props.user, (u) => {
  form.full_name = u.full_name
  form.phone = u.phone ?? ''
})

async function save() {
  error.value = ''
  if (!form.full_name.trim()) {
    error.value = 'Nama lengkap wajib diisi.'
    return
  }
  saving.value = true
  try {
    await updateProfile({ full_name: form.full_name.trim(), phone: form.phone.trim() })
    toast.add({ title: 'Profil tersimpan', color: 'success', icon: 'i-lucide-check' })
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
    <header>
      <h3 class="font-semibold text-highlighted">Data diri</h3>
      <p class="text-sm text-muted">Nama tampil di riwayat pengesahan dan jejak audit.</p>
    </header>
    <UFormField label="Nama lengkap" required>
      <UInput v-model="form.full_name" autocomplete="name" class="w-full" maxlength="120" />
    </UFormField>
    <UFormField label="Nomor telepon" hint="Opsional" help="Untuk masuk dengan kode OTP. Kosongkan untuk menghapus.">
      <UInput v-model="form.phone" type="tel" autocomplete="tel" placeholder="0812 3456 7890" icon="i-lucide-phone" class="w-full" />
    </UFormField>
    <UFormField label="Email" help="Email adalah identitas masuk — hanya Admin yang dapat mengubahnya.">
      <UInput :model-value="user.email" disabled icon="i-lucide-lock" class="w-full" />
    </UFormField>
    <UAlert v-if="error" :title="error" color="error" variant="subtle" icon="i-lucide-circle-alert" />
    <div class="flex justify-end gap-2">
      <UButton v-if="dirty" label="Batalkan" color="neutral" variant="ghost" @click="form.full_name = user.full_name; form.phone = user.phone ?? ''" />
      <UButton type="submit" label="Simpan perubahan" :loading="saving" :disabled="!dirty" />
    </div>
  </form>
</template>
