// Halaman Pengaturan hanya untuk Admin RS, System Admin, Super Admin.
// Server tetap menolak request non-admin; ini hanya mencegah halaman kosong.
export default defineNuxtRouteMiddleware(() => {
  const { isAdmin } = useAuth()
  if (!isAdmin.value) {
    return navigateTo('/dashboard')
  }
})
