// Tema TransCPG-X: gradasi merah, tanpa biru.
// - primary  : merah (identitas, tombol utama bergradasi)
// - secondary: oranye (tahap review awal)
// - info     : stone (keterangan netral — menggantikan biru bawaan)
// - error    : rose (dibedakan dari merah utama; selalu disertai ikon + teks)
export default defineAppConfig({
  ui: {
    colors: {
      primary: 'red',
      secondary: 'orange',
      info: 'stone',
      error: 'rose',
      neutral: 'stone',
    },
    button: {
      compoundVariants: [
        {
          color: 'primary',
          variant: 'solid',
          class: 'bg-linear-to-br from-(--ui-color-primary-500) to-(--ui-color-primary-700) hover:from-(--ui-color-primary-600) hover:to-(--ui-color-primary-800) active:from-(--ui-color-primary-700) active:to-(--ui-color-primary-800) disabled:from-(--ui-color-primary-400) disabled:to-(--ui-color-primary-500) aria-disabled:from-(--ui-color-primary-400) aria-disabled:to-(--ui-color-primary-500) shadow-sm shadow-(--ui-color-primary-900)/20',
        },
      ],
    },
  },
})
