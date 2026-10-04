// Setiap navigasi: arahkan ke /login bila belum masuk, dan muat data diri
// sekali saat aplikasi dibuka.
export default defineNuxtRouteMiddleware(async (to) => {
  const { isLoggedIn, user, refreshMe } = useAuth()
  const isPublic = to.meta.public === true

  if (!isLoggedIn.value) {
    return isPublic ? undefined : navigateTo('/login')
  }
  if (isPublic) {
    return navigateTo('/dashboard')
  }
  if (!user.value) {
    try {
      await refreshMe()
    }
    catch (err) {
      // 401 sudah ditangani useApi (token dihapus → /login). Galat lain
      // (server mati, jaringan) jangan dialihkan ke /login: token masih ada
      // sehingga /login akan memantul balik ke /dashboard tanpa henti.
      const status = (err as { statusCode?: number, status?: number })?.statusCode
        ?? (err as { status?: number })?.status
      if (status === 401 || !useSessionToken().value) {
        useSessionToken().value = null
        return navigateTo('/login')
      }
      return abortNavigation(createError({
        statusCode: 503,
        statusMessage: 'Server tidak dapat dihubungi',
        message: apiErrorMessage(err),
        fatal: true,
      }))
    }
  }
})
