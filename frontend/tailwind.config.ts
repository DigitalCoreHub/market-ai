import type { Config } from 'tailwindcss';

const config: Config = {
  darkMode: 'class', // Enable class-based dark mode
  content: [
    './app/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {},
  },
  plugins: [],
};

export default config;
