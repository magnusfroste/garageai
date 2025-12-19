/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#007AFF',
        secondary: '#FF3B30',
        tertiary: '#34C759',
        'text-primary': '#1D1D1F',
        'text-secondary': '#86868B',
        'bg-primary': '#FFFFFF',
        'bg-secondary': '#F5F5F7',
        'bg-tertiary': '#E5E5E7',
        border: '#D1D1D6',
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
