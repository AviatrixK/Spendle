/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        bg: '#0D0D0F',
        surface: {
          DEFAULT: '#171719',
          elevated: '#202023',
        },
        card: 'rgba(255, 255, 255, 0.04)',
        'card-border': 'rgba(255, 255, 255, 0.08)',
        main: '#EDEAE3',
        muted: '#85827B',
        accent: {
          DEFAULT: '#F5B83D',
          hover: '#E5A72B',
          glow: 'rgba(245, 184, 61, 0.15)',
        },
        semantic: {
          profit: '#73D6A5', // Mint
          loss: '#F07878',   // Coral
        },
        // Categorical ramp used strictly for segmented bar, legend dots, and category icons
        ramp: {
          gold: '#F5B83D',
          sand: '#D9C7A0',
          sage: '#8FB8A0',
          steel: '#8EA3C7',
          rose: '#C98B9A',
          mauve: '#9A8FC2',
        },
      },
      fontFamily: {
        serif: ['"Instrument Serif"', 'Georgia', 'serif'],
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['"JetBrains Mono"', 'monospace'],
      },
      borderRadius: {
        '2xl': '1rem',
        '3xl': '1.75rem', // 28px
      },
      boxShadow: {
        glass: '0 24px 60px -20px rgba(0, 0, 0, 0.6)',
        'glass-elevated': '0 30px 70px -15px rgba(0, 0, 0, 0.7)',
        gold: '0 0 25px -5px rgba(245, 184, 61, 0.3)',
      },
    },
  },
  plugins: [],
};
