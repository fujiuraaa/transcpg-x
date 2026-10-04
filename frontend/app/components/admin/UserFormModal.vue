<script setup lang="ts">
// Dialog tambah / ubah pengguna oleh Admin.
// Tambah: nama, email, telepon, peran, (RS), kata sandi awal.
// Ubah: nama, telepon, peran, status aktif, atur ulang kata sandi (opsional).
import type { SessionUser } from '~/types/api'
import type { Role } from '~/types/domain'

const props = defineProps<{
  user: SessionUser | null // null = tambah
  roleOptions: { value: Role, label: string }[]
}>()
const open = defineModel<boolean>('open', { default: false })
const emit = defineEmits<{ saved: [] }>()

const { createUser, updateUser } = useAdminApi()
const { user: me } = useAuth()
const { hospitals } = useHospitalProfile()
const toast = useToast()

const isEdit = computed(() => !!props.user)
const isSelf = computed(() => !!props.user && props.user.id === me.value?.id)
const canPickHospital = computed(() => me.value?.role === 'SUPER_ADMIN' || me.value?.role === 'SYSTEM_ADMIN')

const form = reactive({
  full_name: '',
  email: '',
  phone: '',
  role: undefined as Role | undefined,
  hospital_id: undefined as number | undefined,
  is_active: true,
  password: '',
  resetPassword: false,
})
const showPassword = ref(false)
const copied = ref(false)
const error = ref('')
const busy = ref(false)

watch(open, (v) => {
  if (!v) return
  error.value = ''
  copied.value = false
  showPassword.value = false
  const u = props.user
  Object.assign(form, {
    full_name: u?.full_name ?? '',
    email: u?.email ?? '',
    phone: u?.phone ?? '',
    role: u?.role,
    hospital_id: u?.hospital_id ?? me.value?.hospital_id ?? undefined,
    is_active: u?.is_active ?? true,
    password: u ? '' : generatePassword(),
    resetPassword: false,
  })
  if (!u) showPassword.value = true
})

function regenerate() {
  form.password = generatePassword()
  showPassword.value = true
  copied.value = false
}
async function copyPassword() {
  try {
    await navigator.clipboard.writeText(form.password)
    copied.value = true
  }
  catch {
    toast.add({ title: 'Tidak bisa menyalin — salin manual', color: 'warning' })
  }
}

const needsPassword = computed(() => !isEdit.value || form.resetPassword)

async function save() {
  error.value = ''
  if (!form.full_name.trim()) return (error.value = 'Nama lengkap wajib diisi.')
  if (!isEdit.value && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email.trim())) return (error.value = 'Format email tidak valid.')
  if (!form.role) return (error.value = 'Pilih peran.')
  if (needsPassword.value && form.password.length < 8) return (error.value = 'Kata sandi minimal 8 karakter.')

  busy.value = true
  try {
    if (!props.user) {
      await createUser({
        full_name: form.full_name.trim(),
        email: form.email.trim(),
        phone: form.phone.trim() || null,
        password: form.password,
        role: form.role,
        hospital_id: canPickHospital.value ? form.hospital_id : undefined,
      })
      toast.add({ title: 'Pengguna ditambahkan', description: `${form.full_name} · sampaikan kata sandi awal secara aman.`, color: 'success' })
    }
    else {
      const body: Parameters<typeof updateUser>[1] = { full_name: form.full_name.trim(), phone: form.phone.trim() }
      if (!isSelf.value) {
        if (form.role !== props.user.role) body.role = form.role
        if (form.is_active !== props.user.is_active) body.is_active = form.is_active
      }
      if (form.resetPassword) body.password = form.password
      await updateUser(props.user.id, body)
      toast.add({ title: 'Pengguna diperbarui', description: form.full_name, color: 'success' })
    }
    open.value = false
    emit('saved')
  }
  catch (err) {
    error.value = apiErrorMessage(err)
  }
  finally {
    busy.value = false
  }
}
</script>

