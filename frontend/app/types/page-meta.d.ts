declare module '#app' {
  interface PageMeta {
    /** Halaman yang bisa dibuka tanpa login (hanya /login). */
    public?: boolean
    /** Judul di bilah atas. */
    title?: string
  }
}

export {}
