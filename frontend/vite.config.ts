import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig(({ mode }) => {
  const isProd = mode === 'production'

  return {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, 'src')
      }
    },
    server: {
      host: '0.0.0.0',
      port: 3000,
      open: false,
      cors: true,
      proxy: {
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true,
          ws: false,
          configure: (proxy) => {
            proxy.on('error', (err) => {
              console.log('vite proxy error:', err)
            })
          }
        }
      }
    },
    build: {
      target: 'es2018',
      outDir: 'dist',
      sourcemap: false,
      minify: isProd ? 'esbuild' : false,
      chunkSizeWarningLimit: 1500,
      rollupOptions: {
        output: {
          manualChunks: {
            'element-plus': ['element-plus'],
            'vue-vendor': ['vue', 'vue-router', 'pinia', 'axios']
          }
        }
      }
    },
    optimizeDeps: {
      include: [
        'vue',
        'vue-router',
        'pinia',
        'axios',
        'element-plus',
        'codemirror'
      ]
    }
  }
})
