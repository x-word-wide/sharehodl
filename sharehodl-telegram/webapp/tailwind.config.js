/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // ShareHODL brand colors - Dark Blue theme
        primary: {
          DEFAULT: '#1E40AF',
          light: '#3B82F6',
          dark: '#1E3A8A'
        },
        accent: {
          blue: '#1976D2',
          green: '#00C853',
          red: '#FF1744',
          orange: '#F59E0B',
          purple: '#8B5CF6'
        },
        dark: {
          bg: '#0D1117',
          card: '#161B22',
          surface: '#21262D',
          border: '#30363D'
        },
        // Telegram theme colors
        tg: {
          bg: 'var(--tg-theme-bg-color, #0D1117)',
          text: 'var(--tg-theme-text-color, #ffffff)',
          hint: 'var(--tg-theme-hint-color, #8b949e)',
          link: 'var(--tg-theme-link-color, #6366F1)',
          button: 'var(--tg-theme-button-color, #6366F1)',
          'button-text': 'var(--tg-theme-button-text-color, #ffffff)',
          secondary: 'var(--tg-theme-secondary-bg-color, #161B22)'
        }
      },
      fontFamily: {
        sans: ['Inter', 'SF Pro Display', '-apple-system', 'BlinkMacSystemFont', 'sans-serif']
      }
    },
  },
  plugins: [],
}
