/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        page: 'var(--page)',
        surface: 'var(--surface)',
        'surface-2': 'var(--surface-2)',
        ink: 'var(--ink)',
        'ink-soft': 'var(--ink-soft)',
        border: 'var(--border)',
        'border-strong': 'var(--border-strong)',
        primary: 'var(--primary-500)',
        'primary-700': 'var(--primary-700)',
        'primary-300': 'var(--primary-300)',
        'primary-100': 'var(--primary-100)',
        accent: 'var(--accent-500)',
        'accent-600': 'var(--accent-600)',
        'accent-100': 'var(--accent-100)',
        overlay: 'var(--overlay)',
      },
      fontFamily: {
        sans: ['var(--font-primary)'],
        display: ['var(--font-secondary)'],
      },
      borderRadius: {
        card: 'var(--radius-card)',
        tile: 'var(--radius-tile)',
        pill: 'var(--radius-pill)',
      },
      boxShadow: {
        card: 'var(--shadow-card)',
      },
    },
  },
  plugins: [],
}
