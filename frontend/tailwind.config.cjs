const defaultTheme = require('tailwindcss/defaultTheme')
const forms = require('@tailwindcss/forms')

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'pass': '#10B981',
        'warning': '#F59E0B',
        'fail': '#EF4444',
      },
      fontFamily: {
        sans: ['Inter var', ...defaultTheme.fontFamily.sans],
      },
    },
  },
  plugins: [forms],
} 