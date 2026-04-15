import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import { visualizer } from "rollup-plugin-visualizer";
import viteCompression from 'vite-plugin-compression';

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    visualizer({ open: false }), // 开发时不自动打开，避免干扰
    // 1. Gzip 压缩 (兼容性好)
    viteCompression({
      verbose: true,
      disable: false,
      threshold: 10240,
      algorithm: 'gzip',
      ext: '.gz',
    }),
    // 2. Brotli 压缩 (比 Gzip 压缩率更高，现代浏览器支持好)
    viteCompression({
      verbose: true,
      disable: false,
      threshold: 10240,
      algorithm: 'brotliCompress',
      ext: '.br',
    })
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  server: {
    host: true, // 允许局域网访问
    proxy: {
      '/api': {
        // 后端接口地址，开发时代理到本地后端服务器
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: true,
      }
    }
  },
  // 优化：阻止特定 Worker 被打包
  optimizeDeps: {
    exclude: [
      'monaco-editor/esm/vs/language/typescript/ts.worker',
      'monaco-editor/esm/vs/language/css/css.worker',
      'monaco-editor/esm/vs/language/html/html.worker',
      'monaco-editor/esm/vs/language/json/json.worker',
    ]
  },
  build: {
    target: 'es2015', // 确保构建产物兼容性
    minify: 'esbuild', // 默认是 esbuild，速度快。生产环境若追求极致体积可换 'terser'
    chunkSizeWarningLimit: 2000,
    rollupOptions: {
      output: {
        // 优化分包策略
        manualChunks(id) {
          if (id.includes('node_modules')) {
            // 将 Monaco Editor 单独拆分，因为它非常巨大
            if (id.includes('monaco-editor')) {
              // 保持 Worker 单独分块以防加载问题
              if (id.includes('.worker')) {
                return 'monaco-workers';
              }
              return 'monaco-editor';
            }
            // 将 Vue 全家桶拆分
            if (id.includes('vue') || id.includes('vue-router') || id.includes('pinia')) {
              return 'vendor-vue';
            }
            // 工具库拆分
            if (id.includes('lodash') || id.includes('axios')) {
              return 'vendor-utils';
            }
            // Phosphor Icons 单独分块
            if (id.includes('phosphor-icons')) {
              return 'phosphor-icons';
            }
            // 其他第三方库
            return 'vendor-libs';
          }
        },
      },
    },
  },
})