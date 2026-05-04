import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import basicSsl from '@vitejs/plugin-basic-ssl'
import { VitePWA } from 'vite-plugin-pwa'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    basicSsl(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.ico', 'apple-touch-icon.png', 'mask-icon.svg'],
      manifest: {
        name: 'Maplas - Akıllı Gezi Rehberi',
        short_name: 'Maplas',
        description: 'Yapay zeka destekli akıllı gezi ve rota planlayıcı.',
        theme_color: '#10b981',
        icons: [
          {
            src: 'Maplas.png',
            sizes: '192x192',
            type: 'image/png'
          },
          {
            src: 'Maplas.png',
            sizes: '512x512',
            type: 'image/png'
          },
          {
            src: 'Maplas.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'any maskable'
          }
        ]
      }
    })
  ],
  server: {
    host: true,
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
        secure: false
      },
      '/uploads': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
        secure: false
      }
    },
    watch: {
      usePolling: true
    }
  }
})
