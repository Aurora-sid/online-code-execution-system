/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#3B82F6',
        secondary: '#27272a', // Zinc 800
        cta: '#2563EB',
        background: '#000000', // Pure Black
        surface: '#18181b',    // Zinc 900
        text: '#e4e4e7',       // Zinc 200
        accent: {
          purple: '#8B5CF6',
          pink: '#EC4899',
          cyan: '#06B6D4',
          emerald: '#10B981',
        }
      },
      fontFamily: {
        sans: ['"IBM Plex Sans"', 'sans-serif'],
        mono: ['"JetBrains Mono"', 'monospace'],
      },
    },
  },
  plugins: [],
}