<template>
  <UModal
    v-model:open="open"
    :title="isEdit ? 'Ubah pengguna' : 'Tambah pengguna'"
    :description="isEdit ? user?.email : 'Akun langsung aktif. Sampaikan kata sandi awal kepada pengguna secara aman.'"
    :ui="{ content: 'max-w-lg' }"
  >
    <template #body>
      <form id="user-form" class="space-y-4" novalidate @submit.prevent="save">
        <UFormField label="Nama lengkap" required>
          <UInput v-model="form.full_name" class="w-full" placeholder="Beserta gelar, mis. dr. Ani, Sp.PD" autofocus />
        </UFormField>

        <div class="grid gap-4 sm:grid-cols-2">
          <UFormField label="Email" :required="!isEdit" :description="isEdit ? 'Email tidak dapat diubah.' : undefined">
            <UInput v-model="form.email" type="email" class="w-full" :disabled="isEdit" autocomplete="off" />
          </UFormField>
          <UFormField label="Nomor telepon" description="Untuk masuk dengan kode OTP.">
            <UInput v-model="form.phone" type="tel" class="w-full" placeholder="0812 3456 7890" />
          </UFormField>
        </div>

        <div class="grid gap-4 sm:grid-cols-2">
          <UFormField label="Peran" required :description="isSelf ? 'Anda tidak dapat mengubah peran sendiri.' : undefined">
            <USelect v-model="form.role" :items="roleOptions" placeholder="Pilih peran" class="w-full" :disabled="isSelf" />
          </UFormField>
          <UFormField v-if="!isEdit && canPickHospital" label="Rumah sakit">
            <USelect v-model="form.hospital_id" :items="hospitals.map(h => ({ label: h.name, value: h.id }))" class="w-full" />
          </UFormField>
          <UFormField v-if="isEdit" label="Status akun" :description="isSelf ? 'Anda tidak dapat menonaktifkan akun sendiri.' : 'Akun nonaktif tidak bisa masuk.'">
            <USwitch v-model="form.is_active" :label="form.is_active ? 'Aktif' : 'Nonaktif'" :disabled="isSelf" />
          </UFormField>
        </div>

        <!-- Kata sandi -->
        <div class="space-y-2 rounded-lg border border-default p-3">
          <USwitch v-if="isEdit" v-model="form.resetPassword" label="Atur ulang kata sandi" @update:model-value="(v) => v && regenerate()" />
          <template v-if="needsPassword">
            <UFormField :label="isEdit ? 'Kata sandi baru' : 'Kata sandi awal'" required description="Minimal 8 karakter. Minta pengguna menggantinya setelah masuk.">
              <UFieldGroup class="w-full">
                <UInput v-model="form.password" :type="showPassword ? 'text' : 'password'" class="w-full font-mono" autocomplete="new-password" />
                <UButton :icon="showPassword ? 'i-lucide-eye-off' : 'i-lucide-eye'" color="neutral" variant="outline" :aria-label="showPassword ? 'Sembunyikan' : 'Tampilkan'" @click="showPassword = !showPassword" />
                <UButton icon="i-lucide-refresh-cw" color="neutral" variant="outline" aria-label="Buat kata sandi acak" title="Buat acak" @click="regenerate" />
                <UButton :icon="copied ? 'i-lucide-check' : 'i-lucide-copy'" color="neutral" variant="outline" :aria-label="copied ? 'Tersalin' : 'Salin kata sandi'" title="Salin" @click="copyPassword" />
              </UFieldGroup>
            </UFormField>
          </template>
        </div>

        <UAlert v-if="error" color="error" variant="subtle" icon="i-lucide-circle-alert" :title="error" />
      </form>
    </template>
    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <UButton label="Batal" color="neutral" variant="ghost" @click="open = false" />
        <UButton type="submit" form="user-form" :label="isEdit ? 'Simpan perubahan' : 'Tambah pengguna'" :loading="busy" />
      </div>
    </template>
  </UModal>
</template>
