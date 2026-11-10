/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        bg: '#0E0E11',
        card: 'rgba(255, 255, 255, 0.04)',
        'card-border': 'rgba(255, 255, 255, 0.08)',
        main: '#EDEAE3',
        muted: '#8A8780',
        accent: {
          DEFAULT: '#F5B83D',
          hover: '#E5A72B',
          glow: 'rgba(245, 184, 61, 0.15)',
        },
        semantic: {
          profit: '#6FD6A0',
          loss: '#FF6B5E',
        },
      },
      fontFamily: {
        serif: ['"Instrument Serif"', 'Georgia', 'serif'],
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['"JetBrains Mono"', 'monospace'],
      },
      borderRadius: {
        '2xl': '1rem',
      },
    },
  },
  plugins: [],
};

