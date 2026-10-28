/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      screens: {
        xs: '576px',
        sm: '768px',
        md: '992px',
        lg: '1200px',
        xl: '1920px',
      },
      colors: {
        'sm-primary': 'var(--sm-color-primary)',
        'sm-success': 'var(--sm-color-success)',
        'sm-warning': 'var(--sm-color-warning)',
        'sm-danger': 'var(--sm-color-danger)',
        'sm-bg-page': 'var(--sm-bg-page)',
        'sm-bg-card': 'var(--sm-bg-card)',
        'sm-bg-header': 'var(--sm-bg-header)',
        'sm-text': 'var(--sm-text-primary)',
        'sm-text-regular': 'var(--sm-text-regular)',
        'sm-text-secondary': 'var(--sm-text-secondary)',
        'sm-border': 'var(--sm-border-base)',
        'sm-border-light': 'var(--sm-border-light)',
      },
      spacing: {
        'sm-xs': 'var(--sm-spacing-xs)',
        'sm-sm': 'var(--sm-spacing-sm)',
        'sm-md': 'var(--sm-spacing-md)',
        'sm-lg': 'var(--sm-spacing-lg)',
        'sm-xl': 'var(--sm-spacing-xl)',
        'sm-xxl': 'var(--sm-spacing-xxl)',
      },
      borderRadius: {
        'sm-sm': 'var(--sm-radius-sm)',
        'sm-md': 'var(--sm-radius-md)',
        'sm-lg': 'var(--sm-radius-lg)',
      },
      fontSize: {
        'sm-xs': 'var(--sm-font-size-xs)',
        'sm-sm': 'var(--sm-font-size-sm)',
        'sm-base': 'var(--sm-font-size-base)',
        'sm-lg': 'var(--sm-font-size-lg)',
        'sm-xl': 'var(--sm-font-size-xl)',
      },
    },
  },
  plugins: [],
  corePlugins: {
    preflight: false,  // 避免与 Element Plus 冲突
  },
}
