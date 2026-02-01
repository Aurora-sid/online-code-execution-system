/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        background: '#f8fafc',      // 浅灰白色背景
        surface: '#ffffff',          // 白色表面
        surfaceHighlight: '#f1f5f9', // 浅灰高亮
        primary: '#6366f1',          // 靛蓝色主色
        secondary: '#06b6d4',        // 青色次色
        accent: '#ec4899',           // 粉色强调
        success: '#10b981',          // 绿色成功
        border: 'rgba(0, 0, 0, 0.1)', // 深色边框
        text: '#1e293b',             // 深色文字
        textMuted: '#64748b',        // 灰色文字
      },
      fontFamily: {
        sans: ['Outfit', 'sans-serif'],
        mono: ['"JetBrains Mono"', 'monospace'],
      },
      boxShadow: {
        'glow-primary': '0 0 20px -5px rgba(99, 102, 241, 0.4)',
        'glow-secondary': '0 0 20px -5px rgba(6, 182, 212, 0.4)',
      },
    },
  },
  plugins: [],
}
