import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import MonacoEditorPlugin from 'vite-plugin-monaco-editor'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    MonacoEditorPlugin.default({
      languageWorkers: ['editorWorkerService', 'typescript', 'json', 'html', 'css'],
      customWorkers: [
        {
          label: 'python',
          entry: 'monaco-editor/esm/vs/language/python/python.worker?worker'
        },
        {
          label: 'go',
          entry: 'monaco-editor/esm/vs/language/go/go.worker?worker'
        },
        {
          label: 'java',
          entry: 'monaco-editor/esm/vs/language/java/java.worker?worker'
        }
      ]
    })
  ],
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
        changeOrigin: true
      }
    }
  }
})