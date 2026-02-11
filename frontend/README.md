# Vue 3 + Vite

This template should help get you started developing with Vue 3 in Vite. The template uses Vue 3 `<script setup>` SFCs, check out the [script setup docs](https://v3.vuejs.org/api/sfc-script-setup.html#sfc-script-setup) to learn more.

Learn more about IDE Support for Vue in the [Vue Docs Scaling up Guide](https://vuejs.org/guide/scaling-up/tooling.html#ide-support).

## 🚀 性能优化 (Build Optimization)

本项目已针对生产环境构建进行了深度优化，集成了 Gzip 和 Brotli 双重压缩，并实施了精细化的分包策略。

### 优化效果对比 (Based on `vite build` output)

| 资源文件 | 原始体积 (Raw) | Gzip 压缩后 | Brotli 压缩后 | 优化幅度 |
| :--- | :--- | :--- | :--- | :--- |
| **Monaco Editor Core** (`monaco-editor-*.js`) | ~4.32 MB | ~1.10 MB | **~856 KB** | **↓ 80%** |
| **TS Worker** (`ts.worker-*.js`) | ~7.03 MB | ~1.50 MB | **~1.10 MB** | **↓ 84%** |
| **Vue Vendor** (`vendor-vue-*.js`) | ~106 KB | ~41 KB | **~37 KB** | **↓ 65%** |
| **General Libs** (`vendor-libs-*.js`) | ~387 KB | ~116 KB | **~95 KB** | **↓ 75%** |

### 优化策略详解

1.  **双重压缩 (Dual Compression)**:
    - 同时生成 `.gz` (Gzip) and `.br` (Brotli) 文件。
    - 现代浏览器优先加载 `.br` 文件，不支持时降级加载 `.gz` 文件。
    - 显著减少首屏资源传输体积。

2.  **精细化分包 (Manual Chunks)**:
    - **Monaco Editor**: 单独拆分。这是一个巨大的库，将其独立后，业务代码更新时用户无需重新下载编辑器核心。
    - **Vendor Vue**: 将 Vue, Vue Router, Pinia 等核心框架库打包在一起，利用浏览器长效缓存。
    - **Vendor Utils**: 将 Lodash, Axios 等工具库打包在一起。
    - **Phosphor Icons**: 图标库单独打包。

3.  **构建配置**:
    - `target: 'es2015'`: 确保构建产物兼容性。
    - `minify: 'esbuild'`: 极速混淆压缩。
    - `chunkSizeWarningLimit: 2000`: 调整警告阈值，适应 Monaco Editor 的大体积特性。
