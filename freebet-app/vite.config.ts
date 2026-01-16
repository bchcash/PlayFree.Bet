import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
  root: ".",
  build: {
    outDir: "dist",
    emptyOutDir: true,
    chunkSizeWarningLimit: 1000, // Увеличить порог предупреждения до 1MB
    rollupOptions: {
      output: {
        manualChunks(id) {
          // Vendor чанки
          if (id.includes('node_modules')) {
            if (id.includes('react') || id.includes('react-dom')) {
              return 'vendor-react';
            }
            if (id.includes('@radix-ui')) {
              return 'vendor-radix';
            }
            if (id.includes('recharts')) {
              return 'vendor-charts';
            }
            if (id.includes('framer-motion') || id.includes('date-fns') || id.includes('clsx')) {
              return 'vendor-utils';
            }
            if (id.includes('@tanstack/react-query') || id.includes('wouter')) {
              return 'vendor-router';
            }
            // Остальные node_modules идут в общий vendor чанк
            return 'vendor';
          }

          // Страницы - каждая страница в отдельный чанк
          if (id.includes('/pages/')) {
            const pageName = id.split('/pages/')[1].split('.')[0];
            return `page-${pageName.toLowerCase()}`;
          }

          // Компоненты - группируем по типу
          if (id.includes('/components/')) {
            if (id.includes('/ui/')) {
              return 'components-ui';
            }
            if (id.includes('/BetCard') || id.includes('/BettingCard') || id.includes('/BettingModal')) {
              return 'components-betting';
            }
            return 'components-main';
          }

          // Хуки и утилиты
          if (id.includes('/hooks/')) {
            return 'hooks';
          }
          if (id.includes('/lib/')) {
            return 'lib';
          }
        },
      },
    },
  },
  server: {
    host: "0.0.0.0",
    port: 5000,
    allowedHosts: true,
    proxy: {
      "/api": {
        target: "http://localhost:3000",
        changeOrigin: true,
      },
    },
  },
});