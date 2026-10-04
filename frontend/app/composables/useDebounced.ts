/** Salinan nilai `source` yang baru berubah setelah `ms` tanpa perubahan. */
export function useDebounced<T>(source: Ref<T>, ms = 300): Ref<T> {
  const out = ref(source.value) as Ref<T>
  let timer: ReturnType<typeof setTimeout> | undefined
  watch(source, (v) => {
    clearTimeout(timer)
    timer = setTimeout(() => { out.value = v }, ms)
  })
  onScopeDispose(() => clearTimeout(timer))
  return out
}
