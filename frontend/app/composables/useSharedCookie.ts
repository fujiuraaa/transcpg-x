import type { CookieOptions, CookieRef } from '#app'

const cache = new WeakMap<object, Map<string, CookieRef<unknown>>>()

/**
 * Seperti useCookie, tetapi setiap pemanggilan dengan nama yang sama
 * mengembalikan ref yang SAMA. useCookie biasa membuat ref terpisah per
 * pemanggilan sehingga perubahan di satu composable tidak terlihat di yang lain
 * (mis. token baru setelah login tidak terbaca oleh klien API).
 */
export function useSharedCookie<T>(name: string, opts?: CookieOptions<T> & { readonly?: false }): CookieRef<T> {
  const nuxtApp = useNuxtApp()
  let refs = cache.get(nuxtApp)
  if (!refs) {
    refs = new Map()
    cache.set(nuxtApp, refs)
  }
  if (!refs.has(name)) {
    refs.set(name, useCookie<T>(name, opts) as CookieRef<unknown>)
  }
  return refs.get(name) as CookieRef<T>
}
