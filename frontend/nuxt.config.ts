// TransCPG-X frontend — Nuxt 4 SPA (tanpa SSR).
// API Go berjalan terpisah; saat dev, /api diteruskan ke backend lokal.
export default defineNuxtConfig({
  compatibilityDate: '2026-10-01',
  ssr: false,
  modules: ['@nuxt/ui'],
  css: ['~/assets/css/main.css'],
  devtools: { enabled: true },

  app: {
    // Perpindahan halaman halus; dimatikan otomatis bila pengguna memilih
    // "kurangi gerakan" (lihat main.css).
    pageTransition: { name: 'page', mode: 'out-in' },
    layoutTransition: { name: 'layout', mode: 'out-in' },
    head: {
      title: 'TransCPG-X',
      htmlAttrs: { lang: 'id' },
    },
  },

  runtimeConfig: {
    public: {
      // Produksi: NUXT_PUBLIC_API_BASE=https://<backend>.vercel.app/api
      apiBase: '/api',
    },
  },

  nitro: {
    devProxy: {
      '/api': {
        target: process.env.NUXT_DEV_API_PROXY ?? 'http://127.0.0.1:8080/api',
        changeOrigin: true,
      },
    },
  },
})
